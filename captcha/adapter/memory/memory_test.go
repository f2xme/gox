package memory

import (
	"context"
	"testing"
	"time"

	"github.com/f2xme/gox/captcha"
)

func TestMemoryStore_SetGet(t *testing.T) {
	store := New()
	ctx := context.Background()

	// 设置
	err := store.Set(ctx, "test-id", "answer123", 5*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 获取
	answer, err := store.Get(ctx, "test-id")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if answer != "answer123" {
		t.Errorf("Get() = %v, want answer123", answer)
	}
}

func TestMemoryStore_GetNotFound(t *testing.T) {
	store := New()
	ctx := context.Background()

	_, err := store.Get(ctx, "non-existent")
	if err != captcha.ErrNotFound {
		t.Errorf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	store := New()
	ctx := context.Background()

	store.Set(ctx, "test-id", "answer", 5*time.Minute)

	// 删除
	err := store.Delete(ctx, "test-id")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 验证已删除
	_, err = store.Get(ctx, "test-id")
	if err != captcha.ErrNotFound {
		t.Errorf("Get() after Delete() error = %v, want ErrNotFound", err)
	}
}

func TestMemoryStore_Exists(t *testing.T) {
	store := New()
	ctx := context.Background()

	// 不存在
	exists, err := store.Exists(ctx, "test-id")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() = true, want false")
	}

	// 设置后存在
	store.Set(ctx, "test-id", "answer", 5*time.Minute)
	exists, err = store.Exists(ctx, "test-id")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() = false, want true")
	}
}

func TestMemoryStore_Expiration(t *testing.T) {
	store := New(WithTTL(100 * time.Millisecond))
	ctx := context.Background()

	store.Set(ctx, "test-id", "answer", 0) // 使用默认 TTL

	// 立即获取应该成功
	_, err := store.Get(ctx, "test-id")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 过期后应该返回 ErrNotFound
	_, err = store.Get(ctx, "test-id")
	if err != captcha.ErrNotFound {
		t.Errorf("Get() after expiration error = %v, want ErrNotFound", err)
	}
}

func TestMemoryStore_Cleanup(t *testing.T) {
	store := New(
		WithTTL(50*time.Millisecond),
		WithCleanupInterval(100*time.Millisecond),
	)
	ctx := context.Background()

	// 添加多个条目
	for i := 0; i < 10; i++ {
		store.Set(ctx, string(rune('a'+i)), "answer", 0)
	}

	// 等待过期和清理
	time.Sleep(200 * time.Millisecond)

	// 验证所有条目都被清理
	s := store.(*memoryStore)
	s.mu.RLock()
	count := len(s.items)
	s.mu.RUnlock()

	if count != 0 {
		t.Errorf("After cleanup, items count = %v, want 0", count)
	}

	// 清理
	s.Close()
}

func TestMemoryStore_Close(t *testing.T) {
	store := New()
	s := store.(*memoryStore)

	err := s.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
