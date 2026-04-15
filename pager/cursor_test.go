package pager

import (
	"encoding/base64"
	"testing"
)

func TestNewCursor(t *testing.T) {
	tests := []struct {
		name         string
		cursor       string
		limit        int
		expectedSize int
	}{
		{"with cursor", "abc123", 20, 20},
		{"empty cursor", "", 20, 20},
		{"zero limit", "abc", 0, 10}, // should use default
		{"negative limit", "abc", -5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewCursor(tt.cursor, tt.limit)
			if page.Cursor != tt.cursor {
				t.Errorf("expected cursor %q, got %q", tt.cursor, page.Cursor)
			}
			if page.Limit != tt.expectedSize {
				t.Errorf("expected limit %d, got %d", tt.expectedSize, page.Limit)
			}
		})
	}
}

func TestEncodeCursor(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", base64.URLEncoding.EncodeToString([]byte("hello"))},
		{"123", base64.URLEncoding.EncodeToString([]byte("123"))},
		{"", ""},
	}

	for _, tt := range tests {
		result := EncodeCursor(tt.input)
		if result != tt.expected {
			t.Errorf("EncodeCursor(%q): expected %q, got %q", tt.input, tt.expected, result)
		}
	}
}

func TestDecodeCursor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{"valid", base64.URLEncoding.EncodeToString([]byte("hello")), "hello", false},
		{"empty", "", "", false},
		{"invalid base64", "not-valid-base64!!!", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeCursor(tt.input)
			if tt.hasError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestNewCursorResult(t *testing.T) {
	page := NewCursor("abc", 10)
	items := []string{"a", "b", "c"}
	nextCursor := "xyz"

	result := NewCursorResult(page, items, nextCursor)

	if result.Cursor != "abc" {
		t.Errorf("expected cursor 'abc', got %q", result.Cursor)
	}
	if result.Limit != 10 {
		t.Errorf("expected limit 10, got %d", result.Limit)
	}
	if result.NextCursor != "xyz" {
		t.Errorf("expected next cursor 'xyz', got %q", result.NextCursor)
	}
	if len(result.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(result.Items))
	}
}

func TestCursorResult_HasNext(t *testing.T) {
	tests := []struct {
		name       string
		nextCursor string
		expected   bool
	}{
		{"has next", "abc123", true},
		{"no next", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewCursor("", 10)
			result := NewCursorResult(page, []int{}, tt.nextCursor)

			if result.HasNext() != tt.expected {
				t.Errorf("expected HasNext() = %v, got %v", tt.expected, result.HasNext())
			}
		})
	}
}

func TestCursorResult_Next(t *testing.T) {
	page := NewCursor("old", 10)
	result := NewCursorResult(page, []int{}, "new")

	next := result.Next()

	if next.Cursor != "new" {
		t.Errorf("expected cursor 'new', got %q", next.Cursor)
	}
	if next.Limit != 10 {
		t.Errorf("expected limit 10, got %d", next.Limit)
	}
}

func TestEncodeDecode_RoundTrip(t *testing.T) {
	tests := []string{
		"simple",
		"with spaces",
		"123456",
		"special!@#$%",
		"unicode-中文",
	}

	for _, original := range tests {
		encoded := EncodeCursor(original)
		decoded, err := DecodeCursor(encoded)
		if err != nil {
			t.Errorf("decode error for %q: %v", original, err)
			continue
		}
		if decoded != original {
			t.Errorf("round trip failed: expected %q, got %q", original, decoded)
		}
	}
}
