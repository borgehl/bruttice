package main

import (
	"strings"
	"testing"
	"time"
)

func TestPayloads(t *testing.T) {
	loc, _ := time.LoadLocation(zone)
	now := time.Date(2030, 1, 1, 12, 0, 0, 0, loc)
	guids := map[string]bool{}

	for _, v := range venues {
		out := payload(v, loc, now)
		if out.VenueInfo.IsAutoIngest == v.pending {
			t.Errorf("%s: is_auto_ingest = %t", v.name, out.VenueInfo.IsAutoIngest)
		}
		for _, e := range out.Events {
			if guids[e.Guid] {
				t.Errorf("duplicate guid %s", e.Guid)
			}
			guids[e.Guid] = true
			if !e.StartAt.After(now) {
				t.Errorf("%s starts in the past", e.Title)
			}
		}
	}
}

func TestVisitors(t *testing.T) {
	loc, _ := time.LoadLocation(zone)
	now := time.Date(2030, 1, 1, 12, 0, 0, 0, loc)
	oldest := now.AddDate(0, 0, -visitorDays)

	ids := map[string]bool{}
	cities := map[string]bool{}
	perDay := map[string]bool{}
	for _, v := range visitors(now, loc) {
		id := v.location.VisitorID
		if !strings.HasPrefix(id, visitorIDPrefix) || len(id) != 36 {
			t.Errorf("visitor id %q is not a seed uuid", id)
		}
		if ids[id] {
			t.Errorf("duplicate visitor id %s", id)
		}
		ids[id] = true
		cities[v.location.City] = true

		for _, at := range v.visits {
			if at.After(now) || at.Before(oldest) {
				t.Errorf("%s: visit at %s is outside the last %d days", id, at, visitorDays)
			}
			perDay[at.Format("2006-01-02")] = true
		}
	}

	if len(ids) != visitorCount {
		t.Errorf("%d visitors, want %d", len(ids), visitorCount)
	}
	// several places, including the visitors without a location
	if len(cities) < len(places) {
		t.Errorf("%d different places, want %d", len(cities), len(places))
	}
	if len(perDay) < visitorDays/2 {
		t.Errorf("visits on only %d days", len(perDay))
	}
}

// The Visitors page scales to the busiest day of its range, so the seed gives the last 7 days and
// the last 30 days different ones.
func TestVisitorsBusiestDays(t *testing.T) {
	loc, _ := time.LoadLocation(zone)
	now := time.Date(2030, 1, 1, 12, 0, 0, 0, loc)

	// different visitors on each day, as the page counts them
	onDay := map[string]map[string]bool{}
	for _, v := range visitors(now, loc) {
		for _, at := range v.visits {
			day := at.Format("2006-01-02")
			if onDay[day] == nil {
				onDay[day] = map[string]bool{}
			}
			onDay[day][v.location.VisitorID] = true
		}
	}

	// the busiest day among the last `days` days, today included
	busiest := func(days int) int {
		most := 0
		for daysAgo := 0; daysAgo < days; daysAgo++ {
			day := now.AddDate(0, 0, -daysAgo).Format("2006-01-02")
			if n := len(onDay[day]); n > most {
				most = n
			}
		}
		return most
	}

	if got := busiest(7); got != busiestDayOfWeek {
		t.Errorf("busiest day of the last 7 days has %d visitors, want %d", got, busiestDayOfWeek)
	}
	if got := busiest(30); got != busiestDayOfMonth {
		t.Errorf("busiest day of the last 30 days has %d visitors, want %d", got, busiestDayOfMonth)
	}
	if got := busiest(visitorDays); got != busiestDayOfMonth {
		t.Errorf("busiest day of the last %d days has %d visitors, want %d", visitorDays, got, busiestDayOfMonth)
	}
	if len(onDay[now.Format("2006-01-02")]) == 0 {
		t.Error("nobody visited today")
	}
}

func TestCancelledListingsAreSent(t *testing.T) {
	loc, _ := time.LoadLocation(zone)
	now := time.Date(2030, 1, 1, 12, 0, 0, 0, loc)
	var live, cancelled, cancelledPending int

	for _, v := range venues {
		for _, e := range payload(v, loc, now).Events {
			switch {
			case !e.Cancelled:
				live++
			case v.pending:
				cancelledPending++
			default:
				cancelled++
			}
		}
	}
	if live == 0 || cancelled == 0 || cancelledPending == 0 {
		t.Errorf("live %d, cancelled %d, cancelled in the moderation queue %d", live, cancelled, cancelledPending)
	}
}
