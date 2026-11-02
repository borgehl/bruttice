package dbx

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/borgehl/bruttice/packages/go-pkg/configs"
	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

// What an admin can change about the lists the site is made of: regions, tags, age groups and
// venues. Each change is one transaction, and a change that can't be made says why with a
// ConfigError, whose Message is meant for the admin.

type ConfigErrorKind int

const (
	// the request is wrong: a bad name, a missing replacement
	ConfigInvalid ConfigErrorKind = iota + 1
	ConfigNotFound
	// the request is fine but the data in the way: a name that exists, events that use it
	ConfigConflict
)

type ConfigError struct {
	Kind    ConfigErrorKind
	Message string
}

func (e *ConfigError) Error() string { return e.Message }

func configErr(kind ConfigErrorKind, format string, args ...any) error {
	return &ConfigError{Kind: kind, Message: fmt.Sprintf(format, args...)}
}

const (
	maxConfigName = 60
	maxVenueName  = 100
	maxVenueLink  = 500
)

// cleanName trims a name and squeezes its whitespace, and refuses an empty, long or control
// character name.
func cleanName(name, what string, max int) (string, error) {
	name = strings.Join(strings.Fields(name), " ")
	switch {
	case name == "":
		return "", configErr(ConfigInvalid, "Enter a name for the %s", what)
	case utf8.RuneCountInString(name) > max:
		return "", configErr(ConfigInvalid, "The %s name is too long (at most %d characters)", what, max)
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", configErr(ConfigInvalid, "The %s name has characters that can't be used", what)
		}
	}
	return name, nil
}

// nameTaken is the id of another row of table whose name is the same ignoring case, 0 if there is none.
func nameTaken(tx *gorm.DB, table, name string, exceptID int64) (int64, error) {
	var id int64
	err := tx.Table(table).Select("COALESCE(MAX(id), 0)").
		Where("lower(name) = lower(?) AND id <> ? AND deleted_at IS NULL", name, exceptID).Scan(&id).Error
	return id, err
}

func loadOne[T any](tx *gorm.DB, id int64, what string) (T, error) {
	var row T
	err := tx.Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, configErr(ConfigNotFound, "That %s does not exist", what)
	}
	return row, err
}

// ---- regions

type RegionRow struct {
	ID   int64
	Name string
	// events not deleted, and venues, that use the region
	Events    int64
	Venues    int64
	RetiredAt int64
	// the name of the region a retired region was merged into, empty if there is none
	MergedInto string
}

func (d *DatabaseClient) ListRegionRows() ([]RegionRow, error) {
	rows := make([]RegionRow, 0)
	err := d.DB.Raw(`
		SELECT r.id, r.name, r.retired_at, COALESCE(m.name, '') AS merged_into,
			(SELECT COUNT(*) FROM events e WHERE e.region_id = r.id AND e.deleted_at IS NULL) AS events,
			(SELECT COUNT(*) FROM venues v WHERE v.default_region = r.name AND v.deleted_at IS NULL) AS venues
		FROM regions r LEFT JOIN regions m ON m.id = r.merged_into_id
		WHERE r.deleted_at IS NULL
		ORDER BY (r.retired_at > 0), r.name`).Scan(&rows).Error
	return rows, err
}

// CreateRegion adds a region. A retired region of that name is brought back instead.
func (d *DatabaseClient) CreateRegion(name string) error {
	name, err := cleanName(name, "region", maxConfigName)
	if err != nil {
		return err
	}
	return d.DB.Transaction(func(tx *gorm.DB) error {
		var existing models.Region
		found := tx.Where("name = ?", name).First(&existing).Error
		if found == nil {
			if !existing.Retired() {
				return configErr(ConfigConflict, "There is a region called %q already", name)
			}
			return tx.Model(&existing).Updates(map[string]any{"retired_at": 0, "merged_into_id": nil}).Error
		}
		if !errors.Is(found, gorm.ErrRecordNotFound) {
			return found
		}
		if id, err := nameTaken(tx, "regions", name, 0); err != nil {
			return err
		} else if id != 0 {
			return configErr(ConfigConflict, "There is a region with that name already, spelled differently")
		}
		return tx.Create(&models.Region{Name: name}).Error
	})
}

