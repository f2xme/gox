package redis

import (
	"context"
	"testing"
	"time"

	"github.com/f2xme/gox/cache"
)

// TestRedisLocker_LockWithRenewal 测试自动续期锁
func TestRedisLocker_LockWithRenewal(t *testing.T) {
	c, mr := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	locker, ok := c.(cache.Locker)
	if !ok {
		t.Fatal("Cache does not implement Locker")
	}

	ctx := context.Background()
	key := "renewal-lock"

	t.Run("Lock with renewal keeps lock alive", func(t *testing.T) {
		unlock, err := locker.LockWithRenewal(ctx, key, 1*time.Second, 500*time.Millisecond)
		if err != nil {
			t.Fatalf("LockWithRenewal failed: %v", err)
		}
		defer unlock()

		// Wait longer than initial TTL
		time.Sleep(1500 * time.Millisecond)

		// Lock should still exist due to renewal
		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("Lock expired despite renewal")
		}

		// Unlock should stop renewal
		err = unlock()
		if err != nil {
			t.Fatalf("Unlock failed: %v", err)
		}

		// Lock should be gone
		exists, err = c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Lock still exists after unlock")
		}
	})

	t.Run("TryLockWithRenewal fails when locked", func(t *testing.T) {
		unlock1, err := locker.TryLockWithRenewal(ctx, key, 5*time.Second, 1*time.Second)
		if err != nil {
			t.Fatalf("First TryLockWithRenewal failed: %v", err)
		}
		defer unlock1()

		_, err = locker.TryLockWithRenewal(ctx, key, 5*time.Second, 1*time.Second)
		if err != cache.ErrLocked {
			t.Errorf("Second TryLockWithRenewal returned error %v, want %v", err, cache.ErrLocked)
		}
	})

	t.Run("TryLockWithRenewal accepts non-positive renew interval", func(t *testing.T) {
		unlock, err := locker.TryLockWithRenewal(ctx, "renewal-lock-no-ticker", 5*time.Second, 0)
		if err != nil {
			t.Fatalf("TryLockWithRenewal failed: %v", err)
		}
		defer unlock()

		exists, err := c.Exists(ctx, "renewal-lock-no-ticker")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Fatal("lock should exist")
		}
	})

	t.Run("LockWithRenewal blocks until lock is released", func(t *testing.T) {
		key := "renewal-lock-blocking"
		unlock1, err := locker.TryLockWithRenewal(ctx, key, 5*time.Second, time.Second)
		if err != nil {
			t.Fatalf("TryLockWithRenewal failed: %v", err)
		}

		releaseDone := make(chan struct{})
		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = unlock1()
			close(releaseDone)
		}()

		waitCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		unlock2, err := locker.LockWithRenewal(waitCtx, key, 5*time.Second, time.Second)
		if err != nil {
			t.Fatalf("LockWithRenewal should wait for release: %v", err)
		}
		defer unlock2()
		<-releaseDone
	})

	t.Run("Renewal stops when lock token no longer matches", func(t *testing.T) {
		key := "renewal-lock-token-check"
		unlock, err := locker.TryLockWithRenewal(ctx, key, 200*time.Millisecond, 20*time.Millisecond)
		if err != nil {
			t.Fatalf("TryLockWithRenewal failed: %v", err)
		}
		defer unlock()

		rc := c.(*redisCache)
		if err := rc.client.Set(ctx, key, "other-token", 50*time.Millisecond).Err(); err != nil {
			t.Fatalf("overwrite lock failed: %v", err)
		}

		time.Sleep(50 * time.Millisecond)
		mr.FastForward(100 * time.Millisecond)

		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Fatal("stale renewal extended a lock owned by another token")
		}
	})

	t.Run("Context cancellation after acquire does not stop renewal", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)

		unlock, err := locker.LockWithRenewal(cancelCtx, key, 1*time.Second, 500*time.Millisecond)
		if err != nil {
			t.Fatalf("LockWithRenewal failed: %v", err)
		}
		defer unlock()

		cancel()

		time.Sleep(1500 * time.Millisecond)

		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("Lock expired after caller context cancellation")
		}
	})
}

