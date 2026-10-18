package ingest_test

import (
	"testing"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/ingest"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

// Existing events are matched by guid, so the guid a scraper builds must not change.
func TestGuidMatchesTheDatabaseModel(t *testing.T) {
	loc, _ := time.LoadLocation("America/Chicago")
	start := time.Date(2030, 1, 10, 19, 0, 0, 0, loc)
	id := "1001"

	for _, tc := range []struct {
		name string
		id   *string
	}{{"source id", &id}, {"hash", nil}} {
		t.Run(tc.name, func(t *testing.T) {
			dto := ingest.Event{Title: "Show", Venue: "V", StartAt: start}
			dto.UpdateGuid(tc.id)

			model := models.Event{Title: "Show", Venue: models.Venue{Name: "V"}, StartAt: start}
			model.UpdateGuid(tc.id)

			if dto.Guid == "" || dto.Guid != model.Guid {
				t.Errorf("dto guid %q, model guid %q", dto.Guid, model.Guid)
			}
		})
	}
}

func TestUpdateGuidKeepsAnExistingGuid(t *testing.T) {
	e := ingest.Event{Guid: "mine", Title: "Show", Venue: "V"}
	e.UpdateGuid(nil)
	if e.Guid != "mine" {
		t.Errorf("guid = %q", e.Guid)
	}
}

func TestSetTimezoneCentralKeepsWallClock(t *testing.T) {
	doors := time.Date(2030, 1, 10, 18, 0, 0, 0, time.UTC)
	e := ingest.Event{StartAt: time.Date(2030, 1, 10, 19, 0, 0, 0, time.UTC), DoorsAt: &doors}
	e.SetTimezoneCentral()
	if got := e.StartAt.Format("15:04 -0700"); got != "19:00 -0600" {
		t.Errorf("start = %s", got)
	}
	if got := e.DoorsAt.Format("15:04 -0700"); got != "18:00 -0600" {
		t.Errorf("doors = %s", got)
	}
}
