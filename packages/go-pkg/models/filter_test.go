package models

import (
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"testing"
	"time"
)

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
