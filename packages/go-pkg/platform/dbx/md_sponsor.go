package dbx

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Sponsors: the paid cards an admin manages. Changes say why they can't be made with a ConfigError,
// like the site's lists.

const (
	maxSponsorName = 100
	maxSponsorText = 160
	maxSponsorLink = 500
	// MaxSponsorImage is the largest picture, in bytes. It is kept in the database.
	MaxSponsorImage = 200 * 1024
)

// sponsorImageTypes are the pictures a card can show. Anything else, SVG included, is refused:
// the picture is served from the site's own address.
var sponsorImageTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/gif":  true,
	"image/webp": true,
}

// SponsorInput is what an admin sends for a sponsor.
type SponsorInput struct {
	Name     string
	Text     string
	Link     string
	StartsOn string
	EndsOn   string
	Enabled  bool
	// a new picture, nil keeps the one the sponsor has
	Image []byte
	// removes the picture. Ignored when Image is set.
	RemoveImage bool
}

// SponsorRow is a sponsor as the admin list shows it, with how often it was seen and clicked.
type SponsorRow struct {
	ID            int64  `gorm:"column:id"`
	Name          string `gorm:"column:name"`
	Text          string `gorm:"column:text"`
	Link          string `gorm:"column:link"`
	StartsOn      string `gorm:"column:starts_on"`
	EndsOn        string `gorm:"column:ends_on"`
	Enabled       bool   `gorm:"column:enabled"`
	HasImage      bool   `gorm:"column:has_image"`
	UpdatedAt     int64  `gorm:"column:updated_at"`
	Impressions   int64  `gorm:"column:impressions"`
	Clicks        int64  `gorm:"column:clicks"`
	Impressions30 int64  `gorm:"column:impressions30"`
	Clicks30      int64  `gorm:"column:clicks30"`
}

// SponsorCount is what a card can be counted for.
type SponsorCount int

const (
	SponsorImpression SponsorCount = iota + 1
	SponsorClick
)

// cleanSponsorLine trims one line of text and refuses a long or control character one. An empty
// line is allowed only if required is false.
func cleanSponsorLine(value, what string, max int, required bool) (string, error) {
	value = strings.Join(strings.Fields(value), " ")
	if value == "" {
		if required {
			return "", configErr(ConfigInvalid, "Enter the %s", what)
		}
		return "", nil
	}
	if utf8.RuneCountInString(value) > max {
		return "", configErr(ConfigInvalid, "The %s is too long (at most %d characters)", what, max)
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return "", configErr(ConfigInvalid, "The %s has characters that can't be used", what)
		}
	}
	return value, nil
}

// cleanSponsorDate accepts an empty date (no limit) or one written 2006-01-02.
func cleanSponsorDate(value, what string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if _, err := time.Parse(models.SponsorDateLayout, value); err != nil {
		return "", configErr(ConfigInvalid, "The %s is not a date", what)
	}
	return value, nil
}

// cleanSponsor checks and tidies a sponsor's fields. The picture is checked by cleanSponsorImage.
func cleanSponsor(in SponsorInput) (SponsorInput, error) {
	var err error
	if in.Name, err = cleanSponsorLine(in.Name, "sponsor's name", maxSponsorName, true); err != nil {
		return in, err
	}
	if in.Text, err = cleanSponsorLine(in.Text, "text", maxSponsorText, false); err != nil {
		return in, err
	}
	if in.Link, err = cleanSponsorLine(in.Link, "link", maxSponsorLink, true); err != nil {
		return in, err
	}
	parsed, err := url.Parse(in.Link)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return in, configErr(ConfigInvalid, "The link has to be a web address starting with https://")
	}
	if in.StartsOn, err = cleanSponsorDate(in.StartsOn, "first day"); err != nil {
		return in, err
	}
	if in.EndsOn, err = cleanSponsorDate(in.EndsOn, "last day"); err != nil {
		return in, err
	}
	if in.StartsOn != "" && in.EndsOn != "" && in.EndsOn < in.StartsOn {
		return in, configErr(ConfigInvalid, "The last day is before the first day")
	}
	return in, nil
}

// cleanSponsorImage returns the type of a picture, or says why it can't be used.
func cleanSponsorImage(image []byte) (string, error) {
	if len(image) > MaxSponsorImage {
		return "", configErr(ConfigInvalid, "The picture is too big (at most %d KB)", MaxSponsorImage/1024)
	}
	kind := http.DetectContentType(image)
	if !sponsorImageTypes[kind] {
		return "", configErr(ConfigInvalid, "The picture has to be a PNG, JPEG, GIF or WebP file")
	}
	return kind, nil
}

