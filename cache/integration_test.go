package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/f2xme/gox/cache"
	"github.com/f2xme/gox/cache/adapter/memory"
	rediscache "github.com/f2xme/gox/cache/adapter/redis"
	"github.com/redis/go-redis/v9"
)

// setupAdapters 创建测试用的缓存适配器
// 返回包含 memory 和 redis 适配器的 map
func setupAdapters(t *testing.T) map[string]cache.Store {
	t.Helper()

	adapters := make(map[string]cache.Store)

	// 内存适配器
	c, err := memory.New()
	if err != nil {
		t.Fatalf("memory.New() error = %v", err)
	}
	adapters["memory"] = c

	// Redis 适配器（使用 miniredis）
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	rc, err := rediscache.New(rediscache.WithClient(client))
	if err != nil {
		t.Fatalf("rediscache.New() error = %v", err)
	}
	adapters["redis"] = rc

	return adapters
}

// TestAdapterConsistency 测试不同适配器的行为一致性
// 验证 memory 和 redis 适配器在基本操作上的行为一致
func TestAdapterConsistency(t *testing.T) {
	adapters := setupAdapters(t)
	defer func() {
		for _, c := range adapters {
			if closer, ok := c.(cache.Closer); ok {
				closer.Close()
			}
		}
	}()

	ctx := context.Background()

	for name, c := range adapters {
		t.Run(name, func(t *testing.T) {
			t.Run("Basic operations", func(t *testing.T) {
				key := "test-key"
				value := []byte("test-value")

				// Set
				err := c.Set(ctx, key, value, 0)
				if err != nil {
					t.Fatalf("Set failed: %v", err)
				}

				// Get
				got, err := c.Get(ctx, key)
				if err != nil {
					t.Fatalf("Get failed: %v", err)
				}
				if string(got) != string(value) {
					t.Errorf("Get returned %q, want %q", got, value)
				}

				// Exists
				exists, err := c.Exists(ctx, key)
				if err != nil {
					t.Fatalf("Exists failed: %v", err)
				}
				if !exists {
					t.Error("Exists returned false, want true")
				}

				// Delete
				err = c.Delete(ctx, key)
				if err != nil {
					t.Fatalf("Delete failed: %v", err)
				}

				// Verify deleted
				_, err = c.Get(ctx, key)
				if err != cache.ErrNotFound {
					t.Errorf("Get after delete returned error %v, want %v", err, cache.ErrNotFound)
				}
			})

			t.Run("Non-existent key", func(t *testing.T) {
				_, err := c.Get(ctx, "non-existent-key")
				if err != cache.ErrNotFound {
					t.Errorf("Get non-existent key returned error %v, want %v", err, cache.ErrNotFound)
				}

				exists, err := c.Exists(ctx, "non-existent-key")
				if err != nil {
					t.Fatalf("Exists failed: %v", err)
				}
				if exists {
					t.Error("Exists returned true for non-existent key")
				}
			})

			t.Run("Delete non-existent key", func(t *testing.T) {
				err := c.Delete(ctx, "non-existent-key")
				if err != nil {
					t.Errorf("Delete non-existent key returned error: %v", err)
				}
			})

			t.Run("Zero TTL means no expiration", func(t *testing.T) {
				key := "no-expire-key"
				value := []byte("no-expire-value")

				err := c.Set(ctx, key, value, 0)
				if err != nil {
					t.Fatalf("Set failed: %v", err)
				}

				// Should still exist after some time
				time.Sleep(50 * time.Millisecond)

				got, err := c.Get(ctx, key)
				if err != nil {
					t.Fatalf("Get failed: %v", err)
				}
				if string(got) != string(value) {
					t.Errorf("Get returned %q, want %q", got, value)
				}

				// Cleanup
				c.Delete(ctx, key)
			})
		})
	}
}

