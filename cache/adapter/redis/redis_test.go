package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/f2xme/gox/cache"
	"github.com/redis/go-redis/v9"
)

// setupTestRedis 创建测试用的 Redis 缓存实例
func setupTestRedis(t *testing.T) (cache.Cache, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	c, err := New(WithClient(client))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return c, mr
}

// TestRedisCacheBasicOperations 测试 Redis 缓存的基本操作
func TestRedisCacheBasicOperations(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	ctx := context.Background()

	t.Run("Set and Get", func(t *testing.T) {
		key := "test-key"
		value := []byte("test-value")

		err := c.Set(ctx, key, value, 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		got, err := c.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if string(got) != string(value) {
			t.Errorf("Get returned %q, want %q", got, value)
		}
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		_, err := c.Get(ctx, "non-existent")
		if err != cache.ErrNotFound {
			t.Errorf("Get returned error %v, want %v", err, cache.ErrNotFound)
		}
	})

	t.Run("Exists", func(t *testing.T) {
		key := "exists-key"
		value := []byte("value")

		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Exists returned true for non-existent key")
		}

		err = c.Set(ctx, key, value, 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		exists, err = c.Exists(ctx, key)
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

		err := c.Set(ctx, key, value, 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		err = c.Delete(ctx, key)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Key still exists after Delete")
		}

		// Delete non-existent key should not error
		err = c.Delete(ctx, "non-existent")
		if err != nil {
			t.Errorf("Delete non-existent key returned error: %v", err)
		}
	})
}

// TestRedisCacheExpiration 测试 Redis 缓存过期功能
func TestRedisCacheExpiration(t *testing.T) {
	c, mr := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	ctx := context.Background()
	key := "ttl-key"
	value := []byte("ttl-value")

	// Set with 1 second TTL
	err := c.Set(ctx, key, value, 1*time.Second)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Should exist immediately
	got, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("Get returned %q, want %q", got, value)
	}

	// Fast-forward time in miniredis
	mr.FastForward(2 * time.Second)

	// Should be expired
	_, err = c.Get(ctx, key)
	if err != cache.ErrNotFound {
		t.Errorf("Get after expiration returned error %v, want %v", err, cache.ErrNotFound)
	}
}

// TestRedisCacheBatchOperations 测试 Redis 批量操作
func TestRedisCacheBatchOperations(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	mc, ok := c.(cache.MultiCache)
	if !ok {
		t.Fatal("Cache does not implement MultiCache")
	}

	ctx := context.Background()

	t.Run("SetMulti and GetMulti", func(t *testing.T) {
		items := map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte("value2"),
			"key3": []byte("value3"),
		}

		err := mc.SetMulti(ctx, items, 0)
		if err != nil {
			t.Fatalf("SetMulti failed: %v", err)
		}

		keys := []string{"key1", "key2", "key3", "non-existent"}
		got, err := mc.GetMulti(ctx, keys)
		if err != nil {
			t.Fatalf("GetMulti failed: %v", err)
		}

		if len(got) != 3 {
			t.Errorf("GetMulti returned %d items, want 3", len(got))
		}

		for k, v := range items {
			if string(got[k]) != string(v) {
				t.Errorf("GetMulti[%q] = %q, want %q", k, got[k], v)
			}
		}

		if _, exists := got["non-existent"]; exists {
			t.Error("GetMulti returned non-existent key")
		}
	})

	t.Run("DeleteMulti", func(t *testing.T) {
		items := map[string][]byte{
			"del1": []byte("value1"),
			"del2": []byte("value2"),
			"del3": []byte("value3"),
		}

		err := mc.SetMulti(ctx, items, 0)
		if err != nil {
			t.Fatalf("SetMulti failed: %v", err)
		}

		keys := []string{"del1", "del2", "del3"}
		err = mc.DeleteMulti(ctx, keys)
		if err != nil {
			t.Fatalf("DeleteMulti failed: %v", err)
		}

		got, err := mc.GetMulti(ctx, keys)
		if err != nil {
			t.Fatalf("GetMulti failed: %v", err)
		}

		if len(got) != 0 {
			t.Errorf("GetMulti after DeleteMulti returned %d items, want 0", len(got))
		}
	})
}

