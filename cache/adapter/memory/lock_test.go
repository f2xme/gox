package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/f2xme/gox/cache"
)

// TestMemCacheTryLock 测试 TryLock 的基本功能
func TestMemCacheTryLock(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	mc := c.(*memCache)

	ctx := context.Background()
	key := "test_lock"
	ttl := 100 * time.Millisecond

	// First TryLock should succeed
	unlock1, err := mc.TryLock(ctx, key, ttl)
	if err != nil {
		t.Fatalf("First TryLock failed: %v", err)
	}
	if unlock1 == nil {
		t.Fatal("unlock function is nil")
	}

	// Second TryLock should fail with ErrLocked
	_, err = mc.TryLock(ctx, key, ttl)
	if err != cache.ErrLocked {
		t.Fatalf("Expected ErrLocked, got: %v", err)
	}

	// Unlock
	if err := unlock1(); err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Third TryLock should succeed after unlock
	unlock2, err := mc.TryLock(ctx, key, ttl)
	if err != nil {
		t.Fatalf("TryLock after unlock failed: %v", err)
	}
	if unlock2 == nil {
		t.Fatal("unlock function is nil")
	}

	// Cleanup
	_ = unlock2()
}

// TestMemCacheLock 测试 Lock 的阻塞行为
func TestMemCacheLock(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	mc := c.(*memCache)

	ctx := context.Background()
	key := "test_lock"
	ttl := 100 * time.Millisecond

	// Acquire lock in goroutine
	unlock1, err := mc.Lock(ctx, key, ttl)
	if err != nil {
		t.Fatalf("First Lock failed: %v", err)
	}

	// Try to acquire same lock in another goroutine
	acquired := make(chan bool, 1)
	go func() {
		ctx2 := context.Background()
		unlock2, err := mc.Lock(ctx2, key, ttl)
		if err != nil {
			acquired <- false
			return
		}
		acquired <- true
		_ = unlock2()
	}()

	// Should not acquire immediately
	select {
	case <-acquired:
		t.Fatal("Lock should block until first lock is released")
	case <-time.After(50 * time.Millisecond):
		// Expected: still blocking
	}

	// Release first lock
	if err := unlock1(); err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Now second lock should acquire
	select {
	case success := <-acquired:
		if !success {
			t.Fatal("Second Lock failed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Second Lock did not acquire after first unlock")
	}
}

// TestMemCacheLockExpiration 测试锁的过期机制
func TestMemCacheLockExpiration(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	mc := c.(*memCache)

	ctx := context.Background()
	key := "test_lock"
	ttl := 50 * time.Millisecond

	// Acquire lock
	unlock1, err := mc.TryLock(ctx, key, ttl)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Should be able to acquire expired lock
	unlock2, err := mc.TryLock(ctx, key, ttl)
	if err != nil {
		t.Fatalf("TryLock after expiration failed: %v", err)
	}

	// Cleanup
	_ = unlock1() // Should be safe to call
	_ = unlock2()
}

// TestMemCacheLockConcurrency 测试锁的并发安全性
func TestMemCacheLockConcurrency(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	mc := c.(*memCache)

	ctx := context.Background()
	key := "counter_lock"
	ttl := 10 * time.Millisecond

	counter := 0
	numGoroutines := 10
	incrementsPerGoroutine := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				unlock, err := mc.Lock(ctx, key, ttl)
				if err != nil {
					t.Errorf("Lock failed: %v", err)
					return
				}
				counter++
				_ = unlock()
			}
		}()
	}

	wg.Wait()

	expected := numGoroutines * incrementsPerGoroutine
	if counter != expected {
		t.Fatalf("Expected counter=%d, got %d (race condition detected)", expected, counter)
	}
}

// TestMemCacheLockContextCancellation 测试 Lock 对 context 取消的响应
func TestMemCacheLockContextCancellation(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	mc := c.(*memCache)

	key := "test_lock"
	ttl := 100 * time.Millisecond

	// Acquire lock
	ctx1 := context.Background()
	unlock1, err := mc.Lock(ctx1, key, ttl)
	if err != nil {
		t.Fatalf("First Lock failed: %v", err)
	}
	defer unlock1()

	// Try to acquire with cancelled context
	ctx2, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = mc.Lock(ctx2, key, ttl)
	if err != context.Canceled {
		t.Fatalf("Expected context.Canceled, got: %v", err)
	}
}

// TestMemCacheLockUnlockMultipleTimes 测试多次调用 unlock 的安全性
func TestMemCacheLockUnlockMultipleTimes(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	mc := c.(*memCache)

	ctx := context.Background()
	key := "test_lock"
	ttl := 100 * time.Millisecond

	unlock, err := mc.TryLock(ctx, key, ttl)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}

	// Call unlock multiple times - should be safe
	if err := unlock(); err != nil {
		t.Fatalf("First unlock failed: %v", err)
	}
	if err := unlock(); err != nil {
		t.Fatalf("Second unlock failed: %v", err)
	}
	if err := unlock(); err != nil {
		t.Fatalf("Third unlock failed: %v", err)
	}
}
