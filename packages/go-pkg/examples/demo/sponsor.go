package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/psqlx"

	"gorm.io/gorm"
)

// One example sponsor, so the card shows in the demo without anyone opening the admin page. Like the
// visitors it is written straight to the database: sponsors are managed by an admin and are not
// part of the ingest API.
//
// It is found by its name, so running again replaces it (and moves its counts forward) and leaves
// every other sponsor alone.
const (
	demoSponsorName = "Example Records & Tapes"
	demoSponsorText = "Local vinyl, cassettes and zines. Fake business, real card."
	// example.com is reserved for examples, so a click goes nowhere that exists
	demoSponsorLink = "https://example.com/sponsor"
	sponsorDays     = 30
)

// demoSponsorImage is a small picture made up on the spot, as tall as a flyer thumbnail: a
// gradient with a stripe, so it is clearly a placeholder.
func demoSponsorImage() []byte {
	const width, height = 64, 88
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := color.RGBA{R: uint8(40 + y*2), G: uint8(90 + x), B: 160, A: 255}
			if y >= 38 && y < 50 {
				c = color.RGBA{R: 245, G: 245, B: 235, A: 255}
			}
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		// encoding a plain RGBA image to memory does not fail
		panic(err)
	}
	return buf.Bytes()
}

// sponsorCounts makes up how often the card was seen and clicked on each of the last sponsorDays
// days, today first. It is a pure function of `now`.
func sponsorCounts(sponsorID int64, now time.Time, loc *time.Location) []models.SponsorDailyCount {
	today := now.In(loc)
	ret := make([]models.SponsorDailyCount, 0, sponsorDays)
	for daysAgo := 0; daysAgo < sponsorDays; daysAgo++ {
		impressions := int64(40 + (daysAgo*13)%35) // 40 to 74
		ret = append(ret, models.SponsorDailyCount{
			Date:        today.AddDate(0, 0, -daysAgo).Format(models.SponsorDateLayout),
			SponsorID:   sponsorID,
			Impressions: impressions,
			Clicks:      impressions / 12,
		})
	}
	return ret
}

// seedSponsor replaces the example sponsor in the database named by the POSTGRES_* environment
// variables, the same ones the backend reads.
func seedSponsor(loc *time.Location, now time.Time) error {
	db, err := psqlx.NewDBFromEnvVars()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var old []models.Sponsor
		if err := tx.Unscoped().Select("id").Where("name = ?", demoSponsorName).Find(&old).Error; err != nil {
			return err
		}
		for _, s := range old {
			if err := tx.Where("sponsor_id = ?", s.ID).Delete(&models.SponsorDailyCount{}).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Delete(&models.Sponsor{}, s.ID).Error; err != nil {
				return err
			}
		}

		sponsor := models.Sponsor{
			Name:      demoSponsorName,
			Text:      demoSponsorText,
			Link:      demoSponsorLink,
			Image:     demoSponsorImage(),
			ImageType: "image/png",
			Enabled:   true,
		}
		if err := tx.Create(&sponsor).Error; err != nil {
			return err
		}
		return tx.Create(sponsorCounts(sponsor.ID, now, loc)).Error
	})
}
