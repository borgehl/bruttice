package mapper

import (
	"errors"
	"testing"
	"time"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
)

func validAddEvent() dto.AddEventRequest {
	return dto.AddEventRequest{
		Title: "Show", Link: "venue.example.com/show", Venue: "V", Region: "central", Age: "21+",
		Date: "2030-01-10", StartTime: "19:30", DoorsTime: "18:30", CoverCost: "$10",
		Tags: []string{"music"}, Email: "a@example.com",
	}
}

func TestAddEventRequest(t *testing.T) {
	event, err := AddEventRequest(validAddEvent())
	if err != nil {
		t.Fatal(err)
	}
	if event.Link != "https://venue.example.com/show" {
		t.Errorf("link = %q", event.Link)
	}
	if event.Venue.Name != "V" || event.Region.Name != "central" || event.Age.Name != "21+" {
		t.Errorf("venue %q region %q age %q", event.Venue.Name, event.Region.Name, event.Age.Name)
	}
	if got := event.StartAt.Format("2006-01-02 15:04"); got != "2030-01-10 19:30" {
		t.Errorf("start = %s", got)
	}
	if event.DoorsAt == nil || event.DoorsAt.Format("15:04") != "18:30" {
		t.Errorf("doors = %v", event.DoorsAt)
	}
	if len(event.Tags) != 1 || event.Tags[0].Name != "music" {
		t.Errorf("tags = %v", event.Tags)
	}
	if event.VerificationEmail == nil || *event.VerificationEmail != "a@example.com" || event.VerificationKey == nil {
		t.Error("a public submission needs its submitter and a verification key")
	}
	if event.VerifiedAt != nil || event.ApprovedAt != nil {
		t.Error("a public submission starts unverified and unapproved")
	}
	if event.Guid == "" {
		t.Error("guid not set")
	}
}

func TestAddEventRequestErrors(t *testing.T) {
	bad := validAddEvent()
	bad.StartTime = "tonight"
	if _, err := AddEventRequest(bad); !errors.Is(err, ErrInvalidStartTime) {
		t.Errorf("err = %v, want ErrInvalidStartTime", err)
	}

	bad = validAddEvent()
	bad.DoorsTime = "early"
	if _, err := AddEventRequest(bad); !errors.Is(err, ErrInvalidDoorsTime) {
		t.Errorf("err = %v, want ErrInvalidDoorsTime", err)
	}

	noDoors := validAddEvent()
	noDoors.DoorsTime = ""
	event, err := AddEventRequest(noDoors)
	if err != nil || event.DoorsAt != nil {
		t.Errorf("doors are optional, got %v, %v", event.DoorsAt, err)
	}
}

func TestAdminAddEventRequestIsApproved(t *testing.T) {
	now := time.Unix(5_000, 0)
	event, err := AdminAddEventRequest(dto.AdminAddEventRequest{
		Title: "Show", Link: "https://v.example.com", Venue: "V", Date: "2030-01-10", StartTime: "19:00",
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if event.VerifiedAt == nil || *event.VerifiedAt != 5_000 || event.ApprovedAt == nil || *event.ApprovedAt != 5_000 {
		t.Errorf("verified %v approved %v", event.VerifiedAt, event.ApprovedAt)
	}
	if event.VerificationEmail != nil {
		t.Error("an admin event has no submitter email")
	}
}

func TestNormalizeURL(t *testing.T) {
	for in, want := range map[string]string{
		"":                   "",
		"  ":                 "",
		"example.com":        "https://example.com",
		" example.com ":      "https://example.com",
		"http://example.com": "http://example.com",
		"https://x.com/a":    "https://x.com/a",
	} {
		if got := NormalizeURL(in); got != want {
			t.Errorf("NormalizeURL(%q) = %q, want %q", in, got, want)
		}
	}
}
