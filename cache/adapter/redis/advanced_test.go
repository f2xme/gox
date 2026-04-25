package redis

import (
	"context"
	"testing"
	"time"

	"github.com/f2xme/gox/cache"
)

// TestRedisCapabilities_TTL 测试 TTL 方法
func TestRedisCapabilities_TTL(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	adv, ok := c.(cache.Expirer)
	if !ok {
		t.Fatal("Cache does not implement Expirer")
	}

	ctx := context.Background()

	t.Run("TTL for existing key with expiration", func(t *testing.T) {
		key := "ttl-key"
		err := c.Set(ctx, key, []byte("value"), 10*time.Second)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		ttl, err := adv.TTL(ctx, key)
		if err != nil {
			t.Fatalf("TTL failed: %v", err)
		}

		if ttl <= 0 || ttl > 10*time.Second {
			t.Errorf("TTL = %v, want between 0 and 10s", ttl)
		}
	})

	t.Run("TTL for non-existent key", func(t *testing.T) {
		ttl, err := adv.TTL(ctx, "non-existent")
		if err != cache.ErrNotFound {
			t.Errorf("TTL returned error %v, want %v", err, cache.ErrNotFound)
		}
		if ttl != cache.NoExpiration {
			t.Errorf("TTL = %v, want NoExpiration", ttl)
		}
	})

	t.Run("TTL for key without expiration", func(t *testing.T) {
		key := "no-ttl-key"
		err := c.Set(ctx, key, []byte("value"), 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		ttl, err := adv.TTL(ctx, key)
		if err != nil {
			t.Fatalf("TTL failed: %v", err)
		}

		if ttl != cache.NoExpiration {
			t.Errorf("TTL = %v, want NoExpiration", ttl)
		}
	})
}

// TestRedisCapabilities_SetNX 测试 SetNX 方法
func TestRedisCapabilities_SetNX(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	adv, ok := c.(cache.ConditionalStore)
	if !ok {
		t.Fatal("Cache does not implement ConditionalStore")
	}

	ctx := context.Background()

	t.Run("SetNX on non-existent key", func(t *testing.T) {
		key := "setnx-key"
		value := []byte("value")

		ok, err := adv.SetNX(ctx, key, value, 0)
		if err != nil {
			t.Fatalf("SetNX failed: %v", err)
		}
		if !ok {
			t.Error("SetNX returned false, want true")
		}

		got, err := c.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if string(got) != string(value) {
			t.Errorf("Get = %q, want %q", got, value)
		}
	})

	t.Run("SetNX on existing key", func(t *testing.T) {
		key := "setnx-existing"
		err := c.Set(ctx, key, []byte("old"), 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		ok, err := adv.SetNX(ctx, key, []byte("new"), 0)
		if err != nil {
			t.Fatalf("SetNX failed: %v", err)
		}
		if ok {
			t.Error("SetNX returned true, want false")
		}

		got, err := c.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if string(got) != "old" {
			t.Errorf("Get = %q, want %q", got, "old")
		}
	})
}

// TestRedisCapabilities_SetXX 测试 SetXX 方法
func TestRedisCapabilities_SetXX(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	adv, ok := c.(interface {
		cache.ConditionalStore
		cache.Expirer
	})
	if !ok {
		t.Fatal("Cache does not implement ConditionalStore")
	}

	ctx := context.Background()

	t.Run("SetXX on existing key", func(t *testing.T) {
		key := "setxx-key"
		err := c.Set(ctx, key, []byte("old"), 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		ok, err := adv.SetXX(ctx, key, []byte("new"), 0)
		if err != nil {
			t.Fatalf("SetXX failed: %v", err)
		}
		if !ok {
			t.Error("SetXX returned false, want true")
		}

		got, err := c.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if string(got) != "new" {
			t.Errorf("Get = %q, want %q", got, "new")
		}
	})

	t.Run("SetXX on non-existent key", func(t *testing.T) {
		key := "setxx-nonexistent"

		ok, err := adv.SetXX(ctx, key, []byte("value"), 0)
		if err != nil {
			t.Fatalf("SetXX failed: %v", err)
		}
		if ok {
			t.Error("SetXX returned true, want false")
		}

		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Key exists after failed SetXX")
		}
	})
}

// TestRedisCapabilities_Swap 测试 Swap 方法
func TestRedisCapabilities_Swap(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	adv, ok := c.(interface {
		cache.ConditionalStore
		cache.Expirer
	})
	if !ok {
		t.Fatal("Cache does not implement ConditionalStore")
	}

	ctx := context.Background()

	t.Run("Swap on existing key", func(t *testing.T) {
		key := "getset-key"
		oldValue := []byte("old")
		newValue := []byte("new")

		err := c.Set(ctx, key, oldValue, 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		got, err := adv.Swap(ctx, key, newValue, cache.NoExpiration)
		if err != nil {
			t.Fatalf("Swap failed: %v", err)
		}
		if string(got) != string(oldValue) {
			t.Errorf("Swap returned %q, want %q", got, oldValue)
		}

		current, err := c.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if string(current) != string(newValue) {
			t.Errorf("Get = %q, want %q", current, newValue)
		}
	})

	t.Run("Swap on non-existent key", func(t *testing.T) {
		key := "getset-nonexistent"
		newValue := []byte("new")

		got, err := adv.Swap(ctx, key, newValue, cache.NoExpiration)
		if err != cache.ErrNotFound {
			t.Errorf("Swap returned error %v, want %v", err, cache.ErrNotFound)
		}
		if got != nil {
			t.Errorf("Swap returned %v, want nil", got)
		}
	})

	t.Run("Swap with TTL", func(t *testing.T) {
		key := "getset-ttl"
		err := c.Set(ctx, key, []byte("old"), 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		_, err = adv.Swap(ctx, key, []byte("new"), 5*time.Second)
		if err != nil {
			t.Fatalf("Swap failed: %v", err)
		}

		ttl, err := adv.TTL(ctx, key)
		if err != nil {
			t.Fatalf("TTL failed: %v", err)
		}
		if ttl <= 0 || ttl > 5*time.Second {
			t.Errorf("TTL = %v, want between 0 and 5s", ttl)
		}
	})
}

// TestRedisCapabilities_Expire 测试 Expire 方法
func TestRedisCapabilities_Expire(t *testing.T) {
	c, mr := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	adv, ok := c.(cache.Expirer)
	if !ok {
		t.Fatal("Cache does not implement Expirer")
	}

	ctx := context.Background()

	t.Run("Expire on existing key", func(t *testing.T) {
		key := "expire-key"
		err := c.Set(ctx, key, []byte("value"), 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		err = adv.Expire(ctx, key, 1*time.Second)
		if err != nil {
			t.Fatalf("Expire failed: %v", err)
		}

		ttl, err := adv.TTL(ctx, key)
		if err != nil {
			t.Fatalf("TTL failed: %v", err)
		}
		if ttl <= 0 || ttl > 1*time.Second {
			t.Errorf("TTL = %v, want between 0 and 1s", ttl)
		}

		mr.FastForward(2 * time.Second)

		_, err = c.Get(ctx, key)
		if err != cache.ErrNotFound {
			t.Errorf("Get after expiration returned error %v, want %v", err, cache.ErrNotFound)
		}
	})

	t.Run("Expire on non-existent key", func(t *testing.T) {
		err := adv.Expire(ctx, "non-existent", 1*time.Second)
		if err != cache.ErrNotFound {
			t.Errorf("Expire returned error %v, want %v", err, cache.ErrNotFound)
		}
	})
}
