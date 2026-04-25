package redis

import (
	"context"
	"sort"
	"testing"

	"github.com/f2xme/gox/cache"
)

// TestRedisScanner_Scan 测试 Scan 方法
func TestRedisScanner_Scan(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	scanner, ok := c.(cache.Scanner)
	if !ok {
		t.Fatal("Cache does not implement Scanner")
	}

	ctx := context.Background()

	t.Run("Scan with exact match", func(t *testing.T) {
		// Set test keys
		testKeys := []string{"scan:test:1", "scan:test:2", "scan:test:3"}
		for _, k := range testKeys {
			err := c.Set(ctx, k, []byte("value"), 0)
			if err != nil {
				t.Fatalf("Set failed: %v", err)
			}
		}

		keys, _, err := scanner.Scan(ctx, "scan:test:*", 0, 100)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		sort.Strings(keys)
		sort.Strings(testKeys)

		if len(keys) != len(testKeys) {
			t.Errorf("Scan returned %d keys, want %d", len(keys), len(testKeys))
		}

		for i, k := range testKeys {
			if i >= len(keys) || keys[i] != k {
				t.Errorf("Scan keys[%d] = %q, want %q", i, keys[i], k)
			}
		}
	})

	t.Run("Scan with no matches", func(t *testing.T) {
		keys, _, err := scanner.Scan(ctx, "nonexistent:*", 0, 100)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		if len(keys) != 0 {
			t.Errorf("Scan returned %d keys, want 0", len(keys))
		}
	})

	t.Run("Scan with wildcard patterns", func(t *testing.T) {
		// Set test keys
		testKeys := []string{
			"pattern:a:1",
			"pattern:a:2",
			"pattern:b:1",
			"pattern:b:2",
		}
		for _, k := range testKeys {
			err := c.Set(ctx, k, []byte("value"), 0)
			if err != nil {
				t.Fatalf("Set failed: %v", err)
			}
		}

		keys, _, err := scanner.Scan(ctx, "pattern:a:*", 0, 100)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		expected := []string{"pattern:a:1", "pattern:a:2"}
		sort.Strings(keys)
		sort.Strings(expected)

		if len(keys) != len(expected) {
			t.Errorf("Scan returned %d keys, want %d", len(keys), len(expected))
		}

		for i, k := range expected {
			if i >= len(keys) || keys[i] != k {
				t.Errorf("Scan keys[%d] = %q, want %q", i, keys[i], k)
			}
		}
	})

	t.Run("Scan with question mark pattern", func(t *testing.T) {
		// Set test keys
		testKeys := []string{"q:a", "q:b", "q:c"}
		for _, k := range testKeys {
			err := c.Set(ctx, k, []byte("value"), 0)
			if err != nil {
				t.Fatalf("Set failed: %v", err)
			}
		}

		keys, _, err := scanner.Scan(ctx, "q:?", 0, 100)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		sort.Strings(keys)
		sort.Strings(testKeys)

		if len(keys) != len(testKeys) {
			t.Errorf("Scan returned %d keys, want %d", len(keys), len(testKeys))
		}
	})

	t.Run("Scan with character class pattern", func(t *testing.T) {
		// Set test keys
		testKeys := []string{"class:1", "class:2", "class:3", "class:a"}
		for _, k := range testKeys {
			err := c.Set(ctx, k, []byte("value"), 0)
			if err != nil {
				t.Fatalf("Set failed: %v", err)
			}
		}

		keys, _, err := scanner.Scan(ctx, "class:[0-9]", 0, 100)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		expected := []string{"class:1", "class:2", "class:3"}
		sort.Strings(keys)
		sort.Strings(expected)

		if len(keys) != len(expected) {
			t.Errorf("Scan returned %d keys, want %d", len(keys), len(expected))
		}

		for i, k := range expected {
			if i >= len(keys) || keys[i] != k {
				t.Errorf("Scan keys[%d] = %q, want %q", i, keys[i], k)
			}
		}
	})
}
