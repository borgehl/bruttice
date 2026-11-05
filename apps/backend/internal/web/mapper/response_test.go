package mapper

import (
	"testing"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

func event(title string, start time.Time) models.Event {
	return models.Event{
		Guid: title, Title: title, StartAt: start,
		Venue: models.Venue{Name: "V", Link: "https://v.example.com"},
		Age:   models.Age{Name: "21+"}, Region: models.Region{Name: "central"},
		Tags: []models.Tag{{Name: "music"}},
	}
}

func TestEventResponseDoors(t *testing.T) {
	start := time.Date(2030, 1, 10, 19, 0, 0, 0, time.UTC)

	e := event("Show", start)
	if got := EventResponse(e).DoorsAt; got != "" {
		t.Errorf("no doors: %q", got)
	}

	same := start
	e.DoorsAt = &same
	if got := EventResponse(e).DoorsAt; got != "" {
		t.Errorf("doors equal to start are left out: %q", got)
	}

	doors := start.Add(-time.Hour)
	e.DoorsAt = &doors
	resp := EventResponse(e)
	if resp.DoorsAt != "2030-01-10T18:00:00Z" {
		t.Errorf("doors = %q", resp.DoorsAt)
	}
	if resp.Venue != "V" || resp.VenueLink != "https://v.example.com" || len(resp.Tags) != 1 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestAdminEventDayResponsesGroupsByDay(t *testing.T) {
	day1 := time.Date(2030, 1, 10, 19, 0, 0, 0, time.UTC)
	events := []models.Event{
		event("a", day1),
		event("b", day1.Add(time.Hour)),
		event("c", day1.AddDate(0, 0, 1)),
	}
	events[0].EventAdminJudgements = []models.EventAdminJudgement{{IsApproval: true}, {IsApproval: true}, {IsApproval: false}}
	email := "s@example.com"
	events[1].VerificationEmail = &email

	days := AdminEventDayResponses(events)
	if len(days) != 2 || len(days[0].Events) != 2 || len(days[1].Events) != 1 {
		t.Fatalf("days = %+v", days)
	}
	if days[0].Events[0].Score != 1 {
		t.Errorf("score = %d, want 2 approvals - 1 denial", days[0].Events[0].Score)
	}
	if days[0].Events[1].Email != email {
		t.Errorf("email = %q", days[0].Events[1].Email)
	}
	if got := AdminEventDayResponses(nil); got == nil || len(got) != 0 {
		t.Error("no events should be an empty list, not null")
	}
}

func TestFilterOptionsLeavesOutUnknownAge(t *testing.T) {
	resp := FilterOptionsResponse(
		[]string{"music", "comedy"},
		[]models.Venue{{Name: "V", Slug: "v"}},
		[]models.Age{{Name: "21+", Value: 21}, {Name: "unknown age limit", Value: 255}},
		[]string{"central"},
	)
	if len(resp.Ages) != 1 || resp.Ages[0].Name != "21+" {
		t.Errorf("ages = %+v", resp.Ages)
	}
	if len(resp.Tags) != 2 || resp.Tags[1].ID != 2 || resp.Regions[0].ID != 1 {
		t.Errorf("tags %+v regions %+v", resp.Tags, resp.Regions)
	}
	if len(resp.Venues) != 1 || resp.Venues[0].Slug != "v" {
		t.Errorf("venues = %+v", resp.Venues)
	}
}

func TestIngestRunResponsesTotalsSources(t *testing.T) {
	runs := []models.IngestRun{{
		BaseModel: models.BaseModel{ID: 3},
		Sources: []models.IngestRunSource{
			{Source: "a", ScrapedCount: 5, SavedCount: 4, FailedCount: 1, Errors: []string{"x"}},
			{Source: "b", ScrapedCount: 2, SavedCount: 2},
		},
	}}
	resp := IngestRunResponses(runs)
	if resp[0].Scraped != 7 || resp[0].Saved != 6 || resp[0].Failed != 1 {
		t.Errorf("totals = %+v", resp[0])
	}
	if resp[0].Sources[1].Errors == nil {
		t.Error("a source with no errors should list [], not null")
	}
}

func TestEventResponseCancelledAt(t *testing.T) {
	event := models.Event{Guid: "g", StartAt: time.Date(2030, 1, 10, 19, 0, 0, 0, time.UTC)}
	if got := EventResponse(event).CancelledAt; got != "" {
		t.Errorf("not cancelled: cancelledAt = %q", got)
	}
	at := int64(1_900_000_000)
	event.CancelledAt = &at
	want := time.Unix(at, 0).UTC().Format(time.RFC3339)
	if got := EventResponse(event).CancelledAt; got != want {
		t.Errorf("cancelledAt = %q, want %q", got, want)
	}
	if got := adminEventResponse(event).CancelledAt; got != want {
		t.Errorf("admin cancelledAt = %q, want %q", got, want)
	}
}
