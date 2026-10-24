package dbx

import (
	"errors"
	"testing"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/configs"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/configx"
	"github.com/google/uuid"
)

// These run against a real database like md_inbox_message_test.go, see testInboxDB.
func testConfigDB(t *testing.T) *DatabaseClient {
	t.Helper()
	d := testInboxDB(t)
	if err := d.DB.AutoMigrate(&models.Age{}, &models.Region{}, &models.Tag{}, &models.Venue{}, &models.Event{}, &models.EventAdminJudgement{}); err != nil {
		t.Fatal(err)
	}
	return d
}

func uniq(prefix string) string { return prefix + "-" + uuid.NewString()[:8] }

// cleanup removes what a test made, whatever it did to it
func cleanup(t *testing.T, d *DatabaseClient, rows ...any) {
	t.Cleanup(func() {
		for _, r := range rows {
			d.DB.Unscoped().Delete(r)
		}
	})
}

func mustRegion(t *testing.T, d *DatabaseClient, name string) models.Region {
	t.Helper()
	r := models.Region{Name: name}
	if err := d.DB.Create(&r).Error; err != nil {
		t.Fatal(err)
	}
	cleanup(t, d, &r)
	return r
}

func mustAge(t *testing.T, d *DatabaseClient) models.Age {
	t.Helper()
	a := models.Age{Name: uniq("age"), Value: 18}
	if err := d.DB.Create(&a).Error; err != nil {
		t.Fatal(err)
	}
	cleanup(t, d, &a)
	return a
}

func mustVenue(t *testing.T, d *DatabaseClient, region string, scraped bool) models.Venue {
	t.Helper()
	v := models.Venue{Name: uniq("venue"), DefaultRegion: region, IsScraped: scraped}
	v.InternalUpdate()
	if err := d.DB.Create(&v).Error; err != nil {
		t.Fatal(err)
	}
	cleanup(t, d, &v)
	return v
}

func mustEvent(t *testing.T, d *DatabaseClient, age models.Age, region models.Region, venue models.Venue, removed bool) models.Event {
	t.Helper()
	e := models.Event{Guid: uuid.NewString(), Title: "Show", StartAt: time.Now().Add(time.Hour), AgeID: age.ID, RegionID: region.ID, VenueID: venue.ID}
	if err := d.DB.Create(&e).Error; err != nil {
		t.Fatal(err)
	}
	if removed {
		if err := d.DB.Delete(&e).Error; err != nil {
			t.Fatal(err)
		}
	}
	// events go first, they refer to the rest
	t.Cleanup(func() { d.DB.Unscoped().Where("id = ?", e.ID).Delete(&models.Event{}) })
	return e
}

func regionOf(t *testing.T, d *DatabaseClient, eventID int64) int64 {
	t.Helper()
	var e models.Event
	if err := d.DB.Unscoped().First(&e, eventID).Error; err != nil {
		t.Fatal(err)
	}
	return e.RegionID
}

func wantKind(t *testing.T, err error, kind ConfigErrorKind) {
	t.Helper()
	var ce *ConfigError
	if !errors.As(err, &ce) || ce.Kind != kind {
		t.Fatalf("error = %v, want kind %d", err, kind)
	}
}

func TestRetireRegionMovesEventsVenuesAndNames(t *testing.T) {
	d := testConfigDB(t)
	a, b, c := mustRegion(t, d, uniq("A")), mustRegion(t, d, uniq("B")), mustRegion(t, d, uniq("C"))
	age := mustAge(t, d)
	venue := mustVenue(t, d, a.Name, false)
	live := mustEvent(t, d, age, a, venue, false)
	gone := mustEvent(t, d, age, a, venue, true)

	if _, err := d.RetireRegion(a.ID, nil, false); err == nil {
		t.Error("a retirement has to say what happens to the events")
	}
	if _, err := d.RetireRegion(a.ID, &a.ID, false); err == nil {
		t.Error("a region can't be moved into itself")
	}
	res, err := d.RetireRegion(a.ID, &b.ID, false)
	if err != nil || res.EventsMoved != 1 {
		t.Fatalf("result = %+v, err = %v", res, err)
	}
	if regionOf(t, d, live.ID) != b.ID || regionOf(t, d, gone.ID) != b.ID {
		t.Error("all events move, the removed ones too")
	}
	var v models.Venue
	d.DB.First(&v, venue.ID)
	if v.DefaultRegion != b.Name {
		t.Errorf("venue default = %q", v.DefaultRegion)
	}
	if got, _ := d.GetRegion(models.Region{Name: a.Name}, d.DB); got.ID != b.ID {
		t.Errorf("the old name must lead to the new region: %+v", got)
	}
	if names, _ := d.GetRegionNames(d.DB); contains(names, a.Name) || !contains(names, b.Name) {
		t.Errorf("names = %v", names)
	}

	// a later retirement carries the older names along
	if _, err := d.RetireRegion(b.ID, &c.ID, false); err != nil {
		t.Fatal(err)
	}
	if got, _ := d.GetRegion(models.Region{Name: a.Name}, d.DB); got.ID != c.ID {
		t.Errorf("the oldest name must follow: %+v", got)
	}
	_, err = d.RetireRegion(a.ID, &c.ID, false)
	wantKind(t, err, ConfigConflict)
}