// apply puts the checked fields on a sponsor, and its picture if the input changes it.
func (in SponsorInput) apply(row *models.Sponsor) error {
	row.Name, row.Text, row.Link = in.Name, in.Text, in.Link
	row.StartsOn, row.EndsOn, row.Enabled = in.StartsOn, in.EndsOn, in.Enabled
	switch {
	case in.Image != nil:
		kind, err := cleanSponsorImage(in.Image)
		if err != nil {
			return err
		}
		row.Image, row.ImageType = in.Image, kind
	case in.RemoveImage:
		row.Image, row.ImageType = nil, ""
	}
	return nil
}

// ListSponsorRows is every sponsor for the admin, with counts for all time and for the 30 days up
// to and including today (2006-01-02).
func (d *DatabaseClient) ListSponsorRows(today string) ([]SponsorRow, error) {
	cutoff := today
	if day, err := time.Parse(models.SponsorDateLayout, today); err == nil {
		cutoff = day.AddDate(0, 0, -29).Format(models.SponsorDateLayout)
	}
	rows := make([]SponsorRow, 0)
	err := d.DB.Raw(`
		SELECT s.id, s.name, s.text, s.link, s.starts_on, s.ends_on, s.enabled, s.updated_at,
			(COALESCE(s.image_type, '') <> '') AS has_image,
			COALESCE(SUM(c.impressions), 0) AS impressions,
			COALESCE(SUM(c.clicks), 0) AS clicks,
			COALESCE(SUM(c.impressions) FILTER (WHERE c.date >= ?), 0) AS impressions30,
			COALESCE(SUM(c.clicks) FILTER (WHERE c.date >= ?), 0) AS clicks30
		FROM sponsors s LEFT JOIN sponsor_daily_counts c ON c.sponsor_id = s.id
		WHERE s.deleted_at IS NULL
		GROUP BY s.id
		ORDER BY s.id`, cutoff, cutoff).Scan(&rows).Error
	return rows, err
}

// ListSponsors is every sponsor without its picture. The caller picks the ones that are active.
func (d *DatabaseClient) ListSponsors() ([]models.Sponsor, error) {
	var rows []models.Sponsor
	err := d.DB.Omit("image").Order("id").Find(&rows).Error
	return rows, err
}

// GetSponsor is one sponsor without its picture.
func (d *DatabaseClient) GetSponsor(id int64) (models.Sponsor, error) {
	var row models.Sponsor
	err := d.DB.Omit("image").Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, configErr(ConfigNotFound, "That sponsor does not exist")
	}
	return row, err
}

// SponsorImage is a sponsor's picture and its type. The picture is empty if there is none.
func (d *DatabaseClient) SponsorImage(id int64) ([]byte, string, error) {
	var row models.Sponsor
	err := d.DB.Select("id", "image", "image_type").Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", configErr(ConfigNotFound, "That sponsor does not exist")
	}
	return row.Image, row.ImageType, err
}

func (d *DatabaseClient) CreateSponsor(in SponsorInput) error {
	in, err := cleanSponsor(in)
	if err != nil {
		return err
	}
	var row models.Sponsor
	if err := in.apply(&row); err != nil {
		return err
	}
	return d.DB.Create(&row).Error
}

func (d *DatabaseClient) UpdateSponsor(id int64, in SponsorInput) error {
	in, err := cleanSponsor(in)
	if err != nil {
		return err
	}
	return d.DB.Transaction(func(tx *gorm.DB) error {
		row, err := loadOne[models.Sponsor](tx, id, "sponsor")
		if err != nil {
			return err
		}
		if err := in.apply(&row); err != nil {
			return err
		}
		return tx.Save(&row).Error
	})
}

func (d *DatabaseClient) DeleteSponsor(id int64) error {
	result := d.DB.Delete(&models.Sponsor{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return configErr(ConfigNotFound, "That sponsor does not exist")
	}
	return nil
}

// CountSponsor adds one impression or click to today's count, in the site's time zone.
func (d *DatabaseClient) CountSponsor(id int64, kind SponsorCount, today string) error {
	count := models.SponsorDailyCount{Date: today, SponsorID: id}
	column := "impressions"
	switch kind {
	case SponsorImpression:
		count.Impressions = 1
	case SponsorClick:
		count.Clicks = 1
		column = "clicks"
	default:
		return errors.New("unknown sponsor count")
	}
	return d.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "date"}, {Name: "sponsor_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			column: gorm.Expr("sponsor_daily_counts." + column + " + 1"),
		}),
	}).Create(&count).Error
}
