package util

import (
	"testing"
)

func TestTrimCleanStr(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain string unchanged",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "trims leading and trailing whitespace",
			input: "  hello  ",
			want:  "hello",
		},
		{
			name:  "unicode whitespace normalized to space",
			input: "hello\u00a0world",
			want:  "hello world",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrimCleanStr(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSnakeCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "spaces become underscores",
			input: "hello world",
			want:  "hello_world",
		},
		{
			name:  "trims and converts",
			input: "  foo bar  ",
			want:  "foo_bar",
		},
		{
			name:  "single word",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCleanUTF8String(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain ASCII unchanged",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "non-breaking space (U+00A0) replaced with regular space",
			input: "hello\u00a0world",
			want:  "hello world",
		},
		{
			name:  "em space (U+2003) replaced with regular space",
			input: "hello\u2003world",
			want:  "hello world",
		},
		{
			name:  "thin space (U+2009) replaced with regular space",
			input: "hello\u2009world",
			want:  "hello world",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanUTF8String(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCleanoutHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "removes span tags",
			input: "<span>hello</span>",
			want:  "hello",
		},
		{
			name:  "replaces br with /n",
			input: "line1<br>line2",
			want:  "line1/nline2",
		},
		{
			name:  "handles all three replacements",
			input: "<span>hello</span><br><span>world</span>",
			want:  "hello/nworld",
		},
		{
			name:  "no HTML unchanged",
			input: "plain text",
			want:  "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanoutHTML(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
