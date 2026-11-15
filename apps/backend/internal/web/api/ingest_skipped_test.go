package api

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/ingest"
	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

func TestDeletedEventError(t *testing.T) {
	deleted := gorm.DeletedAt{Time: time.Unix(1_000, 0), Valid: true}
	for name, tc := range map[string]struct {
		existing models.Event
		want     error
	}{
		"never seen":      {models.Event{}, nil},
		"live":            {models.Event{BaseModel: models.BaseModel{ID: 5}}, nil},
		"deleted":         {models.Event{BaseModel: models.BaseModel{ID: 5, DeletedAt: deleted}}, errEventDeleted},
		"no row, deleted": {models.Event{BaseModel: models.BaseModel{DeletedAt: deleted}}, nil},
	} {
		if got := deletedEventError(tc.existing); !errors.Is(got, tc.want) {
			t.Errorf("%s: got %v, want %v", name, got, tc.want)
		}
	}
}

func TestRecordEventResultSaved(t *testing.T) {
	var source models.IngestRunSource
	got := recordEventResult(&source, 0, ingest.Event{Title: "Show"}, "Hall", nil)
	if got != eventSaved || source.SavedCount != 1 || source.SkippedCount != 0 || source.FailedCount != 0 {
		t.Errorf("outcome %v, source %+v", got, source)
	}
	if len(source.Errors) != 0 || len(source.Skipped) != 0 {
		t.Errorf("a saved event must add no text: errors %v, skipped %v", source.Errors, source.Skipped)
	}
}

func TestRecordEventResultSkippedIsNotSavedOrAnError(t *testing.T) {
	scraped := ingest.Event{Title: "Twen", Link: "https://hall.example/twen", Guid: "Hall-https://hall.example/twen"}
	// a skip stays a skip when something wraps it
	for name, err := range map[string]error{
		"plain":   errEventDeleted,
		"wrapped": fmt.Errorf("savepoint: %w", errEventDeleted),
	} {
		var source models.IngestRunSource
		got := recordEventResult(&source, 3, scraped, "Hall", err)

		if got != eventSkipped {
			t.Errorf("%s: outcome = %v, want skipped", name, got)
		}
		if source.SkippedCount != 1 || source.SavedCount != 0 || source.FailedCount != 0 {
			t.Errorf("%s: counts saved %d, failed %d, skipped %d", name, source.SavedCount, source.FailedCount, source.SkippedCount)
		}
		if len(source.Errors) != 0 {
			t.Errorf("%s: a skip is not an error: %v", name, source.Errors)
		}
		if len(source.Skipped) != 1 {
			t.Fatalf("%s: skipped = %v, want one entry", name, source.Skipped)
		}
		for _, want := range []string{`event 3 ("Twen")`, "deleted", "link: https://hall.example/twen", "guid: Hall-https://hall.example/twen"} {
			if !strings.Contains(source.Skipped[0], want) {
				t.Errorf("%s: skipped entry lacks %q:\n%s", name, want, source.Skipped[0])
			}
		}
	}
}

func TestRecordEventResultFailedIsAnError(t *testing.T) {
	var source models.IngestRunSource
	got := recordEventResult(&source, 1, ingest.Event{Title: "Bad"}, "Hall", errors.New("boom"))

	if got != eventFailed || source.FailedCount != 1 || source.SavedCount != 0 || source.SkippedCount != 0 {
		t.Errorf("outcome %v, source %+v", got, source)
	}
	if len(source.Errors) != 1 || !strings.Contains(source.Errors[0], "boom") {
		t.Errorf("errors = %v", source.Errors)
	}
	if len(source.Skipped) != 0 {
		t.Errorf("a failure is not a skip: %v", source.Skipped)
	}
}

func TestRecordEventResultMixedRun(t *testing.T) {
	var source models.IngestRunSource
	for i, err := range []error{nil, errEventDeleted, nil, errors.New("boom"), errEventDeleted} {
		recordEventResult(&source, i, ingest.Event{Title: fmt.Sprintf("e%d", i)}, "Hall", err)
	}
	if source.SavedCount != 2 || source.SkippedCount != 2 || source.FailedCount != 1 {
		t.Errorf("counts saved %d, skipped %d, failed %d", source.SavedCount, source.SkippedCount, source.FailedCount)
	}
	if len(source.Skipped) != 2 || len(source.Errors) != 1 {
		t.Errorf("skipped %d entries, errors %d entries", len(source.Skipped), len(source.Errors))
	}
}

func TestSettleStatusTakesTheWorst(t *testing.T) {
	for name, tc := range map[string]struct {
		source models.IngestRunSource
		want   string
	}{
		"nothing wrong": {
			models.IngestRunSource{Status: models.IngestSourceStatusOk, SavedCount: 4},
			models.IngestSourceStatusOk,
		},
		"only skips": {
			models.IngestRunSource{Status: models.IngestSourceStatusOk, SavedCount: 3, SkippedCount: 1},
			models.IngestSourceStatusSkipped,
		},
		"a failed event": {
			models.IngestRunSource{Status: models.IngestSourceStatusOk, SavedCount: 3, FailedCount: 1},
			models.IngestSourceStatusErrors,
		},
		"skips and a failed event": {
			models.IngestRunSource{Status: models.IngestSourceStatusOk, SkippedCount: 2, FailedCount: 1},
			models.IngestSourceStatusErrors,
		},
		"skips after the scraper reported errors": {
			models.IngestRunSource{Status: models.IngestSourceStatusErrors, SkippedCount: 1},
			models.IngestSourceStatusErrors,
		},
		"skips after the venue failed": {
			models.IngestRunSource{Status: models.IngestSourceStatusFailed, SkippedCount: 1},
			models.IngestSourceStatusFailed,
		},
	} {
		source := tc.source
		settleStatus(&source)
		if source.Status != tc.want {
			t.Errorf("%s: status = %q, want %q", name, source.Status, tc.want)
		}
	}
}