// RenameRegion gives a region a new name. Scrapers and submissions may still use the old one, so
// this is a retirement of the old name into the new, and the old name keeps working.
func (d *DatabaseClient) RenameRegion(id int64, name string) error {
	name, err := cleanName(name, "region", maxConfigName)
	if err != nil {
		return err
	}
	return d.DB.Transaction(func(tx *gorm.DB) error {
		old, err := activeRegion(tx, id)
		if err != nil {
			return err
		}
		if old.Name == name {
			return nil
		}
		other, err := nameTaken(tx, "regions", name, old.ID)
		if err != nil {
			return err
		}
		if other == 0 {
			return d.retireInto(tx, old, models.Region{Name: name})
		}
		var existing models.Region
		if err := tx.Where("id = ?", other).First(&existing).Error; err != nil {
			return err
		}
		switch {
		case existing.Name != name:
			// the same letters in other case, or the region renamed into its own spelling
			return configErr(ConfigConflict, "There is a region with that name already, spelled differently")
		case !existing.Retired():
			return configErr(ConfigConflict, "There is a region called %q already. Retire this one into it instead", name)
		}
		return d.retireInto(tx, old, existing)
	})
}

func activeRegion(tx *gorm.DB, id int64) (models.Region, error) {
	region, err := loadOne[models.Region](tx, id, "region")
	if err != nil {
		return region, err
	}
	if region.Retired() {
		return region, configErr(ConfigConflict, "%q is retired already", region.Name)
	}
	return region, nil
}

type RetireResult struct {
	EventsMoved   int64
	EventsDeleted int64
}

// RetireRegion takes a region out of use. Its events either move to the region given, and so do
// the venues that default to it and the names scrapers still send, or they are deleted. At least
// one other region has to stay.
func (d *DatabaseClient) RetireRegion(id int64, moveToID *int64, deleteEvents bool) (RetireResult, error) {
	var result RetireResult
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		old, err := activeRegion(tx, id)
		if err != nil {
			return err
		}
		var others int64
		if err := tx.Model(&models.Region{}).Where("retired_at = 0 AND id <> ?", id).Count(&others).Error; err != nil {
			return err
		}
		if others == 0 {
			return configErr(ConfigConflict, "At least one region has to stay")
		}

		switch {
		case moveToID != nil && deleteEvents, moveToID == nil && !deleteEvents:
			return configErr(ConfigInvalid, "Choose a region to move the events to, or delete them")
		case moveToID != nil:
			if *moveToID == id {
				return configErr(ConfigInvalid, "Choose a different region")
			}
			target, err := activeRegion(tx, *moveToID)
			if err != nil {
				return err
			}
			result.EventsMoved, err = d.retireIntoCounting(tx, old, target)
			return err
		}

		// deleted events: the venues would be left with a default nobody can use
		var venues []string
		if err := tx.Model(&models.Venue{}).Where("default_region = ?", old.Name).Order("name").Pluck("name", &venues).Error; err != nil {
			return err
		}
		if len(venues) > 0 {
			return configErr(ConfigConflict, "These venues default to %q, change their region first or move the events instead: %s", old.Name, strings.Join(venues, ", "))
		}
		deleted := tx.Where("region_id = ?", id).Delete(&models.Event{})
		if deleted.Error != nil {
			return deleted.Error
		}
		result.EventsDeleted = deleted.RowsAffected
		return tx.Model(&old).Updates(map[string]any{"retired_at": time.Now().Unix(), "merged_into_id": nil}).Error
	})
	return result, err
}