func TestRedisCacheClose(t *testing.T) {
	c, _ := setupTestRedis(t)

	closer, ok := c.(cache.Closer)
	if !ok {
		t.Fatal("Cache does not implement Closer")
	}

	err := closer.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// TestRedisCacheTryLock 测试 Redis TryLock 功能
func TestRedisCacheTryLock(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	locker, ok := c.(cache.Locker)
	if !ok {
		t.Fatal("Cache does not implement Locker")
	}

	ctx := context.Background()
	key := "test-lock"

	t.Run("Acquire lock successfully", func(t *testing.T) {
		unlock, err := locker.TryLock(ctx, key, 5*time.Second)
		if err != nil {
			t.Fatalf("TryLock failed: %v", err)
		}
		defer unlock()

		// Verify lock exists
		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("Lock key does not exist after TryLock")
		}
	})

	t.Run("TryLock returns ErrLockFailed when already held", func(t *testing.T) {
		unlock1, err := locker.TryLock(ctx, key, 5*time.Second)
		if err != nil {
			t.Fatalf("First TryLock failed: %v", err)
		}
		defer unlock1()

		// Second TryLock should fail immediately
		_, err = locker.TryLock(ctx, key, 5*time.Second)
		if err != cache.ErrLockFailed {
			t.Errorf("Second TryLock returned error %v, want %v", err, cache.ErrLockFailed)
		}
	})

	t.Run("Unlock releases the lock", func(t *testing.T) {
		unlock, err := locker.TryLock(ctx, key, 5*time.Second)
		if err != nil {
			t.Fatalf("TryLock failed: %v", err)
		}

		err = unlock()
		if err != nil {
			t.Fatalf("Unlock failed: %v", err)
		}

		// Should be able to acquire again
		unlock2, err := locker.TryLock(ctx, key, 5*time.Second)
		if err != nil {
			t.Fatalf("TryLock after unlock failed: %v", err)
		}
		defer unlock2()
	})

	t.Run("Unlock is idempotent", func(t *testing.T) {
		unlock, err := locker.TryLock(ctx, key, 5*time.Second)
		if err != nil {
			t.Fatalf("TryLock failed: %v", err)
		}

		err = unlock()
		if err != nil {
			t.Fatalf("First unlock failed: %v", err)
		}

		// Second unlock should not error
		err = unlock()
		if err != nil {
			t.Errorf("Second unlock returned error: %v", err)
		}
	})
}

// TestRedisCacheLock 测试 Redis Lock 的阻塞行为
func TestRedisCacheLock(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	locker, ok := c.(cache.Locker)
	if !ok {
		t.Fatal("Cache does not implement Locker")
	}

	ctx := context.Background()
	key := "test-lock-blocking"

	t.Run("Lock blocks until acquired", func(t *testing.T) {
		// First goroutine acquires lock
		unlock1, err := locker.TryLock(ctx, key, 2*time.Second)
		if err != nil {
			t.Fatalf("TryLock failed: %v", err)
		}

		acquired := make(chan bool, 1)

		// Second goroutine tries to Lock (should block)
		go func() {
			unlock2, err := locker.Lock(ctx, key, 5*time.Second)
			if err != nil {
				t.Errorf("Lock failed: %v", err)
				return
			}
			defer unlock2()
			acquired <- true
		}()

		// Give second goroutine time to start blocking
		time.Sleep(50 * time.Millisecond)

		// Release first lock
		unlock1()

		// Second goroutine should acquire lock
		select {
		case <-acquired:
			// Success
		case <-time.After(1 * time.Second):
			t.Error("Lock did not acquire after first lock released")
		}
	})

	t.Run("Lock respects context cancellation", func(t *testing.T) {
		// Acquire lock
		unlock, err := locker.TryLock(ctx, key, 5*time.Second)
		if err != nil {
			t.Fatalf("TryLock failed: %v", err)
		}
		defer unlock()

		// Try to Lock with cancelled context
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()

		_, err = locker.Lock(cancelCtx, key, 5*time.Second)
		if err != context.Canceled {
			t.Errorf("Lock with cancelled context returned error %v, want %v", err, context.Canceled)
		}
	})
}

