package idgen

import (
	"sync"
	"testing"
)

func TestAutoIncrement(t *testing.T) {
	gen := NewAutoIncrement(100)

	id1 := gen.Next()
	if id1 != 100 {
		t.Errorf("expected 100, got %d", id1)
	}

	id2 := gen.Next()
	if id2 != 101 {
		t.Errorf("expected 101, got %d", id2)
	}
}

func TestAutoIncrementNextN(t *testing.T) {
	gen := NewAutoIncrement(1)

	ids := gen.NextN(5)
	expected := []int64{1, 2, 3, 4, 5}

	if len(ids) != len(expected) {
		t.Fatalf("expected %d IDs, got %d", len(expected), len(ids))
	}

	for i, id := range ids {
		if id != expected[i] {
			t.Errorf("index %d: expected %d, got %d", i, expected[i], id)
		}
	}
}

func TestAutoIncrementConcurrency(t *testing.T) {
	gen := NewAutoIncrement(0)
	const goroutines = 10
	const idsPerGoroutine = 100

	ids := make(chan int64, goroutines*idsPerGoroutine)
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				ids <- gen.Next()
			}
		}()
	}

	wg.Wait()
	close(ids)

	seen := make(map[int64]bool)
	for id := range ids {
		if seen[id] {
			t.Errorf("duplicate ID: %d", id)
		}
		seen[id] = true
	}

	if len(seen) != goroutines*idsPerGoroutine {
		t.Errorf("expected %d unique IDs, got %d", goroutines*idsPerGoroutine, len(seen))
	}
}

func TestAutoIncrementGenerate(t *testing.T) {
	gen := NewAutoIncrement(1000)

	id := gen.Generate()
	if id != "1000" {
		t.Errorf("expected '1000', got %q", id)
	}

	id = gen.Generate()
	if id != "1001" {
		t.Errorf("expected '1001', got %q", id)
	}
}
