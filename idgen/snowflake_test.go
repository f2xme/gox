package idgen

import (
	"testing"
	"time"
)

func TestSnowflake(t *testing.T) {
	id1, err := Snowflake()
	if err != nil {
		t.Fatalf("Snowflake failed: %v", err)
	}
	id2, err := Snowflake()
	if err != nil {
		t.Fatalf("Snowflake failed: %v", err)
	}

	if id1 == 0 {
		t.Error("expected non-zero ID")
	}
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
	if id2 <= id1 {
		t.Error("expected IDs to be monotonically increasing")
	}
}

func TestSnowflakeWithNode(t *testing.T) {
	id1, err := SnowflakeWithNode(1)
	if err != nil {
		t.Fatalf("SnowflakeWithNode failed: %v", err)
	}
	id2, err := SnowflakeWithNode(2)
	if err != nil {
		t.Fatalf("SnowflakeWithNode failed: %v", err)
	}

	info1 := ParseSnowflake(id1)
	info2 := ParseSnowflake(id2)

	if info1.NodeID != 1 {
		t.Errorf("expected node ID 1, got %d", info1.NodeID)
	}
	if info2.NodeID != 2 {
		t.Errorf("expected node ID 2, got %d", info2.NodeID)
	}
}

func TestParseSnowflake(t *testing.T) {
	now := time.Now()
	id, err := SnowflakeWithNode(123)
	if err != nil {
		t.Fatalf("SnowflakeWithNode failed: %v", err)
	}

	info := ParseSnowflake(id)

	if info.NodeID != 123 {
		t.Errorf("expected node ID 123, got %d", info.NodeID)
	}

	// Timestamp should be close to now
	diff := info.Timestamp.Sub(now)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("timestamp diff too large: %v", diff)
	}
}

func TestSnowflakeConcurrency(t *testing.T) {
	const goroutines = 10
	const idsPerGoroutine = 100

	ids := make(chan int64, goroutines*idsPerGoroutine)

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := Snowflake()
				if err != nil {
					t.Errorf("Snowflake failed: %v", err)
					return
				}
				ids <- id
			}
		}()
	}

	seen := make(map[int64]bool)
	for i := 0; i < goroutines*idsPerGoroutine; i++ {
		id := <-ids
		if seen[id] {
			t.Errorf("duplicate ID: %d", id)
		}
		seen[id] = true
	}
}
