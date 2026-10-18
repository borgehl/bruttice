package ingest

import (
	"fmt"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
)

type VenueInfo struct {
	Name string `json:"name"`
	Link string `json:"link"`
	// IsAutoIngest marks the venue as fed by scrapers. The server refuses events for a
	// venue whose flag differs, so one venue is never fed by both scrapers and people.
	IsAutoIngest bool `json:"is_auto_ingest"`
	// RequireApproval sends the payload's events to the admin moderation queue. By
	// default they are published on arrival.
	RequireApproval bool   `json:"require_approval,omitempty"`
	DefaultRegion   string `json:"default_region"`
	SkipVenueSave   bool   `json:"skip_venue_save"`
	Timezone        string `json:"timezone"`
}

// Event is one event as a scraper reports it. It is the wire format of the ingest
// API, independent of how the server stores events.
type Event struct {
	// Guid is the event's identity, a re-sent event with the same guid updates in place.
	// Leave it empty to let the server derive one, or use UpdateGuid for a stable source id.
	Guid string `json:"guid,omitempty"`

	Title     string     `json:"title"`
	Link      string     `json:"link"`
	StartAt   time.Time  `json:"start_at"`
	DoorsAt   *time.Time `json:"doors_at,omitempty"`
	CoverCost string     `json:"cover_cost"`

	// Venue is a venue name. Defaults to the payload's venue.
	Venue string `json:"venue,omitempty"`
	// Region is a region name. For a listed venue the venue's default region wins.
	Region string `json:"region,omitempty"`
	// Age is an age group name.
	Age string `json:"age"`
	// Cancelled marks an event the source has cancelled. A new cancelled event is stored
	// as deleted, so it never shows.
	Cancelled bool `json:"cancelled,omitempty"`
	// Tags are tag names. Names the server does not have are dropped.
	Tags []string `json:"tags,omitempty"`
}

type IngestionOutput struct {
	// RunID groups payloads into one run, 0 if the payload is not part of a run
	RunID     int64     `json:"run_id,omitempty"`
	VenueInfo VenueInfo `json:"venue_info"`
	Events    []Event   `json:"events"`
	Errors    []string  `json:"errors"`
}

// UpdateGuid sets Guid unless it is already set. With a non-empty source id the guid is
// "<Venue>-<id>", otherwise it hashes title, start time and venue. Set Venue, Title and
// StartAt first.
func (e *Event) UpdateGuid(id *string) {
	if e.Guid != "" {
		return
	}
	e.Guid = util.EventGuid(e.Venue, e.Title, e.StartAt, id)
}

func (e *Event) AddTag(tag string) {
	e.Tags = append(e.Tags, tag)
}

// SetTimezoneCentral reads the wall-clock times of StartAt and DoorsAt as US Central.
func (e *Event) SetTimezoneCentral() {
	e.StartAt = util.DatetimeIsCentral(e.StartAt)
	if e.DoorsAt != nil {
		doors := util.DatetimeIsCentral(*e.DoorsAt)
		e.DoorsAt = &doors
	}
}

func (io *IngestionOutput) AppendErrors(errs []error) {
	for _, err := range errs {
		io.Errors = append(io.Errors, fmt.Sprintf("%s: %s", io.VenueInfo.Name, err))
	}
}

// append error to io
//
// return immediately if err is nil
func (io *IngestionOutput) AppendError(err error) {
	if err == nil {
		return
	}
	io.Errors = append(io.Errors, fmt.Sprintf("%s: %s", io.VenueInfo.Name, err))
}

func (io *IngestionOutput) GetErrors() []error {
	ret := make([]error, 0, len(io.Errors))
	for _, e := range io.Errors {
		ret = append(ret, fmt.Errorf("%s", e))
	}
	return ret
}
