package api

import (
	"testing"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/ingest"
)

func TestIngestEventToModel(t *testing.T) {
	now := time.Unix(1_000, 0)
	scraped := ingest.Event{Title: "Show", Age: "21+", Tags: []string{"music", "free"}}

	t.Run("approved on arrival, with venue and region defaults", func(t *testing.T) {
		event := ingestEventToModel(scraped, ingest.VenueInfo{Name: "V", DefaultRegion: "central", IsAutoIngest: true}, now)
		if event.Venue.Name != "V" || event.Region.Name != "central" || event.Age.Name != "21+" {
			t.Errorf("got venue %q region %q age %q", event.Venue.Name, event.Region.Name, event.Age.Name)
		}
		if event.VerifiedAt == nil || event.ApprovedAt == nil || *event.ApprovedAt != 1_000 {
			t.Errorf("verified %v approved %v", event.VerifiedAt, event.ApprovedAt)
		}
		if event.VerificationEmail == nil || *event.VerificationEmail != ingestVerificationEmail {
			t.Error("an ingested event needs a submitter for the approval notification")
		}
		if len(event.Tags) != 2 || event.Tags[0].Name != "music" {
			t.Errorf("tags = %v", event.Tags)
		}
	})

	t.Run("require_approval sends events to the moderation queue", func(t *testing.T) {
		event := ingestEventToModel(scraped, ingest.VenueInfo{Name: "V", RequireApproval: true}, now)
		if event.VerifiedAt == nil || event.ApprovedAt != nil {
			t.Errorf("verified %v approved %v", event.VerifiedAt, event.ApprovedAt)
		}
	})

	t.Run("the event's own venue and region win", func(t *testing.T) {
		own := scraped
		own.Venue, own.Region = "Other", "eastern"
		event := ingestEventToModel(own, ingest.VenueInfo{Name: "V", DefaultRegion: "central"}, now)
		if event.Venue.Name != "Other" || event.Region.Name != "eastern" {
			t.Errorf("got venue %q region %q", event.Venue.Name, event.Region.Name)
		}
	})
}
