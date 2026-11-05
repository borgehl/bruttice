// Package mapper converts between the request and response bodies in package dto and the
// database models. Each file is named after the body it maps.
package mapper

import (
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/ingest"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

// IngestVerificationEmail stands in for the submitter of an ingested event. The
// approval notification lists the submitter of every event awaiting approval.
const IngestVerificationEmail = "ingest@internal.scraper"

// IngestEvent converts a scraper's event into a database event.
//
// Events are approved on arrival unless the payload asks for moderation, in which case
// they are verified only and wait in the moderation queue.
func IngestEvent(e ingest.Event, venue ingest.VenueInfo, now time.Time) models.Event {
	verifiedAt := now.Unix()
	email := IngestVerificationEmail

	event := models.Event{
		Guid:              e.Guid,
		Title:             e.Title,
		Link:              e.Link,
		StartAt:           e.StartAt,
		DoorsAt:           e.DoorsAt,
		CoverCost:         e.CoverCost,
		FlyerURL:          FlyerURL(e.FlyerURL),
		Venue:             models.Venue{Name: firstNonEmpty(e.Venue, venue.Name)},
		Region:            models.Region{Name: firstNonEmpty(e.Region, venue.DefaultRegion)},
		Age:               models.Age{Name: e.Age},
		VerifiedAt:        &verifiedAt,
		VerificationEmail: &email,
	}
	if e.Cancelled {
		cancelledAt := now.Unix()
		event.CancelledAt = &cancelledAt
	}
	if !venue.RequireApproval {
		approvedAt := verifiedAt
		event.ApprovedAt = &approvedAt
	}
	for _, tag := range e.Tags {
		event.AddTagStr(tag)
	}
	return event
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
