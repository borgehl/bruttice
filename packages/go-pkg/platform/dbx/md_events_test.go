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
