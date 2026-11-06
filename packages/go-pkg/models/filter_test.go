package models

import (
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"testing"
	"time"
)

func TestFilterKeyAndRegions(t *testing.T) {
	plain := CalendarFilterBackend{Tags: []string{"music"}, TagsInclude: true}

	// the regions of a request are not part of the filter
	requested := plain
	requested.RequestRegions = []string{"okc"}
	if plain.GetDBFilter().HashKey != requested.GetDBFilter().HashKey {
		t.Error("the regions of a request changed the key of the filter")
	}

	// the regions of a filter are, and they survive being stored
	saved := plain
	saved.Regions = []string{"okc"}
	if plain.GetDBFilter().HashKey == saved.GetDBFilter().HashKey {
		t.Error("the regions of a filter did not change its key")
	}
	stored := saved.GetDBFilter()
	got := stored.GetBackendFilter()
	if len(got.Regions) != 1 || got.Regions[0] != "okc" {
		t.Errorf("stored regions = %v, want [okc]", got.Regions)
	}
}

func TestCustomOffsetsToDatesAcrossDST(t *testing.T) {
	today := util.NowDate()

	// spans at least one DST transition in either direction
	for _, days := range []int{-400, -200, -100, -30, -1, 0, 1, 30, 100, 200, 400} {
		want := today.AddDate(0, 0, days).In(util.TZLocation)
		str := want.Format("2006-01-02")

		var f CalendarFilterBackend
		if err := f.CustomOffsetsToDates(str, str); err != nil {
			t.Fatal(err)
		}

		if got := f.FromDate.In(util.TZLocation).Format("2006-01-02"); got != str {
			t.Errorf("from %+d days: FromDate = %s, want %s", days, got, str)
		}
		if got := f.ToDate.In(util.TZLocation).Format("2006-01-02"); got != str {
			t.Errorf("to %+d days: ToDate = %s, want %s", days, got, str)
		}
		if f.ToDate.Before(f.FromDate) || f.ToDate.Sub(f.FromDate) >= 24*time.Hour+time.Hour {
			t.Errorf("%+d days: single-day window is %v long", days, f.ToDate.Sub(f.FromDate))
		}
	}
}