func TestRetireRegionByDeletingEvents(t *testing.T) {
	d := testConfigDB(t)
	a, b := mustRegion(t, d, uniq("A")), mustRegion(t, d, uniq("B"))
	age := mustAge(t, d)
	venue := mustVenue(t, d, a.Name, false)
	e := mustEvent(t, d, age, a, venue, false)
	keep := mustEvent(t, d, age, b, venue, false)

	_, err := d.RetireRegion(a.ID, nil, true)
	wantKind(t, err, ConfigConflict)
	if err := d.DB.Model(&venue).Update("default_region", b.Name).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := d.RetireRegion(a.ID, &b.ID, true); err == nil {
		t.Error("moving and deleting at once is not a choice")
	}
	res, err := d.RetireRegion(a.ID, nil, true)
	if err != nil || res.EventsDeleted != 1 {
		t.Fatalf("result = %+v, err = %v", res, err)
	}
	var count int64
	d.DB.Model(&models.Event{}).Where("id IN ?", []int64{e.ID, keep.ID}).Count(&count)
	if count != 1 {
		t.Errorf("only the event of the retired region goes: %d left", count)
	}
	if got, _ := d.GetRegion(models.Region{Name: a.Name}, d.DB); got.ID != 0 {
		t.Errorf("a region retired without a replacement is not found: %+v", got)
	}
}

func TestRetireRegionLeavesOneRegion(t *testing.T) {
	d := testConfigDB(t)
	// every other region is retired for the length of the test
	var others []models.Region
	d.DB.Where("retired_at = 0").Find(&others)
	only := mustRegion(t, d, uniq("Only"))
	for _, o := range others {
		d.DB.Model(&models.Region{}).Where("id = ?", o.ID).Update("retired_at", time.Now().Unix())
		o := o
		t.Cleanup(func() { d.DB.Model(&models.Region{}).Where("id = ?", o.ID).Update("retired_at", 0) })
	}
	_, err := d.RetireRegion(only.ID, nil, true)
	wantKind(t, err, ConfigConflict)
}

