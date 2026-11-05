package api

import "testing"

func TestKeepCancelledAt(t *testing.T) {
	first, later := int64(100), int64(200)
	for name, tc := range map[string]struct {
		incoming, existing, want *int64
	}{
		"newly cancelled":                  {&later, nil, &later},
		"still cancelled keeps first time": {&later, &first, &first},
		"no longer cancelled is cleared":   {nil, &first, nil},
		"never cancelled":                  {nil, nil, nil},
	} {
		got := keepCancelledAt(tc.incoming, tc.existing)
		switch {
		case tc.want == nil && got != nil:
			t.Errorf("%s: got %d, want nil", name, *got)
		case tc.want != nil && (got == nil || *got != *tc.want):
			t.Errorf("%s: got %v, want %d", name, got, *tc.want)
		}
	}
}
