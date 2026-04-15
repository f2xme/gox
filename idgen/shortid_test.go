package idgen

import (
	"testing"
)

func TestShortID(t *testing.T) {
	id, err := ShortID()
	if err != nil {
		t.Fatalf("ShortID failed: %v", err)
	}
	if len(id) == 0 {
		t.Error("expected non-empty ID")
	}

	id2, err := ShortID()
	if err != nil {
		t.Fatalf("ShortID failed: %v", err)
	}
	if id == id2 {
		t.Error("expected unique IDs")
	}
}

func TestShortIDWithLength(t *testing.T) {
	tests := []struct {
		length int
	}{
		{6},
		{8},
		{12},
		{16},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			id, err := ShortIDWithLength(tt.length)
			if err != nil {
				t.Fatalf("ShortIDWithLength failed: %v", err)
			}
			if len(id) != tt.length {
				t.Errorf("expected length %d, got %d", tt.length, len(id))
			}
		})
	}
}

func TestShortIDUniqueness(t *testing.T) {
	const count = 1000
	seen := make(map[string]bool)

	for i := 0; i < count; i++ {
		id, err := ShortID()
		if err != nil {
			t.Fatalf("ShortID failed: %v", err)
		}
		if seen[id] {
			t.Errorf("duplicate ID: %s", id)
		}
		seen[id] = true
	}
}