// retireInto retires old into target, creating target if it has no id yet.
func (d *DatabaseClient) retireInto(tx *gorm.DB, old, target models.Region) error {
	if target.ID == 0 {
		if err := tx.Create(&target).Error; err != nil {
			return err
		}
	} else if target.Retired() {
		if err := tx.Model(&target).Updates(map[string]any{"retired_at": 0, "merged_into_id": nil}).Error; err != nil {
			return err
		}
	}
	_, err := d.retireIntoCounting(tx, old, target)
	return err
}

func (d *DatabaseClient) retireIntoCounting(tx *gorm.DB, old, target models.Region) (int64, error) {
	// unscoped: events an admin removed earlier must not be left pointing at a retired region
	moved := tx.Unscoped().Model(&models.Event{}).Where("region_id = ? AND deleted_at IS NULL", old.ID).Update("region_id", target.ID)
	if moved.Error != nil {
		return 0, moved.Error
	}
	if err := tx.Unscoped().Model(&models.Event{}).Where("region_id = ?", old.ID).Update("region_id", target.ID).Error; err != nil {
		return 0, err
	}
	if err := tx.Model(&models.Venue{}).Where("default_region = ?", old.Name).Update("default_region", target.Name).Error; err != nil {
		return 0, err
	}
	// names that were already merged into the old region now lead to the new one
	if err := tx.Model(&models.Region{}).Where("merged_into_id = ?", old.ID).Update("merged_into_id", target.ID).Error; err != nil {
		return 0, err
	}
	err := tx.Model(&models.Region{}).Where("id = ?", old.ID).
		Updates(map[string]any{"retired_at": time.Now().Unix(), "merged_into_id": target.ID}).Error
	return moved.RowsAffected, err
}

// ---- tags

type TagRow struct {
	ID     int64
	Name   string
	Events int64
}

func (d *DatabaseClient) ListTagRows() ([]TagRow, error) {
	rows := make([]TagRow, 0)
	err := d.DB.Raw(`
		SELECT t.id, t.name,
			(SELECT COUNT(*) FROM event_tags et JOIN events e ON e.id = et.event_id
				WHERE et.tag_id = t.id AND e.deleted_at IS NULL) AS events
		FROM tags t WHERE t.deleted_at IS NULL ORDER BY t.name`).Scan(&rows).Error
	return rows, err
}

func (d *DatabaseClient) CreateTag(name string) error {
	name, err := cleanName(name, "tag", maxConfigName)
	if err != nil {
		return err
	}
	return d.DB.Transaction(func(tx *gorm.DB) error {
		if id, err := nameTaken(tx, "tags", name, 0); err != nil {
			return err
		} else if id != 0 {
			return configErr(ConfigConflict, "There is a tag with that name already")
		}
		return tx.Create(&models.Tag{Name: name}).Error
	})
}

// RenameTag renames a tag on all of its events. A scraper that still sends the old name has that
// tag skipped, as with any tag the site doesn't have.
func (d *DatabaseClient) RenameTag(id int64, name string) error {
	name, err := cleanName(name, "tag", maxConfigName)
	if err != nil {
		return err
	}
	return d.DB.Transaction(func(tx *gorm.DB) error {
		tag, err := loadOne[models.Tag](tx, id, "tag")
		if err != nil {
			return err
		}
		if tag.Name == configs.FreeTag {
			return configErr(ConfigConflict, "The %q tag is added to free events automatically and can't be renamed", tag.Name)
		}
		if other, err := nameTaken(tx, "tags", name, id); err != nil {
			return err
		} else if other != 0 {
			return configErr(ConfigConflict, "There is a tag with that name already")
		}
		return tx.Model(&tag).Update("name", name).Error
	})
}