func TestRenameRegionKeepsTheOldNameWorking(t *testing.T) {
	d := testConfigDB(t)
	old, other := mustRegion(t, d, uniq("Old")), mustRegion(t, d, uniq("Other"))
	age := mustAge(t, d)
	venue := mustVenue(t, d, old.Name, false)
	e := mustEvent(t, d, age, old, venue, false)
	newName := uniq("New")
	t.Cleanup(func() { d.DB.Unscoped().Where("name = ?", newName).Delete(&models.Region{}) })

	wantKind(t, d.RenameRegion(old.ID, other.Name), ConfigConflict)
	wantKind(t, d.RenameRegion(old.ID, "  "), ConfigInvalid)
	if err := d.RenameRegion(old.ID, newName); err != nil {
		t.Fatal(err)
	}
	renamed, _ := d.GetRegion(models.Region{Name: newName}, d.DB)
	if renamed.ID == 0 || regionOf(t, d, e.ID) != renamed.ID {
		t.Errorf("the event follows the new name: %+v", renamed)
	}
	if got, _ := d.GetRegion(models.Region{Name: old.Name}, d.DB); got.ID != renamed.ID {
		t.Errorf("the old name must lead to the new: %+v", got)
	}
	var v models.Venue
	d.DB.First(&v, venue.ID)
	if v.DefaultRegion != newName {
		t.Errorf("venue default = %q", v.DefaultRegion)
	}

	// naming a region again brings the retired one back
	if err := d.CreateRegion(old.Name); err != nil {
		t.Fatal(err)
	}
	if got, _ := d.GetRegion(models.Region{Name: old.Name}, d.DB); got.ID != old.ID || got.Retired() {
		t.Errorf("the retired region must be restored: %+v", got)
	}
	wantKind(t, d.CreateRegion(old.Name), ConfigConflict)
	wantKind(t, d.CreateRegion(" "), ConfigInvalid)
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func TestTags(t *testing.T) {
	d := testConfigDB(t)
	name := uniq("tag")
	if err := d.CreateTag(name); err != nil {
		t.Fatal(err)
	}
	var tag models.Tag
	d.DB.Where("name = ?", name).First(&tag)
	cleanup(t, d, &tag)
	wantKind(t, d.CreateTag(name), ConfigConflict)

	region := mustRegion(t, d, uniq("R"))
	age, venue := mustAge(t, d), mustVenue(t, d, region.Name, false)
	e := mustEvent(t, d, age, region, venue, false)
	if err := d.DB.Model(&e).Association("Tags").Append(&tag); err != nil {
		t.Fatal(err)
	}
	rows, _ := d.ListTagRows()
	for _, r := range rows {
		if r.ID == tag.ID && r.Events != 1 {
			t.Errorf("events = %d", r.Events)
		}
	}

	renamed := uniq("tag")
	if err := d.RenameTag(tag.ID, renamed); err != nil {
		t.Fatal(err)
	}
	if err := d.DeleteTag(tag.ID); err != nil {
		t.Fatal(err)
	}
	var left int64
	d.DB.Table("event_tags").Where("tag_id = ?", tag.ID).Count(&left)
	d.DB.Model(&models.Event{}).Where("id = ?", e.ID).Count(&left)
	if left != 1 {
		t.Error("deleting a tag leaves the event")
	}
	if err := d.CreateTag(renamed); err != nil {
		t.Errorf("the name of a deleted tag can be used again: %v", err)
	}
	d.DB.Unscoped().Where("name = ?", renamed).Delete(&models.Tag{})

	free := models.Tag{Name: configs.FreeTag}
	d.DB.Where("name = ?", free.Name).FirstOrCreate(&free)
	wantKind(t, d.RenameTag(free.ID, "gratis"), ConfigConflict)
	wantKind(t, d.DeleteTag(free.ID), ConfigConflict)
	wantKind(t, d.DeleteTag(-1), ConfigNotFound)
}

func TestAges(t *testing.T) {
	d := testConfigDB(t)
	region := mustRegion(t, d, uniq("R"))
	a, b := mustAge(t, d), mustAge(t, d)
	venue := mustVenue(t, d, region.Name, false)
	e := mustEvent(t, d, a, region, venue, false)
	gone := mustEvent(t, d, a, region, venue, true)

	wantKind(t, d.UpdateAge(a.ID, b.Name, 21), ConfigConflict)
	wantKind(t, d.UpdateAge(a.ID, "x", 256), ConfigInvalid)
	name := uniq("age")
	if err := d.UpdateAge(a.ID, name, 21); err != nil {
		t.Fatal(err)
	}
	var got models.Age
	d.DB.First(&got, a.ID)
	if got.Name != name || got.Value != 21 {
		t.Errorf("age = %+v", got)
	}

	_, err := d.DeleteAge(a.ID, nil)
	wantKind(t, err, ConfigInvalid)
	_, err = d.DeleteAge(a.ID, &a.ID)
	wantKind(t, err, ConfigInvalid)
	moved, err := d.DeleteAge(a.ID, &b.ID)
	if err != nil || moved != 2 {
		t.Fatalf("moved = %d, err = %v", moved, err)
	}
	for _, id := range []int64{e.ID, gone.ID} {
		var ev models.Event
		d.DB.Unscoped().First(&ev, id)
		if ev.AgeID != b.ID {
			t.Errorf("event %d age = %d", id, ev.AgeID)
		}
	}
	// no events, no replacement needed
	if _, err := d.DeleteAge(b.ID, nil); err == nil {
		t.Error("b has events now")
	}
	unused := mustAge(t, d)
	if _, err := d.DeleteAge(unused.ID, nil); err != nil {
		t.Errorf("an unused age group needs no replacement: %v", err)
	}
}

func TestVenues(t *testing.T) {
	d := testConfigDB(t)
	region := mustRegion(t, d, uniq("R"))
	retired := mustRegion(t, d, uniq("Retired"))
	d.DB.Model(&retired).Update("retired_at", time.Now().Unix())

	name := uniq("Venue")
	wantKind(t, d.CreateVenue(name, "not a link", region.Name), ConfigInvalid)
	wantKind(t, d.CreateVenue(name, "javascript:alert(1)", region.Name), ConfigInvalid)
	wantKind(t, d.CreateVenue(name, "", retired.Name), ConfigInvalid)
	wantKind(t, d.CreateVenue(name, "", "nowhere"), ConfigInvalid)
	if err := d.CreateVenue(name, "https://example.com", region.Name); err != nil {
		t.Fatal(err)
	}
	var venue models.Venue
	d.DB.Where("name = ?", name).First(&venue)
	cleanup(t, d, &venue)
	wantKind(t, d.CreateVenue(name, "", region.Name), ConfigConflict)

	renamed := uniq("Venue")
	if err := d.UpdateVenue(venue.ID, renamed, "", region.Name); err != nil {
		t.Fatal(err)
	}
	d.DB.First(&venue, venue.ID)
	if venue.Name != renamed || venue.Slug == "" || venue.Link != "" {
		t.Errorf("venue = %+v", venue)
	}

	scraped := mustVenue(t, d, region.Name, true)
	wantKind(t, d.UpdateVenue(scraped.ID, "x", "", region.Name), ConfigConflict)
	wantKind(t, d.DeleteVenue(scraped.ID), ConfigConflict)

	age := mustAge(t, d)
	mustEvent(t, d, age, region, venue, true)
	wantKind(t, d.DeleteVenue(venue.ID), ConfigConflict)
	empty := mustVenue(t, d, region.Name, false)
	if err := d.DeleteVenue(empty.ID); err != nil {
		t.Errorf("a venue without events can go: %v", err)
	}
}

func TestListRowsCount(t *testing.T) {
	d := testConfigDB(t)
	a, b := mustRegion(t, d, uniq("A")), mustRegion(t, d, uniq("B"))
	age := mustAge(t, d)
	venue := mustVenue(t, d, a.Name, false)
	mustEvent(t, d, age, a, venue, false)
	mustEvent(t, d, age, a, venue, true)
	if _, err := d.RetireRegion(b.ID, &a.ID, false); err != nil {
		t.Fatal(err)
	}
	rows, err := d.ListRegionRows()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		switch r.ID {
		case a.ID:
			if r.Events != 1 || r.Venues != 1 || r.RetiredAt != 0 {
				t.Errorf("a = %+v", r)
			}
		case b.ID:
			if r.RetiredAt == 0 || r.MergedInto != a.Name {
				t.Errorf("b = %+v", r)
			}
		}
	}
	for i := range rows {
		if rows[i].RetiredAt > 0 {
			for _, later := range rows[i:] {
				if later.RetiredAt == 0 {
					t.Fatal("retired regions come last")
				}
			}
		}
	}
}

