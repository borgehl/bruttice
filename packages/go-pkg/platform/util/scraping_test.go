package util

import (
	"testing"
)

func TestTagClassSelector(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		classes string
		want    string
	}{
		{
			name:    "single class",
			tag:     "a",
			classes: "single",
			want:    "a.single",
		},
		{
			name:    "multiple classes become dot-separated",
			tag:     "div",
			classes: "foo bar",
			want:    "div.foo.bar",
		},
		{
			name:    "three classes",
			tag:     "span",
			classes: "one two three",
			want:    "span.one.two.three",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TagClassSelector(tt.tag, tt.classes)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAttributeSelector(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		attribute string
		value     string
		want      string
	}{
		{
			name:      "with value",
			tag:       "a",
			attribute: "href",
			value:     "https://example.com",
			want:      "a[href='https://example.com']",
		},
		{
			name:      "empty value omits value from selector",
			tag:       "a",
			attribute: "href",
			value:     "",
			want:      "a[href]",
		},
		{
			name:      "non-anchor tag",
			tag:       "input",
			attribute: "type",
			value:     "text",
			want:      "input[type='text']",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AttributeSelector(tt.tag, tt.attribute, tt.value)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsFree(t *testing.T) {
	free := []string{
		"$0",
		"0",
		"Free",
		"free",
		"FREE",
		"free admission",
		"donations",
		"Donations welcome",
		"no cover",
		"No Cover Charge",
		"pwyc",
		"PWYC",
		"notaflof",
		"no one turned away",
	}

	notFree := []string{
		"$10",
		"$5.00",
		"15",
		"$20 advance",
		"tickets required",
	}

	for _, cost := range free {
		t.Run("free: "+cost, func(t *testing.T) {
			if !IsFree(cost) {
				t.Errorf("IsFree(%q) = false, want true", cost)
			}
		})
	}

	for _, cost := range notFree {
		t.Run("not free: "+cost, func(t *testing.T) {
			if IsFree(cost) {
				t.Errorf("IsFree(%q) = true, want false", cost)
			}
		})
	}
}
