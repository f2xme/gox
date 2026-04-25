package memory

import (
	"context"
	"testing"
	"time"

	"github.com/f2xme/gox/cache"
)

// TestLRUEviction 验证 LRU 淘汰策略淘汰最近最少使用的条目
func TestLRUEviction(t *testing.T) {
	ctx := context.Background()
	c, err := New(WithMaxSize(3), WithEvictionPolicy("lru"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer c.(cache.Closer).Close()

	// Fill cache to capacity
	if err := c.Set(ctx, "key1", []byte("value1"), 0); err != nil {
		t.Fatalf("Set key1 failed: %v", err)
	}
	if err := c.Set(ctx, "key2", []byte("value2"), 0); err != nil {
		t.Fatalf("Set key2 failed: %v", err)
	}
	if err := c.Set(ctx, "key3", []byte("value3"), 0); err != nil {
		t.Fatalf("Set key3 failed: %v", err)
	}

	// 访问 key1 使其成为最近使用
	if _, err := c.Get(ctx, "key1"); err != nil {
		t.Fatalf("Get key1 failed: %v", err)
	}

	// 添加 key4，应淘汰 key2（最近最少使用）
	if err := c.Set(ctx, "key4", []byte("value4"), 0); err != nil {
		t.Fatalf("Set key4 failed: %v", err)
	}

	// 验证 key2 已被淘汰
	exists, err := c.Exists(ctx, "key2")
	if err != nil {
		t.Fatalf("Exists key2 failed: %v", err)
	}
	if exists {
		t.Error("key2 should have been evicted but still exists")
	}

	// 验证其他键仍然存在
	for _, key := range []string{"key1", "key3", "key4"} {
		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists %s failed: %v", key, err)
		}
		if !exists {
			t.Errorf("%s should exist but doesn't", key)
		}
	}
}

// TestLRUEvictionKeepsRecentlyAccessedKey 验证在 LRU 下最近访问的键不会被淘汰
func TestLRUEvictionKeepsRecentlyAccessedKey(t *testing.T) {
	ctx := context.Background()
	c, err := New(WithMaxSize(3), WithEvictionPolicy("lru"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer c.(cache.Closer).Close()

	for i := 0; i < 200; i++ {
		if err := c.Set(ctx, "key1", []byte("value1"), 0); err != nil {
			t.Fatalf("Set key1 failed: %v", err)
		}
		if err := c.Set(ctx, "key2", []byte("value2"), 0); err != nil {
			t.Fatalf("Set key2 failed: %v", err)
		}
		if err := c.Set(ctx, "key3", []byte("value3"), 0); err != nil {
			t.Fatalf("Set key3 failed: %v", err)
		}
		if _, err := c.Get(ctx, "key1"); err != nil {
			t.Fatalf("Get key1 failed: %v", err)
		}
		if err := c.Set(ctx, "key4", []byte("value4"), 0); err != nil {
			t.Fatalf("Set key4 failed: %v", err)
		}

		exists, err := c.Exists(ctx, "key1")
		if err != nil {
			t.Fatalf("Exists key1 failed: %v", err)
		}
		if !exists {
			t.Fatalf("iteration %d: key1 should exist but doesn't", i)
		}

		_ = c.Delete(ctx, "key1")
		_ = c.Delete(ctx, "key2")
		_ = c.Delete(ctx, "key3")
		_ = c.Delete(ctx, "key4")
	}
}

// TestLFUEviction 验证 LFU 淘汰策略淘汰最不经常使用的条目
func TestLFUEviction(t *testing.T) {
	ctx := context.Background()
	c, err := New(WithMaxSize(3), WithEvictionPolicy("lfu"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer c.(cache.Closer).Close()

	// Fill cache to capacity
	if err := c.Set(ctx, "key1", []byte("value1"), 0); err != nil {
		t.Fatalf("Set key1 failed: %v", err)
	}
	if err := c.Set(ctx, "key2", []byte("value2"), 0); err != nil {
		t.Fatalf("Set key2 failed: %v", err)
	}
	if err := c.Set(ctx, "key3", []byte("value3"), 0); err != nil {
		t.Fatalf("Set key3 failed: %v", err)
	}

	// 访问 key1 三次
	for i := 0; i < 3; i++ {
		if _, err := c.Get(ctx, "key1"); err != nil {
			t.Fatalf("Get key1 failed: %v", err)
		}
	}

	// 访问 key2 一次
	if _, err := c.Get(ctx, "key2"); err != nil {
		t.Fatalf("Get key2 failed: %v", err)
	}

	// key3 访问次数为 0（最不经常使用）

	// 添加 key4，应淘汰 key3（最不经常使用）
	if err := c.Set(ctx, "key4", []byte("value4"), 0); err != nil {
		t.Fatalf("Set key4 failed: %v", err)
	}

	// 验证 key3 已被淘汰
	exists, err := c.Exists(ctx, "key3")
	if err != nil {
		t.Fatalf("Exists key3 failed: %v", err)
	}
	if exists {
		t.Error("key3 should have been evicted but still exists")
	}

	// Verify other keys still exist
	for _, key := range []string{"key1", "key2", "key4"} {
		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists %s failed: %v", key, err)
		}
		if !exists {
			t.Errorf("%s should exist but doesn't", key)
		}
	}
}

// TestEvictionUpdateExistingKey 验证更新现有键不会触发淘汰
func TestEvictionUpdateExistingKey(t *testing.T) {
	ctx := context.Background()
	c, err := New(WithMaxSize(2), WithEvictionPolicy("lru"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer c.(cache.Closer).Close()

	// Fill cache to capacity
	if err := c.Set(ctx, "key1", []byte("value1"), 0); err != nil {
		t.Fatalf("Set key1 failed: %v", err)
	}
	if err := c.Set(ctx, "key2", []byte("value2"), 0); err != nil {
		t.Fatalf("Set key2 failed: %v", err)
	}

	// 更新 key1（不应触发淘汰）
	if err := c.Set(ctx, "key1", []byte("updated1"), 0); err != nil {
		t.Fatalf("Set key1 failed: %v", err)
	}

	// 两个键都应该仍然存在
	for _, key := range []string{"key1", "key2"} {
		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists %s failed: %v", key, err)
		}
		if !exists {
			t.Errorf("%s should exist but doesn't", key)
		}
	}

	// 验证 key1 的值已更新
	val, err := c.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get key1 failed: %v", err)
	}
	if string(val) != "updated1" {
		t.Errorf("key1 value = %s, want updated1", string(val))
	}
}

// TestNoEvictionWhenUnlimited 验证当 maxSize 为 0 时不会发生淘汰
func TestNoEvictionWhenUnlimited(t *testing.T) {
	ctx := context.Background()
	c, err := New(WithMaxSize(0)) // 无限制
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer c.(cache.Closer).Close()

	// 添加许多条目
	for i := 0; i < 100; i++ {
		key := string(rune('a' + i))
		if err := c.Set(ctx, key, []byte("value"), 0); err != nil {
			t.Fatalf("Set %s failed: %v", key, err)
		}
	}

	// 所有条目都应该存在
	for i := 0; i < 100; i++ {
		key := string(rune('a' + i))
		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists %s failed: %v", key, err)
		}
		if !exists {
			t.Errorf("%s should exist but doesn't", key)
		}
	}
}

// TestEvictionWithExpiration 验证淘汰与 TTL 的正确配合
func TestEvictionWithExpiration(t *testing.T) {
	ctx := context.Background()
	c, err := New(WithMaxSize(2), WithEvictionPolicy("lru"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer c.(cache.Closer).Close()

	// 添加两个条目，其中一个有短 TTL
	if err := c.Set(ctx, "key1", []byte("value1"), 50*time.Millisecond); err != nil {
		t.Fatalf("Set key1 failed: %v", err)
	}
	if err := c.Set(ctx, "key2", []byte("value2"), 0); err != nil {
		t.Fatalf("Set key2 failed: %v", err)
	}

	// 等待 key1 过期
	time.Sleep(100 * time.Millisecond)

	// 添加 key3，不应触发淘汰因为 key1 已过期
	if err := c.Set(ctx, "key3", []byte("value3"), 0); err != nil {
		t.Fatalf("Set key3 failed: %v", err)
	}

	// key2 和 key3 应该存在
	for _, key := range []string{"key2", "key3"} {
		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists %s failed: %v", key, err)
		}
		if !exists {
			t.Errorf("%s should exist but doesn't", key)
		}
	}
}

func TestEvictionOnSetNXAndCounterCreation(t *testing.T) {
	ctx := context.Background()

	t.Run("SetNX evicts when creating new key", func(t *testing.T) {
		c, err := New(WithMaxSize(2), WithEvictionPolicy("lru"))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer c.(cache.Closer).Close()

		if err := c.Set(ctx, "key1", []byte("value1"), 0); err != nil {
			t.Fatalf("Set key1 failed: %v", err)
		}
		if err := c.Set(ctx, "key2", []byte("value2"), 0); err != nil {
			t.Fatalf("Set key2 failed: %v", err)
		}

		ok, err := c.(cache.ConditionalStore).SetNX(ctx, "key3", []byte("value3"), 0)
		if err != nil {
			t.Fatalf("SetNX key3 failed: %v", err)
		}
		if !ok {
			t.Fatal("SetNX key3 returned false")
		}

		assertMaxSize(t, c, []string{"key1", "key2", "key3"}, 2)
	})

	t.Run("Incr evicts when creating new key", func(t *testing.T) {
		c, err := New(WithMaxSize(2), WithEvictionPolicy("lru"))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer c.(cache.Closer).Close()

		if err := c.Set(ctx, "key1", []byte("value1"), 0); err != nil {
			t.Fatalf("Set key1 failed: %v", err)
		}
		if err := c.Set(ctx, "key2", []byte("value2"), 0); err != nil {
			t.Fatalf("Set key2 failed: %v", err)
		}

		if _, err := c.(cache.Counter).Incr(ctx, "counter", 1); err != nil {
			t.Fatalf("Incr counter failed: %v", err)
		}

		assertMaxSize(t, c, []string{"key1", "key2", "counter"}, 2)
	})

	t.Run("IncrFloat evicts when creating new key", func(t *testing.T) {
		c, err := New(WithMaxSize(2), WithEvictionPolicy("lru"))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer c.(cache.Closer).Close()

		if err := c.Set(ctx, "key1", []byte("value1"), 0); err != nil {
			t.Fatalf("Set key1 failed: %v", err)
		}
		if err := c.Set(ctx, "key2", []byte("value2"), 0); err != nil {
			t.Fatalf("Set key2 failed: %v", err)
		}

		if _, err := c.(cache.Counter).IncrFloat(ctx, "float-counter", 1.5); err != nil {
			t.Fatalf("IncrFloat counter failed: %v", err)
		}

		assertMaxSize(t, c, []string{"key1", "key2", "float-counter"}, 2)
	})
}

func assertMaxSize(t *testing.T, c cache.Store, keys []string, max int) {
	t.Helper()

	ctx := context.Background()
	count := 0
	for _, key := range keys {
		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists %s failed: %v", key, err)
		}
		if exists {
			count++
		}
	}
	if count > max {
		t.Fatalf("cache has %d keys, want at most %d", count, max)
	}
}
