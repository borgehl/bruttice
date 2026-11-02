package mapper

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
)

// FormDateTimeFormat is the layout of a date and a time submitted by the add event forms.
const FormDateTimeFormat = "2006-01-02 15:04"

var (
	ErrInvalidStartTime = errors.New("invalid date or start time format")
	ErrInvalidDoorsTime = errors.New("invalid doors time format")
)

// AddEventRequest converts a public submission to an event that waits for verification
// and approval. It fails with ErrInvalidStartTime or ErrInvalidDoorsTime.
func AddEventRequest(req dto.AddEventRequest) (models.Event, error) {
	event, err := eventFromForm(formEvent{
		Title: req.Title, Link: req.Link, Venue: req.Venue, Region: req.Region, Age: req.Age,
		CoverCost: req.CoverCost, Date: req.Date, StartTime: req.StartTime, DoorsTime: req.DoorsTime,
		Tags: req.Tags,
	})
	if err != nil {
		return event, err
	}
	event.VerificationEmail = &req.Email
	event.CreateVerificationKey()
	event.UpdateGuid(nil)
	return event, nil
}

// formEvent holds the fields the public and admin add event requests share.
type formEvent struct {
	Title, Link, Venue, Region, Age, CoverCost string
	Date, StartTime, DoorsTime                 string
	Tags                                       []string
}

func eventFromForm(f formEvent) (models.Event, error) {
	var event models.Event
	event.Title = f.Title
	event.Link = NormalizeURL(f.Link)
	event.CoverCost = f.CoverCost
	event.Venue.Name = f.Venue
	event.Region.Name = f.Region
	event.Age.Name = f.Age

	startAt, err := time.ParseInLocation(FormDateTimeFormat, fmt.Sprintf("%s %s", f.Date, f.StartTime), util.TZLocation)
	if err != nil {
		return event, ErrInvalidStartTime
	}
	event.StartAt = startAt

	if f.DoorsTime != "" {
		doorsAt, err := time.ParseInLocation(FormDateTimeFormat, fmt.Sprintf("%s %s", f.Date, f.DoorsTime), util.TZLocation)
		if err != nil {
			return event, ErrInvalidDoorsTime
		}
		event.DoorsAt = &doorsAt
	}

	for _, tag := range f.Tags {
		event.AddTagStr(tag)
	}
	return event, nil
}

// NormalizeURL adds https:// prefix if no protocol is present
func NormalizeURL(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return url
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return "https://" + url
}

// FlyerURL is the flyer link to store: an http(s) URL as given, or empty if the value is
// blank or is not one. It is shown as an image, so other schemes are never kept.
func FlyerURL(url string) string {
	url = strings.TrimSpace(url)
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return ""
}
