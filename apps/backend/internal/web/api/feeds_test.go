package api

import (
	"strings"
	"testing"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

func TestFeedEventTitleMarksCancelledEvents(t *testing.T) {
	event := models.Event{Title: "Open Mic", StartAt: time.Date(2030, 1, 10, 19, 0, 0, 0, time.UTC)}
	if got := feedEventTitle(event); strings.Contains(got, "CANCELLED") || !strings.HasSuffix(got, ": Open Mic") {
		t.Errorf("title = %q", got)
	}

	at := int64(1_000)
	event.CancelledAt = &at
	if got := feedEventTitle(event); !strings.HasSuffix(got, ": CANCELLED: Open Mic") {
		t.Errorf("cancelled title = %q", got)
	}
}
