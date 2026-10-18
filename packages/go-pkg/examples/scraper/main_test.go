package main

import "testing"

func TestScrape(t *testing.T) {
	out := scrape()
	if len(out.Errors) != 0 {
		t.Fatalf("errors: %v", out.Errors)
	}
	if len(out.Events) != 2 {
		t.Fatalf("got %d events, want 2", len(out.Events))
	}
	e := out.Events[0]
	if e.Guid != "Example Venue-1001" {
		t.Errorf("guid = %q", e.Guid)
	}
	if got := e.StartAt.UTC().Format("15:04"); got != "01:00" { // 7pm CST is 01:00 UTC next day
		t.Errorf("start = %s UTC, want 01:00", got)
	}
}
