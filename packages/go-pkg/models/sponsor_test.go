package models

import "testing"

func TestSponsorActiveOn(t *testing.T) {
	cases := []struct {
		name    string
		sponsor Sponsor
		day     string
		want    bool
	}{
		{"enabled with no dates", Sponsor{Enabled: true}, "2026-11-10", true},
		{"disabled", Sponsor{Enabled: false}, "2026-11-10", false},
		{"before the first day", Sponsor{Enabled: true, StartsOn: "2026-11-11"}, "2026-11-10", false},
		{"on the first day", Sponsor{Enabled: true, StartsOn: "2026-11-10"}, "2026-11-10", true},
		{"on the last day", Sponsor{Enabled: true, EndsOn: "2026-11-10"}, "2026-11-10", true},
		{"after the last day", Sponsor{Enabled: true, EndsOn: "2026-11-09"}, "2026-11-10", false},
		{"inside both", Sponsor{Enabled: true, StartsOn: "2026-11-01", EndsOn: "2026-11-30"}, "2026-11-10", true},
		{"disabled inside both", Sponsor{StartsOn: "2026-11-01", EndsOn: "2026-11-30"}, "2026-11-10", false},
	}
	for _, c := range cases {
		if got := c.sponsor.ActiveOn(c.day); got != c.want {
			t.Errorf("%s: ActiveOn(%s) = %v, want %v", c.name, c.day, got, c.want)
		}
	}
}
