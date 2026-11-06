package models

import (
	"encoding/json"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"math"
	"time"
)

const ToDateDefaultDays = 30

// the filter struct used
// the bools indicate if the list is in/exclude, lt or gt
type CalendarFilterBackend struct {
	Tags        []string
	TagsInclude bool

	Age          uint8
	AgesLessThan bool

	Venues        []string
	VenuesInclude bool

	// Regions are the names of the regions a saved filter shows, all of them when empty. omitempty
	// keeps the stored body, and so the key, of every filter without regions as it was.
	Regions []string `json:",omitempty"`

	// RequestRegions are the regions one request asks for, from the calendar's region picker. They
	// are not saved. They narrow the filter's own Regions: an event has to be in both lists, so a
	// filter for one region and a request for another matches nothing.
	RequestRegions []string `json:"-"`

	// dates
	FromDate time.Time
	ToDate   time.Time

	// offsets to date are relative to today for from
	// relative to ToDateDefaultDays for to
	FromOffsetDays int
	ToOffsetDays   int
}

type CalendarFilter struct {
	BaseModel
	HashKey string
	Body    []byte
}

func (c *CalendarFilterBackend) GetDBFilter() CalendarFilter {
	body, err := json.Marshal(c)
	if err != nil {
		return CalendarFilter{}
	}
	return CalendarFilter{
		HashKey: util.Sha256Base64Bytes(body),
		Body:    body,
	}
}

// sets FromDate and ToDate relative to today
// takinginto account FromOffsetDays and ToOffsetDays
func (c *CalendarFilterBackend) UpdateDates() {
	// day math must be in the local zone, AddDate on a UTC time ignores DST
	today := util.NowDate().In(util.TZLocation)
	c.FromDate = today.AddDate(0, 0, c.FromOffsetDays)

	// add 24 hours to get all events on the day
	c.ToDate = today.AddDate(0, 0, ToDateDefaultDays).AddDate(0, 0, c.ToOffsetDays).Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999*time.Microsecond)
}

// wholeDaysBetween returns a - b in days, rounded. A local day is 23 or 25 hours
// across a DST change, so truncating hours/24 is off by one for far-off dates.
func wholeDaysBetween(a time.Time, b time.Time) int {
	return int(math.Round(a.Sub(b).Hours() / 24.0))
}

func (c *CalendarFilterBackend) CustomOffsetsToDates(fromDateStr string, toDateStr string) error {
	today := util.NowDate()

	if fromDateStr != "" {
		fromDate, err := time.ParseInLocation("2006-01-02", fromDateStr, util.TZLocation)
		if err != nil {
			return err
		}
		c.FromOffsetDays = -wholeDaysBetween(today, fromDate)
	}

	if toDateStr != "" {
		toDate, err := time.ParseInLocation("2006-01-02", toDateStr, util.TZLocation)
		if err != nil {
			return err
		}
		defaultToDate := today.AddDate(0, 0, ToDateDefaultDays)
		c.ToOffsetDays = -wholeDaysBetween(defaultToDate, toDate)
	}

	c.UpdateDates()
	return nil
}

func (c *CalendarFilter) GetBackendFilter() CalendarFilterBackend {
	ret := CalendarFilterBackend{}
	if err := json.Unmarshal(c.Body, &ret); err != nil {
		return CalendarFilterBackend{}
	}
	return ret
}
