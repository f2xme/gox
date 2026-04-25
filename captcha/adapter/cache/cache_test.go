package cache

import (
	"context"
	"testing"
	"time"

	cachemem "github.com/f2xme/gox/cache/adapter/mem"
	"github.com/f2xme/gox/captcha"
)

func TestCacheStore_SetGet(t *testing.T) {
	c, _ := cachemem.New()
	store := New(c)
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

func TestCacheStore_GetNotFound(t *testing.T) {
	c, _ := cachemem.New()
	store := New(c)
	ctx := context.Background()

	_, err := store.Get(ctx, "non-existent")
	if err != captcha.ErrNotFound {
		t.Errorf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestCacheStore_Delete(t *testing.T) {
	c, _ := cachemem.New()
	store := New(c)
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

func TestCacheStore_Exists(t *testing.T) {
	c, _ := cachemem.New()
	store := New(c)
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

func TestCacheStore_Prefix(t *testing.T) {
	c, _ := cachemem.New()
	store := New(c, WithPrefix("test:"))
	ctx := context.Background()

	store.Set(ctx, "id", "answer", 5*time.Minute)

	// 验证使用了前缀
	data, err := c.Get(ctx, "test:id")
	if err != nil {
		t.Fatalf("Get() with prefix error = %v", err)
	}
	if string(data) != "answer" {
		t.Errorf("Get() = %v, want answer", string(data))
	}
}

func TestCacheStore_Expiration(t *testing.T) {
	c, _ := cachemem.New()
	store := New(c, WithTTL(100*time.Millisecond))
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