func TestConfigListsAreOnlySeededOnce(t *testing.T) {
	d := testConfigDB(t)
	if err := d.DB.AutoMigrate(&models.MigrationFlag{}); err != nil {
		t.Fatal(err)
	}
	d.DB.Unscoped().Where("name = ?", seedListsKey).Delete(&models.MigrationFlag{})
	t.Cleanup(func() { d.DB.Unscoped().Where("name = ?", seedListsKey).Delete(&models.MigrationFlag{}) })

	name, tag := uniq("Seeded"), uniq("seedtag")
	cfg := configx.Config{InitialData: configx.InitialData{
		Regions: []configx.ConfigRawRegion{{Name: name}},
		Tags:    []string{tag},
	}}
	t.Cleanup(func() {
		d.DB.Unscoped().Where("name = ?", name).Delete(&models.Region{})
		d.DB.Unscoped().Where("name = ?", tag).Delete(&models.Tag{})
	})

	if err := d.seedLists(cfg, d.DB); err != nil {
		t.Fatal(err)
	}
	var region models.Region
	if err := d.DB.Where("name = ?", name).First(&region).Error; err != nil {
		t.Fatalf("the first start saves the lists: %v", err)
	}

	// an admin retires the region and deletes the tag, a restart must not bring them back
	d.DB.Model(&region).Update("retired_at", time.Now().Unix())
	d.DB.Unscoped().Where("name = ?", tag).Delete(&models.Tag{})
	if err := d.seedLists(cfg, d.DB); err != nil {
		t.Fatal(err)
	}
	var regions, tags int64
	d.DB.Unscoped().Model(&models.Region{}).Where("name = ? AND retired_at = 0", name).Count(&regions)
	d.DB.Model(&models.Tag{}).Where("name = ?", tag).Count(&tags)
	if regions != 0 || tags != 0 {
		t.Errorf("seeded again: regions = %d, tags = %d", regions, tags)
	}
}
