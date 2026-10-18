package util

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestBase64Encode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			// "hello" in standard base64 is "aGVsbG8=" — padding stripped
			name:  "basic encoding strips padding",
			input: []byte("hello"),
			want:  "aGVsbG8",
		},
		{
			// 0xfb produces '+' in standard base64 — replaced with '_'
			name:  "replaces + with _",
			input: []byte{0xfb},
			want:  "_w",
		},
		{
			// 0xfc produces '/' in standard base64 — replaced with '-'
			name:  "replaces / with -",
			input: []byte{0xfc},
			want:  "-A",
		},
		{
			name:  "empty input",
			input: []byte{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Base64Encode(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
			if strings.ContainsAny(got, "+/=") {
				t.Errorf("output contains non-URL-safe characters: %q", got)
			}
		})
	}
}

func TestBase64EncodeString(t *testing.T) {
	got := Base64EncodeString("hello")
	want := "aGVsbG8"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBase64DecodeString(t *testing.T) {
	input := []byte("hello world")
	encoded := base64.StdEncoding.EncodeToString(input)
	got := Base64DecodeString(encoded)
	if string(got) != string(input) {
		t.Errorf("got %q, want %q", string(got), string(input))
	}
}

func TestBase64DecodeStringPanicsOnInvalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid base64 input, got none")
		}
	}()
	Base64DecodeString("not valid base64!!!")
}

func TestSha256Base64String(t *testing.T) {
	t.Run("is deterministic", func(t *testing.T) {
		a := Sha256Base64String("hello")
		b := Sha256Base64String("hello")
		if a != b {
			t.Error("same input produced different outputs")
		}
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		a := Sha256Base64String("hello")
		b := Sha256Base64String("world")
		if a == b {
			t.Error("different inputs produced the same hash")
		}
	})

	t.Run("output is URL-safe with no padding", func(t *testing.T) {
		got := Sha256Base64String("hello")
		if strings.ContainsAny(got, "+/=") {
			t.Errorf("output contains non-URL-safe characters: %q", got)
		}
	})
}

func TestSha256Base64Bytes(t *testing.T) {
	str := "hello"
	fromStr := Sha256Base64String(str)
	fromBytes := Sha256Base64Bytes([]byte(str))
	if fromStr != fromBytes {
		t.Errorf("Sha256Base64String and Sha256Base64Bytes differ: %q vs %q", fromStr, fromBytes)
	}
}