// TestRedisCacheLockExpiration 测试 Redis 锁的过期机制
func TestRedisCacheLockExpiration(t *testing.T) {
	c, mr := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	locker, ok := c.(cache.Locker)
	if !ok {
		t.Fatal("Cache does not implement Locker")
	}

	ctx := context.Background()
	key := "test-lock-expiration"

	// Acquire lock with short TTL
	unlock, err := locker.TryLock(ctx, key, 1*time.Second)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	defer unlock()

	// Fast-forward time
	mr.FastForward(2 * time.Second)

	// Lock should have expired, another TryLock should succeed
	unlock2, err := locker.TryLock(ctx, key, 5*time.Second)
	if err != nil {
		t.Fatalf("TryLock after expiration failed: %v", err)
	}
	defer unlock2()
}

// TestRedisCacheLockConcurrency 测试 Redis 锁的并发安全性
func TestRedisCacheLockConcurrency(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	locker, ok := c.(cache.Locker)
	if !ok {
		t.Fatal("Cache does not implement Locker")
	}

	ctx := context.Background()
	key := "test-lock-concurrent"

	const numGoroutines = 10
	counter := 0
	done := make(chan bool, numGoroutines)

	// Multiple goroutines competing for the same lock
	for range numGoroutines {
		go func() {
			unlock, err := locker.Lock(ctx, key, 1*time.Second)
			if err != nil {
				t.Errorf("Lock failed: %v", err)
				done <- false
				return
			}

			// Critical section - protected by distributed lock
			// Note: counter access is safe here because only one goroutine
			// can hold the distributed lock at a time
			temp := counter
			time.Sleep(10 * time.Millisecond) // Simulate work
			counter = temp + 1

			unlock()
			done <- true
		}()
	}

	// Wait for all goroutines
	for range numGoroutines {
		<-done
	}

	// Counter should be exactly numGoroutines if lock works correctly
	if counter != numGoroutines {
		t.Errorf("Counter = %d, want %d (lock did not protect critical section)", counter, numGoroutines)
	}
}

// TestRedisCacheUnlockOnlyOwnLock 测试只能解锁自己持有的锁
func TestRedisCacheUnlockOnlyOwnLock(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	locker, ok := c.(cache.Locker)
	if !ok {
		t.Fatal("Cache does not implement Locker")
	}

	ctx := context.Background()
	key := "test-lock-ownership"

	// First lock acquires
	unlock1, err := locker.TryLock(ctx, key, 5*time.Second)
	if err != nil {
		t.Fatalf("First TryLock failed: %v", err)
	}

	// Manually set a different value (simulating another process's lock)
	err = c.Set(ctx, key, []byte("different-token"), 5*time.Second)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Unlock should not delete the lock (different token)
	err = unlock1()
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Lock should still exist
	exists, err := c.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Lock was deleted even though token didn't match")
	}
}

// TestNew_Validation 测试 New 函数的参数验证
func TestNew_Validation(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty addr without client",
			opts:    []Option{WithAddr(""), WithPassword("pass")},
			wantErr: true,
			errMsg:  "addr cannot be empty",
		},
		{
			name:    "invalid db negative",
			opts:    []Option{WithAddr("localhost:6379"), WithDB(-1)},
			wantErr: true,
			errMsg:  "db must be between 0 and 15",
		},
		{
			name:    "invalid db too large",
			opts:    []Option{WithAddr("localhost:6379"), WithDB(16)},
			wantErr: true,
			errMsg:  "db must be between 0 and 15",
		},
		{
			name:    "valid with client",
			opts:    []Option{WithClient(redis.NewClient(&redis.Options{Addr: "localhost:6379"}))},
			wantErr: false,
		},
		{
			name:    "valid with addr",
			opts:    []Option{WithAddr("localhost:6379")},
			wantErr: false,
		},
		{
			name:    "valid with addr and valid db",
			opts:    []Option{WithAddr("localhost:6379"), WithDB(5)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.opts...)
			if tt.wantErr {
				if err == nil {
					t.Error("New() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("New() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("New() unexpected error = %v", err)
					return
				}
				if c != nil {
					if closer, ok := c.(cache.Closer); ok {
						closer.Close()
					}
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
