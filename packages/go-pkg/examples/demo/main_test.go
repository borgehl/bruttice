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
			if e.VerifiedAt == nil || (e.ApprovedAt == nil) != v.pending {
				t.Errorf("%s: wrong approval for venue pending=%t", e.Title, v.pending)
			}
		}
	}
}
