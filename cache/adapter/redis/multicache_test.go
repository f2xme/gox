package redis

import (
	"context"
	"testing"

	"github.com/f2xme/gox/cache"
)

// TestRedisMultiCacheV2_ExistsMulti 测试 ExistsMulti 方法
func TestRedisMultiCacheV2_ExistsMulti(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	mc, ok := c.(cache.MultiCacheV2)
	if !ok {
		t.Fatal("Cache does not implement MultiCacheV2")
	}

	ctx := context.Background()

	t.Run("ExistsMulti with mixed keys", func(t *testing.T) {
		// Set some keys
		err := c.Set(ctx, "key1", []byte("value1"), 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		err = c.Set(ctx, "key2", []byte("value2"), 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		keys := []string{"key1", "key2", "key3", "key4"}
		result, err := mc.ExistsMulti(ctx, keys)
		if err != nil {
			t.Fatalf("ExistsMulti failed: %v", err)
		}

		expected := map[string]bool{
			"key1": true,
			"key2": true,
			"key3": false,
			"key4": false,
		}

		if len(result) != len(expected) {
			t.Errorf("ExistsMulti returned %d keys, want %d", len(result), len(expected))
		}

		for k, want := range expected {
			if got, ok := result[k]; !ok {
				t.Errorf("ExistsMulti missing key %q", k)
			} else if got != want {
				t.Errorf("ExistsMulti[%q] = %v, want %v", k, got, want)
			}
		}
	})

	t.Run("ExistsMulti with empty keys", func(t *testing.T) {
		result, err := mc.ExistsMulti(ctx, []string{})
		if err != nil {
			t.Fatalf("ExistsMulti failed: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("ExistsMulti returned %d keys, want 0", len(result))
		}
	})

	t.Run("ExistsMulti with all existing keys", func(t *testing.T) {
		err := c.Set(ctx, "all1", []byte("v1"), 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		err = c.Set(ctx, "all2", []byte("v2"), 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		keys := []string{"all1", "all2"}
		result, err := mc.ExistsMulti(ctx, keys)
		if err != nil {
			t.Fatalf("ExistsMulti failed: %v", err)
		}

		for _, k := range keys {
			if !result[k] {
				t.Errorf("ExistsMulti[%q] = false, want true", k)
			}
		}
	})

	t.Run("ExistsMulti with all non-existent keys", func(t *testing.T) {
		keys := []string{"none1", "none2", "none3"}
		result, err := mc.ExistsMulti(ctx, keys)
		if err != nil {
			t.Fatalf("ExistsMulti failed: %v", err)
		}

		for _, k := range keys {
			if result[k] {
				t.Errorf("ExistsMulti[%q] = true, want false", k)
			}
		}
	})
}
