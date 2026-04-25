package redis

import (
	"context"
	"testing"
	"time"

	"github.com/f2xme/gox/cache"
)

// TestRedisLockerV2_LockWithRenewal 测试自动续期锁
func TestRedisLockerV2_LockWithRenewal(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	lockerV2, ok := c.(cache.LockerV2)
	if !ok {
		t.Fatal("Cache does not implement LockerV2")
	}

	ctx := context.Background()
	key := "renewal-lock"

	t.Run("Lock with renewal keeps lock alive", func(t *testing.T) {
		unlock, err := lockerV2.LockWithRenewal(ctx, key, 1*time.Second, 500*time.Millisecond)
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
		unlock1, err := lockerV2.TryLockWithRenewal(ctx, key, 5*time.Second, 1*time.Second)
		if err != nil {
			t.Fatalf("First TryLockWithRenewal failed: %v", err)
		}
		defer unlock1()

		_, err = lockerV2.TryLockWithRenewal(ctx, key, 5*time.Second, 1*time.Second)
		if err != cache.ErrLockFailed {
			t.Errorf("Second TryLockWithRenewal returned error %v, want %v", err, cache.ErrLockFailed)
		}
	})

	t.Run("Context cancellation stops renewal", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)

		unlock, err := lockerV2.LockWithRenewal(cancelCtx, key, 1*time.Second, 500*time.Millisecond)
		if err != nil {
			t.Fatalf("LockWithRenewal failed: %v", err)
		}
		defer unlock()

		// Cancel context to stop renewal
		cancel()

		// Wait for lock to expire
		time.Sleep(1500 * time.Millisecond)

		// Lock should have expired
		exists, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Lock still exists after context cancellation and expiration")
		}
	})
}

// TestRedisLockerV2_ReentrantLock 测试可重入锁
func TestRedisLockerV2_ReentrantLock(t *testing.T) {
	c, _ := setupTestRedis(t)
	defer c.(cache.Closer).Close()

	lockerV2, ok := c.(cache.LockerV2)
	if !ok {
		t.Fatal("Cache does not implement LockerV2")
	}

	ctx := context.Background()
	key := "reentrant-lock"
	ownerID := "owner-123"

	t.Run("Same owner can acquire lock multiple times", func(t *testing.T) {
		unlock1, err := lockerV2.LockReentrant(ctx, key, ownerID, 5*time.Second)
		if err != nil {
			t.Fatalf("First LockReentrant failed: %v", err)
		}
		defer unlock1()

		unlock2, err := lockerV2.LockReentrant(ctx, key, ownerID, 5*time.Second)
		if err != nil {
			t.Fatalf("Second LockReentrant failed: %v", err)
		}
		defer unlock2()

		unlock3, err := lockerV2.LockReentrant(ctx, key, ownerID, 5*time.Second)
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
		unlock1, err := lockerV2.TryLockReentrant(ctx, key, "owner-1", 5*time.Second)
		if err != nil {
			t.Fatalf("First TryLockReentrant failed: %v", err)
		}
		defer unlock1()

		_, err = lockerV2.TryLockReentrant(ctx, key, "owner-2", 5*time.Second)
		if err != cache.ErrLockFailed {
			t.Errorf("Second TryLockReentrant returned error %v, want %v", err, cache.ErrLockFailed)
		}
	})

	t.Run("TryLockReentrant succeeds for same owner", func(t *testing.T) {
		unlock1, err := lockerV2.TryLockReentrant(ctx, key, ownerID, 5*time.Second)
		if err != nil {
			t.Fatalf("First TryLockReentrant failed: %v", err)
		}
		defer unlock1()

		unlock2, err := lockerV2.TryLockReentrant(ctx, key, ownerID, 5*time.Second)
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

	lockerV2, ok := c.(cache.LockerV2)
	if !ok {
		t.Fatal("Cache does not implement LockerV2")
	}

	metadata, ok := c.(cache.LockMetadata)
	if !ok {
		t.Fatal("Cache does not implement LockMetadata")
	}

	ctx := context.Background()
	key := "metadata-lock"
	ownerID := "owner-456"

	t.Run("GetLockInfo returns correct info", func(t *testing.T) {
		unlock, err := lockerV2.LockReentrant(ctx, key, ownerID, 5*time.Second)
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
		unlock1, err := lockerV2.LockReentrant(ctx, key, ownerID, 5*time.Second)
		if err != nil {
			t.Fatalf("First LockReentrant failed: %v", err)
		}
		defer unlock1()

		unlock2, err := lockerV2.LockReentrant(ctx, key, ownerID, 5*time.Second)
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
