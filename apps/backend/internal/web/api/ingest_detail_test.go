package api

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/ingest"
)

func TestEventErrorSaysWhatTheScraperSent(t *testing.T) {
	doors := time.Date(2030, 1, 10, 18, 0, 0, 0, time.FixedZone("", -6*3600))
	e := ingest.Event{
		Title: "NUNS", Link: "https://venue.example/events/13", Guid: "The Vanguard-13",
		StartAt: time.Date(2030, 1, 10, 19, 30, 0, 0, time.FixedZone("", -6*3600)), DoorsAt: &doors,
		CoverCost: "$10", Region: "tulsa", Tags: []string{"music", "punk"}, FlyerURL: "https://venue.example/f.jpg", Cancelled: true,
	}
	got := eventError(13, e, "The Vanguard", fmt.Errorf("missing age"))
	want := `event 13 ("NUNS"): missing age
  link: https://venue.example/events/13
  guid: The Vanguard-13
  venue: The Vanguard (the payload's)
  start_at: 2030-01-10T19:30:00-06:00
  doors_at: 2030-01-10T18:00:00-06:00
  age: (none)
  region: tulsa
  cover_cost: $10
  tags: music, punk
  flyer_url: https://venue.example/f.jpg
  cancelled: yes`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestEventErrorShowsTheMissingFieldsPlainly(t *testing.T) {
	got := eventError(28, ingest.Event{Title: "Closed for Thanksgiving", Age: "all-ages", Venue: "Bricktown Comedy Club"}, "Bricktown Comedy Club", fmt.Errorf("missing start_at"))
	for _, want := range []string{
		`event 28 ("Closed for Thanksgiving"): missing start_at`,
		"  venue: Bricktown Comedy Club\n",
		"  start_at: (none)\n",
		"  age: all-ages",
	} {
		if !strings.Contains(got+"\n", want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "link:") || strings.Contains(got, "guid:") {
		t.Errorf("fields that are empty are left out:\n%s", got)
	}
}

func TestEventErrorKeepsEachDetailOnOneLine(t *testing.T) {
	e := ingest.Event{Title: "x", Link: "https://a.example/\n  age: forged", Age: "all-ages", Venue: "V"}
	got := eventError(0, e, "V", fmt.Errorf("bad"))
	if strings.Count(got, "\n  age:") != 1 {
		t.Errorf("a value must not make a detail of its own:\n%s", got)
	}
	long := eventError(0, ingest.Event{Title: "x", Link: strings.Repeat("a", 1000)}, "V", fmt.Errorf("bad"))
	if len(long) > 1000 {
		t.Errorf("a detail is cut short, got %d bytes", len(long))
	}
}