// TestTypedWithAdapters 测试 Typed 包装器与不同适配器的兼容性
// 验证类型安全包装器在不同适配器上的正确性
func TestTypedWithAdapters(t *testing.T) {
	type User struct {
		ID   int
		Name string
		Age  int
	}

	adapters := setupAdapters(t)
	defer func() {
		for _, c := range adapters {
			if closer, ok := c.(cache.Closer); ok {
				closer.Close()
			}
		}
	}()

	ctx := context.Background()

	for name, c := range adapters {
		t.Run(name, func(t *testing.T) {
			typed := cache.NewTyped[User](c)

			user := User{ID: 1, Name: "Alice", Age: 30}
			key := "user:1"

			// Set
			err := typed.Set(ctx, key, user, 0)
			if err != nil {
				t.Fatalf("Set failed: %v", err)
			}

			// Get
			got, err := typed.Get(ctx, key)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}

			if got.ID != user.ID || got.Name != user.Name || got.Age != user.Age {
				t.Errorf("Get returned %+v, want %+v", got, user)
			}

			// Exists
			exists, err := typed.Exists(ctx, key)
			if err != nil {
				t.Fatalf("Exists failed: %v", err)
			}
			if !exists {
				t.Error("Exists returned false, want true")
			}

			// Delete
			err = typed.Delete(ctx, key)
			if err != nil {
				t.Fatalf("Delete failed: %v", err)
			}

			// Verify deleted
			_, err = typed.Get(ctx, key)
			if err != cache.ErrNotFound {
				t.Errorf("Get after delete returned error %v, want %v", err, cache.ErrNotFound)
			}
		})
	}
}

// TestTypedGetOrLoad 测试 GetOrLoad 在不同适配器上的行为
// 验证 cache-aside 模式的正确实现
func TestTypedGetOrLoad(t *testing.T) {
	type Product struct {
		ID    int
		Name  string
		Price float64
	}

	adapters := setupAdapters(t)
	defer func() {
		for _, c := range adapters {
			if closer, ok := c.(cache.Closer); ok {
				closer.Close()
			}
		}
	}()

	ctx := context.Background()

	for name, c := range adapters {
		t.Run(name, func(t *testing.T) {
			typed := cache.NewTyped[Product](c)
			key := "product:1"

			callCount := 0
			loadFunc := func(context.Context) (Product, error) {
				callCount++
				return Product{ID: 1, Name: "Laptop", Price: 999.99}, nil
			}

			// 第一次调用：缓存未命中
			product, err := typed.GetOrLoad(ctx, key, 5*time.Minute, loadFunc)
			if err != nil {
				t.Fatalf("GetOrLoad failed: %v", err)
			}
			if callCount != 1 {
				t.Errorf("loadFunc called %d times, want 1", callCount)
			}
			if product.Name != "Laptop" {
				t.Errorf("GetOrLoad returned product %q, want %q", product.Name, "Laptop")
			}

			// 第二次调用：缓存命中
			product, err = typed.GetOrLoad(ctx, key, 5*time.Minute, loadFunc)
			if err != nil {
				t.Fatalf("GetOrLoad failed: %v", err)
			}
			if callCount != 1 {
				t.Errorf("loadFunc called %d times, want 1 (should use cache)", callCount)
			}
			if product.Name != "Laptop" {
				t.Errorf("GetOrLoad returned product %q, want %q", product.Name, "Laptop")
			}

			// Cleanup
			typed.Delete(ctx, key)
		})
	}
}

