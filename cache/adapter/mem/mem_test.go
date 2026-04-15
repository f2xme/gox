package mem

import (
	"context"
	"testing"
	"time"

	"github.com/f2xme/gox/cache"
)

// TestMemCacheBasicOperations 测试内存缓存的基本操作
func TestMemCacheBasicOperations(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer c.(cache.Closer).Close()

	ctx := context.Background()

	// Test Set and Get
	key := "test-key"
	value := []byte("test-value")
	err = c.Set(ctx, key, value, 0)
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

	// Test Exists
	exists, err := c.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Exists returned false, want true")
	}

	// Test Delete
	err = c.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	exists, err = c.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists after delete failed: %v", err)
	}
	if exists {
		t.Error("Exists after delete returned true, want false")
	}

	// Get should return ErrNotFound
	_, err = c.Get(ctx, key)
	if err != cache.ErrNotFound {
		t.Errorf("Get after delete returned error %v, want %v", err, cache.ErrNotFound)
	}
}

// TestMemCacheExpiration 测试缓存过期功能
func TestMemCacheExpiration(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer c.(cache.Closer).Close()

	ctx := context.Background()
	key := "expiring-key"
	value := []byte("expiring-value")

	// Set with short TTL
	err = c.Set(ctx, key, value, 100*time.Millisecond)
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

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, err = c.Get(ctx, key)
	if err != cache.ErrNotFound {
		t.Errorf("Get after expiration returned error %v, want %v", err, cache.ErrNotFound)
	}

	exists, err := c.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists after expiration failed: %v", err)
	}
	if exists {
		t.Error("Exists after expiration returned true, want false")
	}
}

// TestMemCacheNoExpiration 测试无过期时间的缓存
func TestMemCacheNoExpiration(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer c.(cache.Closer).Close()

	ctx := context.Background()
	key := "no-expire-key"
	value := []byte("no-expire-value")

	// Set with TTL=0 (no expiration)
	err = c.Set(ctx, key, value, 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Should still exist
	got, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("Get returned %q, want %q", got, value)
	}

	exists, err := c.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Exists returned false, want true")
	}
}

// TestMemCacheClose 测试缓存关闭功能
func TestMemCacheClose(t *testing.T) {
	c, err := New(WithCleanupInterval(50 * time.Millisecond))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := context.Background()
	key := "test-key"
	value := []byte("test-value")

	err = c.Set(ctx, key, value, 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Close the cache
	closer := c.(cache.Closer)
	err = closer.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Wait a bit to ensure cleanup goroutine has stopped
	time.Sleep(100 * time.Millisecond)

	// Operations after close should still work (data is still in memory)
	got, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get after close failed: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("Get after close returned %q, want %q", got, value)
	}
}
