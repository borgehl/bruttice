package util

import (
	"regexp"
	"testing"
)

func TestGetFirstSubmatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		want    string
	}{
		{
			name:    "returns first captured group",
			pattern: `(\d+)`,
			input:   "abc123def",
			want:    "123",
		},
		{
			name:    "returns first group when multiple groups exist",
			pattern: `(\w+)\s+(\w+)`,
			input:   "hello world",
			want:    "hello",
		},
		{
			name:    "returns empty string when no match",
			pattern: `(\d+)`,
			input:   "no numbers here",
			want:    "",
		},
		{
			name:    "returns empty string when pattern has no capture groups",
			pattern: `\d+`,
			input:   "abc123",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(tt.pattern)
			got := GetFirstSubmatch(re, tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