// TestLockConsistency 测试锁在不同适配器上的行为一致性
// 验证分布式锁的正确实现
func TestLockConsistency(t *testing.T) {
	adapters := setupAdapters(t)
	defer func() {
		for _, c := range adapters {
			if closer, ok := c.(cache.Closer); ok {
				closer.Close()
			}
		}
	}()

	ctx := context.Background()

	for name, c := range adapters {
		locker, ok := c.(cache.Locker)
		if !ok {
			t.Fatalf("%s does not implement Locker", name)
		}

		t.Run(name, func(t *testing.T) {
			t.Run("TryLock succeeds when lock is free", func(t *testing.T) {
				key := "lock:test1"
				unlock, err := locker.TryLock(ctx, key, 5*time.Second)
				if err != nil {
					t.Fatalf("TryLock failed: %v", err)
				}
				defer unlock()
			})

			t.Run("TryLock fails when lock is held", func(t *testing.T) {
				key := "lock:test2"
				unlock1, err := locker.TryLock(ctx, key, 5*time.Second)
				if err != nil {
					t.Fatalf("First TryLock failed: %v", err)
				}
				defer unlock1()

				_, err = locker.TryLock(ctx, key, 5*time.Second)
				if err != cache.ErrLocked {
					t.Errorf("Second TryLock returned error %v, want %v", err, cache.ErrLocked)
				}
			})

			t.Run("Unlock releases the lock", func(t *testing.T) {
				key := "lock:test3"
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

			t.Run("Lock blocks until acquired", func(t *testing.T) {
				key := "lock:test4"

				// First lock
				unlock1, err := locker.TryLock(ctx, key, 200*time.Millisecond)
				if err != nil {
					t.Fatalf("TryLock failed: %v", err)
				}

				acquired := make(chan bool, 1)

				// Second goroutine tries to Lock
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
				key := "lock:test5"

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
		})
	}
}

// TestBatchStoreOperations 测试批量操作（仅 redis 支持）
// 验证 BatchStore 接口的批量操作功能
func TestBatchStoreOperations(t *testing.T) {
	adapters := setupAdapters(t)
	defer func() {
		for _, c := range adapters {
			if closer, ok := c.(cache.Closer); ok {
				closer.Close()
			}
		}
	}()

	ctx := context.Background()

	for name, c := range adapters {
		mc, ok := c.(cache.BatchStore)
		if !ok {
			t.Logf("%s does not implement BatchStore, skipping", name)
			continue
		}

		t.Run(name, func(t *testing.T) {
			t.Run("SetMany and GetMany", func(t *testing.T) {
				items := map[string][]byte{
					"multi:key1": []byte("value1"),
					"multi:key2": []byte("value2"),
					"multi:key3": []byte("value3"),
				}

				err := mc.SetMany(ctx, items, 0)
				if err != nil {
					t.Fatalf("SetMany failed: %v", err)
				}

				keys := []string{"multi:key1", "multi:key2", "multi:key3", "multi:non-existent"}
				got, err := mc.GetMany(ctx, keys)
				if err != nil {
					t.Fatalf("GetMany failed: %v", err)
				}

				if len(got) != 3 {
					t.Errorf("GetMany returned %d items, want 3", len(got))
				}

				for k, v := range items {
					if string(got[k]) != string(v) {
						t.Errorf("GetMany[%q] = %q, want %q", k, got[k], v)
					}
				}

				if _, exists := got["multi:non-existent"]; exists {
					t.Error("GetMany returned non-existent key")
				}
			})

			t.Run("DeleteMany", func(t *testing.T) {
				items := map[string][]byte{
					"del:key1": []byte("value1"),
					"del:key2": []byte("value2"),
				}

				err := mc.SetMany(ctx, items, 0)
				if err != nil {
					t.Fatalf("SetMany failed: %v", err)
				}

				keys := []string{"del:key1", "del:key2"}
				err = mc.DeleteMany(ctx, keys)
				if err != nil {
					t.Fatalf("DeleteMany failed: %v", err)
				}

				got, err := mc.GetMany(ctx, keys)
				if err != nil {
					t.Fatalf("GetMany failed: %v", err)
				}

				if len(got) != 0 {
					t.Errorf("GetMany after DeleteMany returned %d items, want 0", len(got))
				}
			})

			t.Run("Empty operations", func(t *testing.T) {
				// Empty SetMany
				err := mc.SetMany(ctx, map[string][]byte{}, 0)
				if err != nil {
					t.Errorf("SetMany with empty map failed: %v", err)
				}

				// Empty GetMany
				got, err := mc.GetMany(ctx, []string{})
				if err != nil {
					t.Errorf("GetMany with empty keys failed: %v", err)
				}
				if len(got) != 0 {
					t.Errorf("GetMany with empty keys returned %d items, want 0", len(got))
				}

				// Empty DeleteMany
				err = mc.DeleteMany(ctx, []string{})
				if err != nil {
					t.Errorf("DeleteMany with empty keys failed: %v", err)
				}
			})
		})
	}
}

// TestSerializerCompatibility 测试不同序列化器的兼容性
// 验证 JSON 和 Gob 序列化器在不同适配器上的正确性
func TestSerializerCompatibility(t *testing.T) {
	type Data struct {
		Text   string
		Number int
		Flag   bool
	}

	adapters := setupAdapters(t)
	defer func() {
		for _, c := range adapters {
			if closer, ok := c.(cache.Closer); ok {
				closer.Close()
			}
		}
	}()

	ctx := context.Background()
	testData := Data{Text: "hello", Number: 42, Flag: true}

	serializers := map[string]cache.Serializer{
		"JSON": cache.JSONSerializer,
		"Gob":  cache.GobSerializer,
	}

	for adapterName, c := range adapters {
		for serializerName, serializer := range serializers {
			t.Run(adapterName+"/"+serializerName, func(t *testing.T) {
				typed := cache.NewTyped[Data](c, cache.WithSerializer(serializer))
				key := "data:" + serializerName

				// Set
				err := typed.Set(ctx, key, testData, 0)
				if err != nil {
					t.Fatalf("Set failed: %v", err)
				}

				// Get
				got, err := typed.Get(ctx, key)
				if err != nil {
					t.Fatalf("Get failed: %v", err)
				}

				if got.Text != testData.Text || got.Number != testData.Number || got.Flag != testData.Flag {
					t.Errorf("Get returned %+v, want %+v", got, testData)
				}

				// Cleanup
				typed.Delete(ctx, key)
			})
		}
	}
}
