package util

import (
	"testing"
	"time"
)

func TestParseDateTime(t *testing.T) {
	today := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		input     string
		format    string
		wantYear  int
		wantMonth time.Month
		wantDay   int
		wantErr   bool
	}{
		{
			name:      "future date returned as-is",
			input:     "2024-12-01T10:00:00Z",
			format:    time.RFC3339,
			wantYear:  2024,
			wantMonth: time.December,
			wantDay:   1,
		},
		{
			name:      "past date gets one year added",
			input:     "2024-01-01T10:00:00Z",
			format:    time.RFC3339,
			wantYear:  2025,
			wantMonth: time.January,
			wantDay:   1,
		},
		{
			// condition is strictly <, so equal to today is not rolled forward
			name:      "exactly today is not rolled forward",
			input:     "2024-06-15T00:00:00Z",
			format:    time.RFC3339,
			wantYear:  2024,
			wantMonth: time.June,
			wantDay:   15,
		},
		{
			name:    "invalid input returns error",
			input:   "not-a-date",
			format:  time.RFC3339,
			wantErr: true,
		},
		{
			name:    "wrong format returns error",
			input:   "2024-12-01T10:00:00Z",
			format:  "2006-01-02",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDateTime(tt.input, tt.format, today)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Year() != tt.wantYear || got.Month() != tt.wantMonth || got.Day() != tt.wantDay {
				t.Errorf("got %v, want %d-%02d-%02d", got, tt.wantYear, tt.wantMonth, tt.wantDay)
			}
		})
	}
}

func TestMultipleDatetimeParsingTries(t *testing.T) {
	today := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	t.Run("uses first matching format", func(t *testing.T) {
		got, err := MultipleDatetimeParsingTries(
			"2024-12-01T10:00:00Z",
			[]string{"2006-01-02", time.RFC3339},
			today,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Month() != time.December || got.Day() != 1 {
			t.Errorf("unexpected result: %v", got)
		}
	})

	t.Run("returns error when no format matches", func(t *testing.T) {
		_, err := MultipleDatetimeParsingTries(
			"not-a-date",
			[]string{"2006-01-02", time.RFC3339},
			today,
		)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("empty formats list returns error", func(t *testing.T) {
		_, err := MultipleDatetimeParsingTries("2024-12-01T10:00:00Z", []string{}, today)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestNowDate(t *testing.T) {
	central, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Fatal(err)
	}
	prev := TZLocation
	SetTZLocation(central)
	t.Cleanup(func() { SetTZLocation(prev) })

	got := NowDate()

	// NowDate() returns the UTC instant of midnight Central time.
	// That is 5 or 6 hours after midnight UTC, so the result must NOT be at UTC midnight.
	if got.Hour() == 0 && got.Minute() == 0 && got.Second() == 0 {
		t.Errorf("NowDate() returned midnight UTC (%v); expected midnight Central expressed in UTC (e.g. 05:00 or 06:00 UTC)", got)
	}

	// The result must be in UTC.
	if got.Location() != time.UTC {
		t.Errorf("NowDate() location = %v, want UTC", got.Location())
	}

	// The Central-local calendar date must match today.
	now := time.Now().In(TZLocation)
	if got.In(TZLocation).Day() != now.Day() || got.In(TZLocation).Month() != now.Month() || got.In(TZLocation).Year() != now.Year() {
		t.Errorf("NowDate() local date = %v, want %v", got.In(TZLocation).Format("2006-01-02"), now.Format("2006-01-02"))
	}
}

func TestDisplayDateFormat(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
		want  string
	}{
		{
			name:  "January gets period",
			input: time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC),
			want:  "Monday Jan. 15, 2024",
		},
		{
			name:  "May has no period",
			input: time.Date(2024, time.May, 1, 0, 0, 0, 0, time.UTC),
			want:  "Wednesday May 1, 2024",
		},
		{
			name:  "July gets period",
			input: time.Date(2024, time.July, 4, 0, 0, 0, 0, time.UTC),
			want:  "Thursday Jul. 4, 2024",
		},
		{
			name:  "September gets period",
			input: time.Date(2024, time.September, 3, 0, 0, 0, 0, time.UTC),
			want:  "Tuesday Sep. 3, 2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DisplayDateFormat(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
