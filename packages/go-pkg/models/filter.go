package models

import (
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"encoding/json"
	"time"
)

const ToDateDefaultDays = 30

type CalendarFilterBackend struct {
	Tags        []string
	TagsInclude bool

	Age          uint8
	AgesLessThan bool

	Venues        []string
	VenuesInclude bool

	FromDate time.Time
	ToDate   time.Time

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

func (c *CalendarFilterBackend) UpdateDates() {
	today := util.NowDate()
	c.FromDate = today.AddDate(0, 0, c.FromOffsetDays)
	c.ToDate = today.AddDate(0, 0, ToDateDefaultDays).AddDate(0, 0, c.ToOffsetDays).Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999*time.Microsecond)
}

func (c *CalendarFilterBackend) CustomOffsetsToDates(fromDateStr string, toDateStr string) error {
	today := util.NowDate()

	if fromDateStr != "" {
		fromDate, err := time.ParseInLocation("2006-01-02", fromDateStr, util.TZLocation)
		if err != nil {
			return err
		}
		c.FromOffsetDays = -int(today.Sub(fromDate).Hours() / 24.0)
	}

	if toDateStr != "" {
		toDate, err := time.ParseInLocation("2006-01-02", toDateStr, util.TZLocation)
		if err != nil {
			return err
		}
		defaultToDate := today.AddDate(0, 0, ToDateDefaultDays)
		c.ToOffsetDays = -int(defaultToDate.Sub(toDate).Hours() / 24.0)
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