// DeleteTag removes a tag from every event. No event is deleted.
func (d *DatabaseClient) DeleteTag(id int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		tag, err := loadOne[models.Tag](tx, id, "tag")
		if err != nil {
			return err
		}
		if tag.Name == configs.FreeTag {
			return configErr(ConfigConflict, "The %q tag is added to free events automatically and can't be deleted", tag.Name)
		}
		if err := tx.Exec("DELETE FROM event_tags WHERE tag_id = ?", id).Error; err != nil {
			return err
		}
		// unscoped, so the name can be used again
		return tx.Unscoped().Delete(&models.Tag{}, id).Error
	})
}

// ---- age groups

type AgeRow struct {
	ID     int64
	Name   string
	Value  uint8
	Events int64
}

func (d *DatabaseClient) ListAgeRows() ([]AgeRow, error) {
	rows := make([]AgeRow, 0)
	err := d.DB.Raw(`
		SELECT a.id, a.name, a.value,
			(SELECT COUNT(*) FROM events e WHERE e.age_id = a.id AND e.deleted_at IS NULL) AS events
		FROM ages a WHERE a.deleted_at IS NULL ORDER BY a.value, a.name`).Scan(&rows).Error
	return rows, err
}

func cleanAgeValue(value int) (uint8, error) {
	if value < 0 || value > 255 {
		return 0, configErr(ConfigInvalid, "The age has to be between 0 and 255")
	}
	return uint8(value), nil
}

func (d *DatabaseClient) CreateAge(name string, value int) error {
	name, err := cleanName(name, "age group", maxConfigName)
	if err != nil {
		return err
	}
	age, err := cleanAgeValue(value)
	if err != nil {
		return err
	}
	return d.DB.Transaction(func(tx *gorm.DB) error {
		if id, err := nameTaken(tx, "ages", name, 0); err != nil {
			return err
		} else if id != 0 {
			return configErr(ConfigConflict, "There is an age group with that name already")
		}
		return tx.Create(&models.Age{Name: name, Value: age}).Error
	})
}

// UpdateAge changes the name and the age of an age group. Scrapers and submissions pick an age group
// by name, so one that still sends the old name will have its events refused.
func (d *DatabaseClient) UpdateAge(id int64, name string, value int) error {
	name, err := cleanName(name, "age group", maxConfigName)
	if err != nil {
		return err
	}
	age, err := cleanAgeValue(value)
	if err != nil {
		return err
	}
	return d.DB.Transaction(func(tx *gorm.DB) error {
		row, err := loadOne[models.Age](tx, id, "age group")
		if err != nil {
			return err
		}
		if other, err := nameTaken(tx, "ages", name, id); err != nil {
			return err
		} else if other != 0 {
			return configErr(ConfigConflict, "There is an age group with that name already")
		}
		return tx.Model(&row).Updates(map[string]any{"name": name, "value": age}).Error
	})
}

// DeleteAge removes an age group. Events that have it move to the replacement, which is needed if
// there are any.
func (d *DatabaseClient) DeleteAge(id int64, replacementID *int64) (moved int64, err error) {
	err = d.DB.Transaction(func(tx *gorm.DB) error {
		row, err := loadOne[models.Age](tx, id, "age group")
		if err != nil {
			return err
		}
		var used int64
		if err := tx.Unscoped().Model(&models.Event{}).Where("age_id = ?", id).Count(&used).Error; err != nil {
			return err
		}
		if used > 0 {
			if replacementID == nil || *replacementID == id {
				return configErr(ConfigInvalid, "Choose another age group for the events of %q", row.Name)
			}
			if _, err := loadOne[models.Age](tx, *replacementID, "age group"); err != nil {
				return err
			}
			// unscoped: events that were removed earlier still refer to it
			res := tx.Unscoped().Model(&models.Event{}).Where("age_id = ?", id).Update("age_id", *replacementID)
			if res.Error != nil {
				return res.Error
			}
			moved = used
		}
		return tx.Unscoped().Delete(&models.Age{}, id).Error
	})
	return moved, err
}

