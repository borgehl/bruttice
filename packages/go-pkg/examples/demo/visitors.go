package main

import (
	"fmt"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/psqlx"

	"gorm.io/gorm"
)

// Example visitors for the admin Visitors page. They are written straight to the
// database, since the only way in is the consent cookie of a real browser.
//
// The visitor ids all start with visitorIDPrefix, so running again replaces the
// earlier rows (and moves the dates forward) and leaves real visitors alone.
const (
	visitorIDPrefix = "00000000-0000-4000-8000-"
	visitorDays     = 60

	// The busiest day of the last 7 days, and of the last 30. The Visitors page scales its chart to
	// the busiest day of the range, so the two ranges are shown on different axes.
	busiestDayOfWeek  = 15
	busiestDayOfMonth = 55

	// one visitor for each on the busiest day, who is then also the pool for every other day
	visitorCount = busiestDayOfMonth
)

// weekVisitors is how many visitors came on each of the last 7 days, today first.
var weekVisitors = [7]int{12, busiestDayOfWeek, 9, 6, 11, 14, 8}

// dailyVisitors is how many different visitors came `daysAgo` days ago, 0 being today. The last 7
// days stay at or under busiestDayOfWeek, days 7 to 29 are higher with the busiest day of the last
// 30 among them, and older days are quieter.
func dailyVisitors(daysAgo int) int {
	switch {
	case daysAgo < len(weekVisitors):
		return weekVisitors[daysAgo]
	case daysAgo == 12:
		return busiestDayOfMonth
	case daysAgo < 30:
		return 22 + (daysAgo*11)%30 // 22 to 51
	default:
		return 8 + (daysAgo*7)%25 // 8 to 32
	}
}

// place is the share of visitors from [from, to) out of 100. A place without a
// country is a visitor whose location lookup failed.
type place struct {
	from, to             int
	country, state, city string
}

var places = []place{
	{0, 30, "US", "US-TX", "Austin"},
	{30, 50, "US", "US-TX", "Houston"},
	{50, 65, "US", "US-IL", "Chicago"},
	{65, 75, "US", "US-CA", "Los Angeles"},
	{75, 83, "US", "US-NY", "New York"},
	{83, 90, "CA", "CA-ON", "Toronto"},
	{90, 95, "GB", "GB-ENG", "London"},
	{95, 100, "", "", ""},
}

type visitorSeed struct {
	location models.VisitorLocation
	// when the visitor loaded a page
	visits []time.Time
}

// visitors makes up the visitors of the last visitorDays days. It is a pure
// function of `now`. Each day has the number of visitors dailyVisitors gives: visitor n comes on a
// day when n is within it, so the same few come most often and the busiest days bring the rest.
func visitors(now time.Time, loc *time.Location) []visitorSeed {
	today := now.In(loc)
	ret := make([]visitorSeed, 0, visitorCount)

	for n := 1; n <= visitorCount; n++ {
		id := fmt.Sprintf("%s%012d", visitorIDPrefix, n)
		seed := visitorSeed{location: models.VisitorLocation{VisitorID: id}}

		pick := (n * 37) % 100
		for _, p := range places {
			if pick >= p.from && pick < p.to {
				seed.location.Country, seed.location.StateCode, seed.location.City = p.country, p.state, p.city
			}
		}

		for daysAgo := 0; daysAgo < visitorDays; daysAgo++ {
			if n <= dailyVisitors(daysAgo) {
				day := today.AddDate(0, 0, -daysAgo)
				at := time.Date(day.Year(), day.Month(), day.Day(), 8+n%12, 0, 0, 0, loc)
				if at.After(today) {
					at = today // nobody visits later today yet
				}
				seed.visits = append(seed.visits, at)
			}
		}
		ret = append(ret, seed)
	}
	return ret
}

// seedVisitors replaces the example visitors in the database named by the
// POSTGRES_* environment variables, the same ones the backend reads.
func seedVisitors(host string, loc *time.Location, now time.Time) (int, error) {
	db, err := psqlx.NewDBFromEnvVars()
	if err != nil {
		return 0, err
	}

	logs := 0
	err = db.Transaction(func(tx *gorm.DB) error {
		like := visitorIDPrefix + "%"
		if err := tx.Unscoped().Where("visitor_id like ?", like).Delete(&models.VisitorLog{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("visitor_id like ?", like).Delete(&models.VisitorLocation{}).Error; err != nil {
			return err
		}

		for _, v := range visitors(now, loc) {
			location := v.location
			if err := tx.Create(&location).Error; err != nil {
				return err
			}
			for _, at := range v.visits {
				log := models.VisitorLog{
					VisitorID:         location.VisitorID,
					Hostname:          host,
					PathName:          "/api/calendar",
					VisitorLocationID: location.ID,
				}
				log.CreatedAt, log.UpdatedAt = at.Unix(), at.Unix()
				if err := tx.Create(&log).Error; err != nil {
					return err
				}
				logs++
			}
		}
		return nil
	})
	return logs, err
}
