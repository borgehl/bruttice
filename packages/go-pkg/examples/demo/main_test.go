package main

import (
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