// ---- venues

type VenueRow struct {
	ID            int64
	Name          string
	Link          string
	DefaultRegion string
	IsScraped     bool
	Events        int64
}

func (d *DatabaseClient) ListVenueRows() ([]VenueRow, error) {
	rows := make([]VenueRow, 0)
	err := d.DB.Raw(`
		SELECT v.id, v.name, v.link, v.default_region, v.is_scraped,
			(SELECT COUNT(*) FROM events e WHERE e.venue_id = v.id AND e.deleted_at IS NULL) AS events
		FROM venues v WHERE v.deleted_at IS NULL ORDER BY v.slug`).Scan(&rows).Error
	return rows, err
}

func cleanVenue(name, link, region string, tx *gorm.DB) (string, string, string, error) {
	name, err := cleanName(name, "venue", maxVenueName)
	if err != nil {
		return "", "", "", err
	}
	link = strings.TrimSpace(link)
	if link != "" {
		u, err := url.Parse(link)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || len(link) > maxVenueLink {
			return "", "", "", configErr(ConfigInvalid, "The link has to be a web address starting with http:// or https://")
		}
	}
	var count int64
	if err := tx.Model(&models.Region{}).Where("name = ? AND retired_at = 0", region).Count(&count).Error; err != nil {
		return "", "", "", err
	}
	if count == 0 {
		return "", "", "", configErr(ConfigInvalid, "Choose one of the regions in use as the venue's region")
	}
	return name, link, region, nil
}

// Venues that a scraper reports are rewritten by it on every run, and the reserved unlisted
// venue is how events without a venue are saved, so neither is changed here.
func editableVenue(venue models.Venue) error {
	switch {
	case venue.Name == configs.UnlistedName:
		return configErr(ConfigConflict, "This venue is how events without a venue are saved, it can't be changed")
	case venue.IsScraped:
		return configErr(ConfigConflict, "%q is kept up to date by its scraper, change it there", venue.Name)
	}
	return nil
}

func (d *DatabaseClient) CreateVenue(name, link, region string) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		name, link, region, err := cleanVenue(name, link, region, tx)
		if err != nil {
			return err
		}
		if id, err := nameTaken(tx, "venues", name, 0); err != nil {
			return err
		} else if id != 0 {
			return configErr(ConfigConflict, "There is a venue with that name already")
		}
		venue := models.Venue{Name: name, Link: link, DefaultRegion: region}
		venue.InternalUpdate()
		return tx.Create(&venue).Error
	})
}

func (d *DatabaseClient) UpdateVenue(id int64, name, link, region string) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		venue, err := loadOne[models.Venue](tx, id, "venue")
		if err != nil {
			return err
		}
		if err := editableVenue(venue); err != nil {
			return err
		}
		name, link, region, err := cleanVenue(name, link, region, tx)
		if err != nil {
			return err
		}
		if other, err := nameTaken(tx, "venues", name, id); err != nil {
			return err
		} else if other != 0 {
			return configErr(ConfigConflict, "There is a venue with that name already")
		}
		venue.Name, venue.Link, venue.DefaultRegion = name, link, region
		venue.InternalUpdate()
		return tx.Save(&venue).Error
	})
}

// DeleteVenue removes a venue that no event has ever used, removed events included.
func (d *DatabaseClient) DeleteVenue(id int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		venue, err := loadOne[models.Venue](tx, id, "venue")
		if err != nil {
			return err
		}
		if err := editableVenue(venue); err != nil {
			return err
		}
		var used int64
		if err := tx.Unscoped().Model(&models.Event{}).Where("venue_id = ?", id).Count(&used).Error; err != nil {
			return err
		}
		if used > 0 {
			return configErr(ConfigConflict, "%q has events, so it can't be deleted", venue.Name)
		}
		return tx.Unscoped().Delete(&models.Venue{}, id).Error
	})
}
