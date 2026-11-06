package api

import (
	"fmt"
	"slices"
	"testing"
)

func TestRegionsFromQuery(t *testing.T) {
	many := make([]string, 0, maxRegionsInQuery+10)
	for i := 0; i < maxRegionsInQuery+10; i++ {
		many = append(many, fmt.Sprintf("region-%d", i))
	}

	tests := []struct {
		name   string
		values []string
		want   []string
	}{
		{"none", nil, nil},
		{"one", []string{"tulsa"}, []string{"tulsa"}},
		{"several keep their order", []string{"okc", "tulsa"}, []string{"okc", "tulsa"}},
		{"space is trimmed", []string{" okc ", "\ttulsa"}, []string{"okc", "tulsa"}},
		{"empty values are dropped", []string{"", "  ", "okc"}, []string{"okc"}},
		{"repeats are dropped", []string{"okc", "tulsa", "okc"}, []string{"okc", "tulsa"}},
		{"a comma is part of a name", []string{"Dallas, TX"}, []string{"Dallas, TX"}},
		{"the number of regions is capped", many, many[:maxRegionsInQuery]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := regionsFromQuery(tt.values); !slices.Equal(got, tt.want) {
				t.Errorf("regionsFromQuery(%q) = %q, want %q", tt.values, got, tt.want)
			}
		})
	}
}
