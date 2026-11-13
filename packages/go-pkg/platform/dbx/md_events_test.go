package dbx

import (
	"testing"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

// These run against a real database like md_config_test.go, see testInboxDB.

func eventByID(t *testing.T, d *DatabaseClient, id int64) models.Event {
	t.Helper()
	var e models.Event
	if err := d.DB.Unscoped().First(&e, id).Error; err != nil {
		t.Fatal(err)
	}
	return e
}

func TestMarkEventCancelledOnlySetsCancelledAt(t *testing.T) {
	d := testConfigDB(t)
	region := mustRegion(t, d, uniq("region"))
	e := mustEvent(t, d, mustAge(t, d), region, mustVenue(t, d, region.Name, true), false)

	marked, err := d.MarkEventCancelled(e.Guid, 1_000, d.DB)
	if err != nil || !marked {
		t.Fatalf("marked %v, err %v", marked, err)
	}

	got := eventByID(t, d, e.ID)
	if got.CancelledAt == nil || *got.CancelledAt != 1_000 {
		t.Errorf("cancelled_at = %v", got.CancelledAt)
	}
	if got.DeletedAt.Valid {
		t.Error("a cancelled event must stay visible")
	}
	if got.Title != e.Title || !got.StartAt.Equal(e.StartAt) || got.AgeID != e.AgeID || got.RegionID != e.RegionID || got.VenueID != e.VenueID {
		t.Errorf("other fields changed: %+v", got)
	}
}

func TestMarkEventCancelledKeepsTheFirstTime(t *testing.T) {
	d := testConfigDB(t)
	region := mustRegion(t, d, uniq("region"))
	e := mustEvent(t, d, mustAge(t, d), region, mustVenue(t, d, region.Name, true), false)

	if _, err := d.MarkEventCancelled(e.Guid, 1_000, d.DB); err != nil {
		t.Fatal(err)
	}
	marked, err := d.MarkEventCancelled(e.Guid, 2_000, d.DB)
	if err != nil || marked {
		t.Fatalf("second mark: marked %v, err %v", marked, err)
	}
	if got := eventByID(t, d, e.ID); got.CancelledAt == nil || *got.CancelledAt != 1_000 {
		t.Errorf("cancelled_at = %v, want the first time", got.CancelledAt)
	}
}

func TestMarkEventCancelledIgnoresUnknownAndRemovedEvents(t *testing.T) {
	d := testConfigDB(t)
	region := mustRegion(t, d, uniq("region"))
	removed := mustEvent(t, d, mustAge(t, d), region, mustVenue(t, d, region.Name, true), true)

	if marked, err := d.MarkEventCancelled("no-such-guid", 1_000, d.DB); err != nil || marked {
		t.Errorf("unknown guid: marked %v, err %v", marked, err)
	}
	if marked, err := d.MarkEventCancelled(removed.Guid, 1_000, d.DB); err != nil || marked {
		t.Errorf("removed event: marked %v, err %v", marked, err)
	}
	if got := eventByID(t, d, removed.ID); got.CancelledAt != nil {
		t.Errorf("a removed event was marked: %v", got.CancelledAt)
	}
}

// The ingest handler skips an event whose guid is found deleted here, see deletedEventError.
func TestGetEventByGuidUnscopedFindsDeletedEvents(t *testing.T) {
	d := testConfigDB(t)
	region := mustRegion(t, d, uniq("region"))
	age := mustAge(t, d)
	venue := mustVenue(t, d, region.Name, true)
	live := mustEvent(t, d, age, region, venue, false)
	removed := mustEvent(t, d, age, region, venue, true)

	got, err := d.GetEventByGuidUnscoped(removed.Guid, d.DB)
	if err != nil || got.ID != removed.ID || !got.DeletedAt.Valid {
		t.Errorf("deleted event: id %d (want %d), deleted %v, err %v", got.ID, removed.ID, got.DeletedAt.Valid, err)
	}
	got, err = d.GetEventByGuidUnscoped(live.Guid, d.DB)
	if err != nil || got.ID != live.ID || got.DeletedAt.Valid {
		t.Errorf("live event: id %d (want %d), deleted %v, err %v", got.ID, live.ID, got.DeletedAt.Valid, err)
	}
	got, err = d.GetEventByGuidUnscoped("no-such-guid", d.DB)
	if err != nil || got.ID != 0 {
		t.Errorf("unknown guid: id %d, err %v", got.ID, err)
	}
}

func TestSaveEventSetsAndClearsCancelledAt(t *testing.T) {
	d := testConfigDB(t)
	region := mustRegion(t, d, uniq("region"))
	age := mustAge(t, d)
	venue := mustVenue(t, d, region.Name, true)
	e := mustEvent(t, d, age, region, venue, false)

	save := func(cancelledAt *int64) models.Event {
		t.Helper()
		_, err := d.SaveEvent(models.Event{
			Guid: e.Guid, Title: "Renamed", StartAt: time.Now().Add(2 * time.Hour), CancelledAt: cancelledAt,
			Age: models.Age{Name: age.Name}, Region: models.Region{Name: region.Name}, Venue: models.Venue{Name: venue.Name},
		}, true, d.DB)
		if err != nil {
			t.Fatal(err)
		}
		return eventByID(t, d, e.ID)
	}

	at := int64(1_000)
	if got := save(&at); got.CancelledAt == nil || *got.CancelledAt != 1_000 || got.Title != "Renamed" {
		t.Errorf("cancelled save: cancelled_at %v, title %q", got.CancelledAt, got.Title)
	}
	if got := save(nil); got.CancelledAt != nil {
		t.Errorf("a save that is no longer cancelled left cancelled_at = %d", *got.CancelledAt)
	}
}

func TestDeduplicateEventsLeavesCancelledEventsAlone(t *testing.T) {
	d := testConfigDB(t)
	region := mustRegion(t, d, uniq("region"))
	age := mustAge(t, d)
	venue := mustVenue(t, d, region.Name, true)

	// same venue, same hour: a pair this close is a duplicate, one of them is removed
	live := mustEvent(t, d, age, region, venue, false)
	sameTime := mustEvent(t, d, age, region, venue, false)
	if err := d.DB.Model(&models.Event{}).Where("id = ?", sameTime.ID).Update("start_at", live.StartAt).Error; err != nil {
		t.Fatal(err)
	}
	if err := d.DeduplicateEvents([]int64{venue.ID}, d.DB); err != nil {
		t.Fatal(err)
	}
	if !eventByID(t, d, live.ID).DeletedAt.Valid {
		t.Fatal("setup: the duplicate pair was not deduplicated")
	}

	// the same pair, one of them cancelled, is left alone
	region2 := mustRegion(t, d, uniq("region"))
	venue2 := mustVenue(t, d, region2.Name, true)
	a := mustEvent(t, d, age, region2, venue2, false)
	b := mustEvent(t, d, age, region2, venue2, false)
	if err := d.DB.Model(&models.Event{}).Where("id = ?", b.ID).Update("start_at", a.StartAt).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := d.MarkEventCancelled(a.Guid, 1_000, d.DB); err != nil {
		t.Fatal(err)
	}
	if err := d.DeduplicateEvents([]int64{venue2.ID}, d.DB); err != nil {
		t.Fatal(err)
	}
	for _, e := range []models.Event{a, b} {
		if eventByID(t, d, e.ID).DeletedAt.Valid {
			t.Errorf("event %d was deduplicated next to a cancelled one", e.ID)
		}
	}
}

func TestGetIndexEventsFiltersByRegion(t *testing.T) {
	d := testConfigDB(t)
	tulsa := mustRegion(t, d, uniq("tulsa"))
	okc := mustRegion(t, d, uniq("okc"))
	age := mustAge(t, d)
	inTulsa := mustEvent(t, d, age, tulsa, mustVenue(t, d, tulsa.Name, true), false)
	inOKC := mustEvent(t, d, age, okc, mustVenue(t, d, okc.Name, true), false)

	// only approved events are on the calendar
	for _, e := range []models.Event{inTulsa, inOKC} {
		if err := d.DB.Model(&models.Event{}).Where("id = ?", e.ID).Update("approved_at", time.Now().Unix()).Error; err != nil {
			t.Fatal(err)
		}
	}

	// the events start an hour from now
	shown := func(regions ...string) map[int64]bool {
		t.Helper()
		filter := models.CalendarFilterBackend{
			FromDate: time.Now(),
			ToDate:   time.Now().Add(24 * time.Hour),
			Regions:  regions,
		}
		events, err := d.GetIndexEvents(filter, d.DB)
		if err != nil {
			t.Fatal(err)
		}
		ids := make(map[int64]bool, len(events))
		for _, e := range events {
			ids[e.ID] = true
		}
		return ids
	}

	tests := []struct {
		name    string
		regions []string
		tulsa   bool
		okc     bool
	}{
		{"no regions shows every region", nil, true, true},
		{"one region", []string{tulsa.Name}, true, false},
		{"the other region", []string{okc.Name}, false, true},
		{"two regions", []string{tulsa.Name, okc.Name}, true, true},
		{"a name that is no region shows nothing", []string{uniq("no-such-region")}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shown(tt.regions...)
			if got[inTulsa.ID] != tt.tulsa || got[inOKC.ID] != tt.okc {
				t.Errorf("regions %q: tulsa event shown %v, okc event shown %v; want %v and %v",
					tt.regions, got[inTulsa.ID], got[inOKC.ID], tt.tulsa, tt.okc)
			}
		})
	}

	// a filter's own regions and the request's must both match
	both := func(saved []string, requested []string) map[int64]bool {
		t.Helper()
		filter := models.CalendarFilterBackend{
			FromDate:       time.Now(),
			ToDate:         time.Now().Add(24 * time.Hour),
			Regions:        saved,
			RequestRegions: requested,
		}
		events, err := d.GetIndexEvents(filter, d.DB)
		if err != nil {
			t.Fatal(err)
		}
		ids := make(map[int64]bool, len(events))
		for _, e := range events {
			ids[e.ID] = true
		}
		return ids
	}

	t.Run("saved and requested regions are intersected", func(t *testing.T) {
		tests := []struct {
			name             string
			saved, requested []string
			tulsa, okc       bool
		}{
			{"saved only", []string{tulsa.Name}, nil, true, false},
			{"requested only", nil, []string{okc.Name}, false, true},
			{"requested inside saved", []string{tulsa.Name, okc.Name}, []string{okc.Name}, false, true},
			{"nothing in common shows nothing", []string{tulsa.Name}, []string{okc.Name}, false, false},
		}
		for _, tt := range tests {
			got := both(tt.saved, tt.requested)
			if got[inTulsa.ID] != tt.tulsa || got[inOKC.ID] != tt.okc {
				t.Errorf("%s: tulsa event shown %v, okc event shown %v; want %v and %v",
					tt.name, got[inTulsa.ID], got[inOKC.ID], tt.tulsa, tt.okc)
			}
		}
	})

	t.Run("a retired region is not matched", func(t *testing.T) {
		if err := d.DB.Model(&models.Region{}).Where("id = ?", okc.ID).Update("retired_at", time.Now().Unix()).Error; err != nil {
			t.Fatal(err)
		}
		if got := shown(okc.Name); got[inOKC.ID] {
			t.Error("an event of a retired region was shown for its name")
		}
	})
}
