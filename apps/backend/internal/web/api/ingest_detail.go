package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/ingest"
)

// eventError is the message for an event that could not be ingested. The first line says which
// event and why, the lines under it say what the scraper sent, so the event can be found at the
// source without running the scraper again. The admin dashboard and the error email show them.
//
//	event 28 ("Closed for Thanksgiving"): missing start_at
//	  link: https://venue.example/events/28
//	  start_at: (none)
//	  age: all-ages
func eventError(index int, e ingest.Event, payloadVenue string, reason error) string {
	var b strings.Builder
	fmt.Fprintf(&b, "event %d (%q): %s", index, e.Title, reason)

	line := func(key, value string) { fmt.Fprintf(&b, "\n  %s: %s", key, oneLineDetail(value)) }
	// the fields an event is refused for are always shown, so the one that is empty is plain to see
	required := func(key, value string) {
		if strings.TrimSpace(value) == "" {
			value = "(none)"
		}
		line(key, value)
	}

	if e.Link != "" {
		line("link", e.Link)
	}
	if e.Guid != "" {
		line("guid", e.Guid)
	}
	venue := e.Venue
	if venue == "" && payloadVenue != "" {
		venue = payloadVenue + " (the payload's)"
	}
	required("venue", venue)
	if e.StartAt.IsZero() {
		required("start_at", "")
	} else {
		line("start_at", e.StartAt.Format(time.RFC3339))
	}
	if e.DoorsAt != nil && !e.DoorsAt.IsZero() {
		line("doors_at", e.DoorsAt.Format(time.RFC3339))
	}
	required("age", e.Age)
	if e.Region != "" {
		line("region", e.Region)
	}
	if e.CoverCost != "" {
		line("cover_cost", e.CoverCost)
	}
	if len(e.Tags) > 0 {
		line("tags", strings.Join(e.Tags, ", "))
	}
	if e.FlyerURL != "" {
		line("flyer_url", e.FlyerURL)
	}
	if e.Cancelled {
		line("cancelled", "yes")
	}
	return b.String()
}

// oneLineDetail keeps a value from spilling onto more lines than its own, which would be read as
// more details, and from growing without limit.
func oneLineDetail(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > 300 {
		return string(r[:300]) + "…"
	}
	return s
}
