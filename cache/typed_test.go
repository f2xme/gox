package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

// User 是测试用的用户结构体
type User struct {
	ID   int
	Name string
	Age  int
}

// TestTypedBasicOperations 测试 Typed 包装器的基本操作
// 验证 Set、Get、Exists、Delete 的正确性
func TestTypedBasicOperations(t *testing.T) {
	ctx := context.Background()
	cache := newMockCache()
	typed := NewTyped[User](cache)

	user := User{ID: 1, Name: "Alice", Age: 30}

	// 测试 Set
	err := typed.Set(ctx, "user:1", user, 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 测试 Get
	got, err := typed.Get(ctx, "user:1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.ID != user.ID || got.Name != user.Name || got.Age != user.Age {
		t.Errorf("Get returned %+v, want %+v", got, user)
	}

	// 测试 Exists
	exists, err := typed.Exists(ctx, "user:1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Exists returned false, want true")
	}

	// 测试 Delete
	err = typed.Delete(ctx, "user:1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 测试 Get 不存在的键返回 ErrNotFound
	_, err = typed.Get(ctx, "user:1")
	if err != ErrNotFound {
		t.Errorf("Get non-existent returned %v, want ErrNotFound", err)
	}

	// 测试删除后 Exists 返回 false
	exists, err = typed.Exists(ctx, "user:1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Exists returned true, want false")
	}
}

// TestTypedGetOrLoad 测试 GetOrLoad 的 cache-aside 模式
// 验证缓存命中和未命中时的行为
func TestTypedGetOrLoad(t *testing.T) {
	ctx := context.Background()
	cache := newMockCache()
	typed := NewTyped[User](cache)

	user := User{ID: 2, Name: "Bob", Age: 25}
	callCount := 0

	loader := func(context.Context) (User, error) {
		callCount++
		return user, nil
	}

	// 第一次调用：缓存未命中，应调用 loader
	got, err := typed.GetOrLoad(ctx, "user:2", time.Minute, loader)
	if err != nil {
		t.Fatalf("GetOrLoad failed: %v", err)
	}
	if got.ID != user.ID || got.Name != user.Name || got.Age != user.Age {
		t.Errorf("GetOrLoad returned %+v, want %+v", got, user)
	}
	if callCount != 1 {
		t.Errorf("loader called %d times, want 1", callCount)
	}

	// 第二次调用：缓存命中，不应调用 loader
	got, err = typed.GetOrLoad(ctx, "user:2", time.Minute, loader)
	if err != nil {
		t.Fatalf("GetOrLoad failed: %v", err)
	}
	if got.ID != user.ID || got.Name != user.Name || got.Age != user.Age {
		t.Errorf("GetOrLoad returned %+v, want %+v", got, user)
	}
	if callCount != 1 {
		t.Errorf("loader called %d times, want 1 (should use cached value)", callCount)
	}
}

// TestTypedWithGobSerializer 测试使用 Gob 序列化器
// 验证 Gob 序列化器的正确性
func TestTypedWithGobSerializer(t *testing.T) {
	ctx := context.Background()
	cache := newMockCache()
	typed := NewTyped[User](cache, WithSerializer(GobSerializer))

	user := User{ID: 3, Name: "Charlie", Age: 35}

	// 使用 Gob 序列化器测试 Set
	err := typed.Set(ctx, "user:3", user, 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 使用 Gob 序列化器测试 Get
	got, err := typed.Get(ctx, "user:3")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.ID != user.ID || got.Name != user.Name || got.Age != user.Age {
		t.Errorf("Get returned %+v, want %+v", got, user)
	}
}

// TestTypedGetOrLoadCacheFailure 测试缓存写入失败时 GetOrLoad 的默认行为。
func TestTypedGetOrLoadCacheFailure(t *testing.T) {
	ctx := context.Background()
	cache := &failingCache{mockCache: newMockCache()}
	typed := NewTyped[User](cache)

	user := User{ID: 4, Name: "Dave", Age: 40}
	loader := func(context.Context) (User, error) {
		return user, nil
	}

	_, err := typed.GetOrLoad(ctx, "user:4", time.Minute, loader)
	if err == nil {
		t.Fatal("GetOrLoad should return cache Set error by default")
	}

	typed = NewTyped[User](cache, WithIgnoreSetErrors())

	got, err := typed.GetOrLoad(ctx, "user:4", time.Minute, loader)
	if err != nil {
		t.Errorf("GetOrLoad should succeed with WithIgnoreSetErrors, got error: %v", err)
	}

	if got.ID != user.ID || got.Name != user.Name || got.Age != user.Age {
		t.Errorf("GetOrLoad returned %+v, want %+v", got, user)
	}
}

// failingCache 包装 mockCache 并在 Set 操作时失败
type failingCache struct {
	*mockCache
}

func (f *failingCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return errors.New("cache set failed")
}
