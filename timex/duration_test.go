package timex

import (
	"testing"
	"time"
)

func TestHumanDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30 seconds"},
		{1 * time.Minute, "1 minute"},
		{2 * time.Minute, "2 minutes"},
		{1 * time.Hour, "1 hour"},
		{2 * time.Hour, "2 hours"},
		{24 * time.Hour, "1 day"},
		{48 * time.Hour, "2 days"},
		{7 * 24 * time.Hour, "1 week"},
		{14 * 24 * time.Hour, "2 weeks"},
		{30 * 24 * time.Hour, "1 month"},
		{60 * 24 * time.Hour, "2 months"},
		{365 * 24 * time.Hour, "1 year"},
		{730 * 24 * time.Hour, "2 years"},
	}

	for _, tt := range tests {
		result := HumanDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("HumanDuration(%v): expected %q, got %q", tt.duration, tt.expected, result)
		}
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"1s", 1 * time.Second},
		{"1m", 1 * time.Minute},
		{"1h", 1 * time.Hour},
		{"1d", 24 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"2h30m", 2*time.Hour + 30*time.Minute},
		{"2.5d", time.Duration(2.5 * float64(24*time.Hour))},
	}

	for _, tt := range tests {
		result, err := ParseDuration(tt.input)
		if err != nil {
			t.Errorf("ParseDuration(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("ParseDuration(%q): expected %v, got %v", tt.input, tt.expected, result)
		}
	}

	// Test error cases
	errorTests := []string{
		"abc",
		"123",      // missing unit
		"1x",       // unknown unit
		"d",        // missing number
		"1.2.3d",   // invalid number
	}

	for _, input := range errorTests {
		_, err := ParseDuration(input)
		if err == nil {
			t.Errorf("ParseDuration(%q): expected error, got nil", input)
		}
	}
}

func TestDaysBetween(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)

	days := DaysBetween(t1, t2)
	if days != 5 {
		t.Errorf("expected 5 days, got %d", days)
	}

	// Reverse order
	days = DaysBetween(t2, t1)
	if days != -5 {
		t.Errorf("expected -5 days, got %d", days)
	}
}

func TestHoursBetween(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC)

	hours := HoursBetween(t1, t2)
	if hours != 5 {
		t.Errorf("expected 5 hours, got %d", hours)
	}
}

func TestMinutesBetween(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	minutes := MinutesBetween(t1, t2)
	if minutes != 30 {
		t.Errorf("expected 30 minutes, got %d", minutes)
	}
}
