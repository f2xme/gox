package idgen

import (
	"testing"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

func TestUUID(t *testing.T) {
	id := UUID()
	if id == uuid.Nil {
		t.Error("expected non-nil UUID")
	}

	id2 := UUID()
	if id == id2 {
		t.Error("expected unique UUIDs")
	}
}

func TestUUIDString(t *testing.T) {
	str := UUIDString()
	if len(str) != 36 {
		t.Errorf("expected length 36, got %d", len(str))
	}

	// Verify it's a valid UUID
	_, err := uuid.Parse(str)
	if err != nil {
		t.Errorf("invalid UUID string: %v", err)
	}
}

func TestULID(t *testing.T) {
	id := ULID()
	if id == (ulid.ULID{}) {
		t.Error("expected non-zero ULID")
	}

	id2 := ULID()
	if id == id2 {
		t.Error("expected unique ULIDs")
	}

	// ULIDs should be sortable by time
	if id.Compare(id2) >= 0 {
		t.Error("expected ULIDs to be time-ordered")
	}
}

func TestULIDString(t *testing.T) {
	str := ULIDString()
	if len(str) != 26 {
		t.Errorf("expected length 26, got %d", len(str))
	}

	// Verify it's a valid ULID
	_, err := ulid.Parse(str)
	if err != nil {
		t.Errorf("invalid ULID string: %v", err)
	}
}
