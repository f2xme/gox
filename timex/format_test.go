package timex

import (
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	tm := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)

	tests := []struct {
		format   string
		expected string
	}{
		{DateTime, "2024-01-15 14:30:45"},
		{Date, "2024-01-15"},
		{Time, "14:30:45"},
	}

	for _, tt := range tests {
		result := Format(tm, tt.format)
		if result != tt.expected {
			t.Errorf("Format(%s): expected %q, got %q", tt.format, tt.expected, result)
		}
	}
}

func TestFormatUnix(t *testing.T) {
	// 2024-01-15 14:30:45 UTC
	unix := int64(1705329045)
	result := FormatUnix(unix, DateTime)
	expected := "2024-01-15 14:30:45"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"2024-01-15 14:30:45", time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)},
		{"2024-01-15", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"2024-01-15T14:30:45Z", time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)},
	}

	for _, tt := range tests {
		result, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if !result.Equal(tt.expected) {
			t.Errorf("Parse(%q): expected %v, got %v", tt.input, tt.expected, result)
		}
	}

	// Test error case
	_, err := Parse("invalid-date-format")
	if err == nil {
		t.Error("Parse(invalid): expected error, got nil")
	}
}

func TestFromUnix(t *testing.T) {
	unix := int64(1705329045)
	result := FromUnix(unix)
	expected := time.Unix(unix, 0)
	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestFromUnixMilli(t *testing.T) {
	milli := int64(1705329045000)
	result := FromUnixMilli(milli)
	expected := time.UnixMilli(milli)
	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
