package main

import (
	"bytes"
	"image/png"
	"net/http"
	"testing"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
)

func TestDemoSponsorImageIsAPictureTheServerTakes(t *testing.T) {
	image := demoSponsorImage()
	if len(image) > dbx.MaxSponsorImage {
		t.Errorf("the picture is %d bytes, the limit is %d", len(image), dbx.MaxSponsorImage)
	}
	if kind := http.DetectContentType(image); kind != "image/png" {
		t.Errorf("type = %q", kind)
	}
	if _, err := png.Decode(bytes.NewReader(image)); err != nil {
		t.Errorf("the picture does not decode: %v", err)
	}
}

func TestSponsorCountsCoverTheLastDays(t *testing.T) {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 10, 15, 0, 0, 0, loc)
	counts := sponsorCounts(7, now, loc)

	if len(counts) != sponsorDays {
		t.Fatalf("got %d days, want %d", len(counts), sponsorDays)
	}
	if counts[0].Date != "2026-11-10" || counts[1].Date != "2026-11-09" {
		t.Errorf("dates start %s, %s", counts[0].Date, counts[1].Date)
	}
	seen := map[string]bool{}
	for _, c := range counts {
		if seen[c.Date] {
			t.Errorf("%s twice: the date is part of the key", c.Date)
		}
		seen[c.Date] = true
		if c.SponsorID != 7 || c.Impressions <= 0 || c.Clicks < 0 || c.Clicks > c.Impressions {
			t.Errorf("bad count %+v", c)
		}
	}
}
