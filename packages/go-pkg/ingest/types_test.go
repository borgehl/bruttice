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

func TestValidate(t *testing.T) {
	start := time.Date(2030, 1, 10, 19, 0, 0, 0, time.UTC)
	valid := ingest.Event{Title: "Show", StartAt: start, Age: "21+"}

	if err := valid.Validate("V"); err != nil {
		t.Errorf("a complete event: %v", err)
	}
	if err := valid.Validate(""); err == nil {
		t.Error("no venue on the event or the payload should fail")
	}
	own := valid
	own.Venue = "Other"
	if err := own.Validate(""); err != nil {
		t.Errorf("the event's own venue is enough: %v", err)
	}

	for name, e := range map[string]ingest.Event{
		"title":    {StartAt: start, Age: "21+"},
		"blank":    {Title: "  ", StartAt: start, Age: "21+"},
		"start_at": {Title: "Show", Age: "21+"},
		"age":      {Title: "Show", StartAt: start},
	} {
		if err := e.Validate("V"); err == nil {
			t.Errorf("missing %s was accepted", name)
		}
	}

	err := ingest.Event{}.Validate("V")
	if err == nil || err.Error() != "missing title, start_at, age" {
		t.Errorf("err = %v", err)
	}
}

func TestCancellationOnly(t *testing.T) {
	start := time.Date(2030, 1, 10, 19, 0, 0, 0, time.UTC)
	for name, tc := range map[string]struct {
		event ingest.Event
		only  bool
	}{
		"stub without start":       {ingest.Event{Cancelled: true, Age: "21+"}, true},
		"cancelled without age":    {ingest.Event{Cancelled: true, StartAt: start}, true},
		"complete cancelled event": {ingest.Event{Cancelled: true, StartAt: start, Age: "21+"}, false},
		"not cancelled":            {ingest.Event{StartAt: start, Age: "21+"}, false},
		"not cancelled, no start":  {ingest.Event{Age: "21+"}, false},
	} {
		if got := tc.event.IsCancellationOnly(); got != tc.only {
			t.Errorf("%s: IsCancellationOnly = %v, want %v", name, got, tc.only)
		}
	}
}

func TestValidateCancellationOnlyNeedsGuidAndVenue(t *testing.T) {
	stub := ingest.Event{Guid: "V-abc", Cancelled: true}
	if err := stub.Validate("V"); err != nil {
		t.Errorf("a stub with a guid: %v", err)
	}
	if err := stub.Validate(""); err == nil {
		t.Error("a stub with no venue was accepted")
	}
	stub.Guid = ""
	if err := stub.Validate("V"); err == nil {
		t.Error("a stub with no guid was accepted")
	}
	// not cancelled: the full rules still apply
	if err := (ingest.Event{Guid: "V-abc"}).Validate("V"); err == nil {
		t.Error("an event that is not cancelled was accepted without start_at and age")
	}
}

func TestSetTimezoneCentralKeepsAZeroStartZero(t *testing.T) {
	e := ingest.Event{Cancelled: true, Age: "21+"}
	e.SetTimezoneCentral()
	if !e.StartAt.IsZero() {
		t.Errorf("start = %v", e.StartAt)
	}
	if !e.IsCancellationOnly() {
		t.Error("a stub stopped being a cancellation-only event")
	}
}
