package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

func getPreloadWhere(include bool) string {
	if include {
		return "in ?"
	}
	return "not in ?"
}

func (d *DatabaseClient) getTagFilter(filter models.CalendarFilterBackend, tx *gorm.DB) (*gorm.DB, error) {
	tagIds, err := d.GetTagIds(filter.Tags, d.DB)
	if err != nil {
		return nil, err
	}
	inStr := "in"
	if !filter.TagsInclude {
		inStr = "not " + inStr
	}
	return tx.Where(
		"events.id "+inStr+" (select distinct event_id from event_tags where tag_id in ?)",
		tagIds,
	), nil
}

func (d *DatabaseClient) getAgeFilter(filter models.CalendarFilterBackend, tx *gorm.DB) (*gorm.DB, error) {
	age, err := d.GetAge(models.Age{Value: filter.Age}, d.DB)
	if err != nil {
		return nil, err
	}
	comp := ">="
	if filter.AgesLessThan {
		comp = "<="
	}
	return tx.Joins(
		"inner join ages on ages.id = events.age_id and ages.value "+comp+" ?",
		age.Value,
	), nil
}

func (d *DatabaseClient) getVenueFilter(filter models.CalendarFilterBackend, tx *gorm.DB) (*gorm.DB, error) {
	venueIds, err := d.GetVenueIds(filter.Venues, d.DB)
	if err != nil {
		return nil, err
	}
	return tx.Joins(
		"inner join venues on venues.id = events.venue_id and venues.id "+getPreloadWhere(filter.VenuesInclude),
		venueIds,
	), nil
}

// getRegionFilter keeps the events of the named regions. A name that is not a region, or is
// retired, matches nothing. Retiring a region moves its events to the one it was merged into.
func getRegionFilter(names []string, tx *gorm.DB) *gorm.DB {
	return tx.Where(
		"events.region_id in (select id from regions where name in ? and retired_at = 0)",
		names,
	)
}

func (d *DatabaseClient) GetTxWithFilter(filter models.CalendarFilterBackend, tx *gorm.DB) (*gorm.DB, error) {
	var err error
	// two separate conditions, so an event has to pass both. Merging the lists would turn two
	// regions with nothing in common into an empty list, which means every region.
	if len(filter.Regions) > 0 {
		tx = getRegionFilter(filter.Regions, tx)
	}
	if len(filter.RequestRegions) > 0 {
		tx = getRegionFilter(filter.RequestRegions, tx)
	}
	if len(filter.Tags) > 0 {
		tx, err = d.getTagFilter(filter, tx)
		if err != nil {
			return nil, err
		}
	}
	if len(filter.Venues) > 0 {
		tx, err = d.getVenueFilter(filter, tx)
		if err != nil {
			return nil, err
		}
	}
	if filter.Age != 0 {
		tx, err = d.getAgeFilter(filter, tx)
		if err != nil {
			return nil, err
		}
	}
	return tx.Preload("Region"), nil
}
