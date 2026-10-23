package mapper

import (
	"testing"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/ingest"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/configx"
)

func TestFlyerURL(t *testing.T) {
	cases := map[string]string{
		"":                           "",
		"   ":                        "",
		"https://x.example/a.jpg":    "https://x.example/a.jpg",
		" http://x.example/a.jpg ":   "http://x.example/a.jpg",
		"javascript:alert(1)":        "",
		"data:image/png;base64,AAAA": "",
		"x.example/a.jpg":            "",
	}
	for in, want := range cases {
		if got := FlyerURL(in); got != want {
			t.Errorf("FlyerURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFlyerReachesResponses(t *testing.T) {
	e := event("Show", time.Date(2030, 1, 10, 19, 0, 0, 0, time.UTC))
	if got := EventResponse(e).FlyerURL; got != "" {
		t.Errorf("no flyer: %q", got)
	}
	e.FlyerURL = "https://x.example/a.jpg"
	if got := EventResponse(e).FlyerURL; got != e.FlyerURL {
		t.Errorf("public flyer = %q", got)
	}
	if got := adminEventResponse(e).FlyerURL; got != e.FlyerURL {
		t.Errorf("admin flyer = %q", got)
	}
}

func TestIngestEventFlyer(t *testing.T) {
	now := time.Now()
	in := ingest.Event{Title: "T", StartAt: now, Age: "21+", FlyerURL: "https://x.example/a.jpg"}
	if got := IngestEvent(in, ingest.VenueInfo{Name: "V"}, now).FlyerURL; got != in.FlyerURL {
		t.Errorf("flyer = %q", got)
	}
	in.FlyerURL = "javascript:alert(1)"
	if got := IngestEvent(in, ingest.VenueInfo{Name: "V"}, now).FlyerURL; got != "" {
		t.Errorf("a non-http flyer is dropped: %q", got)
	}
}

func TestVenueResponsesRegion(t *testing.T) {
	got := VenueResponses([]models.Venue{{Name: "A", DefaultRegion: "tulsa"}, {Name: "B"}})
	if got[0].Region != "tulsa" || got[1].Region != "" {
		t.Errorf("regions = %q, %q", got[0].Region, got[1].Region)
	}
}

func TestSiteResponse(t *testing.T) {
	got := SiteResponse(configx.ConfigRawSite{Name: "N", Region: "R", ContactEmail: "a@b.c", Description: "D"})
	if got.Name != "N" || got.Region != "R" || got.ContactForm || got.Description != "D" {
		t.Errorf("site = %+v", got)
	}
}