// TestRedisLocker_ReentrantLock 测试可重入锁
func TestRedisLocker_ReentrantLock(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	locker, ok := c.(cache.Locker)
	if !ok {
		t.Fatal("Cache does not implement Locker")
	}

	ctx := context.Background()
	key := "reentrant-lock"
	ownerID := "owner-123"

	t.Run("Same owner can acquire lock multiple times", func(t *testing.T) {
		unlock1, err := locker.LockReentrant(ctx, key, ownerID, 5*time.Second)
		if err != nil {
			t.Fatalf("First LockReentrant failed: %v", err)
		}
		defer unlock1()

		unlock2, err := locker.LockReentrant(ctx, key, ownerID, 5*time.Second)
		if err != nil {
			t.Fatalf("Second LockReentrant failed: %v", err)
		}
		defer unlock2()

		unlock3, err := locker.LockReentrant(ctx, key, ownerID, 5*time.Second)
		if err != nil {
			t.Fatalf("Third LockReentrant failed: %v", err)
		}
		defer unlock3()

		// All unlocks should succeed
		err = unlock3()
		if err != nil {
			t.Fatalf("Third unlock failed: %v", err)
		}

		err = unlock2()
		if err != nil {
			t.Fatalf("Second unlock failed: %v", err)
		}

		// Lock should still exist after 2 unlocks
		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("Lock released too early")
		}

		// Final unlock should release the lock
		err = unlock1()
		if err != nil {
			t.Fatalf("First unlock failed: %v", err)
		}

		exists, err = c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Lock still exists after all unlocks")
		}
	})

	t.Run("Different owner cannot acquire lock", func(t *testing.T) {
		unlock1, err := locker.TryLockReentrant(ctx, key, "owner-1", 5*time.Second)
		if err != nil {
			t.Fatalf("First TryLockReentrant failed: %v", err)
		}
		defer unlock1()

		_, err = locker.TryLockReentrant(ctx, key, "owner-2", 5*time.Second)
		if err != cache.ErrLocked {
			t.Errorf("Second TryLockReentrant returned error %v, want %v", err, cache.ErrLocked)
		}
	})

	t.Run("TryLockReentrant succeeds for same owner", func(t *testing.T) {
		unlock1, err := locker.TryLockReentrant(ctx, key, ownerID, 5*time.Second)
		if err != nil {
			t.Fatalf("First TryLockReentrant failed: %v", err)
		}
		defer unlock1()

		unlock2, err := locker.TryLockReentrant(ctx, key, ownerID, 5*time.Second)
		if err != nil {
			t.Fatalf("Second TryLockReentrant failed: %v", err)
		}
		defer unlock2()

		unlock2()
		unlock1()
	})
}

// TestRedisLockMetadata 测试锁元数据查询
func TestRedisLockMetadata(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	locker, ok := c.(cache.Locker)
	if !ok {
		t.Fatal("Cache does not implement Locker")
	}

	metadata, ok := c.(cache.LockMetadata)
	if !ok {
		t.Fatal("Cache does not implement LockMetadata")
	}

	ctx := context.Background()
	key := "metadata-lock"
	ownerID := "owner-456"

	t.Run("GetLockInfo returns correct info", func(t *testing.T) {
		unlock, err := locker.LockReentrant(ctx, key, ownerID, 5*time.Second)
		if err != nil {
			t.Fatalf("LockReentrant failed: %v", err)
		}
		defer unlock()

		info, err := metadata.GetLockInfo(ctx, key)
		if err != nil {
			t.Fatalf("GetLockInfo failed: %v", err)
		}

		if info.Owner != ownerID {
			t.Errorf("GetLockInfo Owner = %q, want %q", info.Owner, ownerID)
		}

		if info.TTL <= 0 || info.TTL > 5*time.Second {
			t.Errorf("GetLockInfo TTL = %v, want between 0 and 5s", info.TTL)
		}

		if !info.Reentrant {
			t.Error("GetLockInfo Reentrant = false, want true")
		}

		if info.Count != 1 {
			t.Errorf("GetLockInfo Count = %d, want 1", info.Count)
		}
	})

	t.Run("GetLockInfo on non-existent lock", func(t *testing.T) {
		_, err := metadata.GetLockInfo(ctx, "non-existent")
		if err != cache.ErrNotFound {
			t.Errorf("GetLockInfo returned error %v, want %v", err, cache.ErrNotFound)
		}
	})

	t.Run("GetLockInfo shows reentrant count", func(t *testing.T) {
		unlock1, err := locker.LockReentrant(ctx, key, ownerID, 5*time.Second)
		if err != nil {
			t.Fatalf("First LockReentrant failed: %v", err)
		}
		defer unlock1()

		unlock2, err := locker.LockReentrant(ctx, key, ownerID, 5*time.Second)
		if err != nil {
			t.Fatalf("Second LockReentrant failed: %v", err)
		}
		defer unlock2()

		info, err := metadata.GetLockInfo(ctx, key)
		if err != nil {
			t.Fatalf("GetLockInfo failed: %v", err)
		}

		if info.Count != 2 {
			t.Errorf("GetLockInfo Count = %d, want 2", info.Count)
		}

		unlock2()
		unlock1()
	})
}
