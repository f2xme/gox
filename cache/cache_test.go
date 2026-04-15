package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockCache 实现 Cache 接口用于测试
type mockCache struct {
	data map[string][]byte
}

func newMockCache() *mockCache {
	return &mockCache{
		data: make(map[string][]byte),
	}
}

func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, ErrNotFound
}

func (m *mockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockCache) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

// TestCacheInterface 测试 Cache 接口的基本功能
// 使用 mockCache 验证接口的正确实现
func TestCacheInterface(t *testing.T) {
	ctx := context.Background()
	cache := newMockCache()

	t.Run("Set and Get", func(t *testing.T) {
		key := "test-key"
		value := []byte("test-value")

		err := cache.Set(ctx, key, value, 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		got, err := cache.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if string(got) != string(value) {
			t.Errorf("Get returned %q, want %q", got, value)
		}
	})

	t.Run("Get non-existent key returns ErrNotFound", func(t *testing.T) {
		_, err := cache.Get(ctx, "non-existent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Get returned error %v, want ErrNotFound", err)
		}
	})

	t.Run("Exists", func(t *testing.T) {
		key := "exists-key"
		value := []byte("value")

		exists, err := cache.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Exists returned true for non-existent key")
		}

		cache.Set(ctx, key, value, 0)

		exists, err = cache.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("Exists returned false for existing key")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		key := "delete-key"
		value := []byte("value")

		cache.Set(ctx, key, value, 0)

		err := cache.Delete(ctx, key)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		exists, _ := cache.Exists(ctx, key)
		if exists {
			t.Error("Key still exists after Delete")
		}

		// Delete non-existent key should not error
		err = cache.Delete(ctx, "non-existent")
		if err != nil {
			t.Errorf("Delete non-existent key returned error: %v", err)
		}
	})

	t.Run("Set with TTL", func(t *testing.T) {
		key := "ttl-key"
		value := []byte("value")

		// TTL of 0 means no expiration (just verify it doesn't error)
		err := cache.Set(ctx, key, value, 0)
		if err != nil {
			t.Fatalf("Set with TTL=0 failed: %v", err)
		}

		// Set with positive TTL (mock doesn't implement expiration, just verify signature)
		err = cache.Set(ctx, key, value, time.Minute)
		if err != nil {
			t.Fatalf("Set with TTL failed: %v", err)
		}
	})
}
