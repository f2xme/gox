package cache_test

import (
	"context"
	"fmt"
	"time"

	"github.com/f2xme/gox/cache"
	"github.com/f2xme/gox/cache/adapter/mem"
	rediscache "github.com/f2xme/gox/cache/adapter/redis"
	"github.com/redis/go-redis/v9"
)

// ExampleCache_basic 演示基本的缓存操作
func ExampleCache_basic() {
	// 创建内存缓存
	c, _ := mem.New()
	defer c.(cache.Closer).Close()

	ctx := context.Background()

	// 设置值
	_ = c.Set(ctx, "user:1", []byte("Alice"), 5*time.Minute)

	// 获取值
	value, err := c.Get(ctx, "user:1")
	if err == nil {
		fmt.Println(string(value))
	}

	// 检查键是否存在
	exists, _ := c.Exists(ctx, "user:1")
	fmt.Println(exists)

	// 删除键
	_ = c.Delete(ctx, "user:1")

	// Output:
	// Alice
	// true
}

// ExampleTyped 演示类型安全的缓存包装器
func ExampleTyped() {
	type User struct {
		ID   int
		Name string
	}

	// 创建内存缓存
	c, _ := mem.New()
	defer c.(cache.Closer).Close()

	// 创建类型安全的包装器
	typed := cache.NewTyped[User](c)

	ctx := context.Background()

	// 存储结构体
	user := User{ID: 1, Name: "Alice"}
	_ = typed.Set(ctx, "user:1", user, 5*time.Minute)

	// 获取结构体
	retrieved, err := typed.Get(ctx, "user:1")
	if err == nil {
		fmt.Printf("%s (ID: %d)\n", retrieved.Name, retrieved.ID)
	}

	// Output:
	// Alice (ID: 1)
}

// ExampleTyped_GetOrSet 演示 cache-aside 模式
func ExampleTyped_GetOrSet() {
	type Product struct {
		ID    int
		Name  string
		Price float64
	}

	c, _ := mem.New()
	defer c.(cache.Closer).Close()

	typed := cache.NewTyped[Product](c)
	ctx := context.Background()

	// 模拟数据库查询函数
	loadFromDB := func() (Product, error) {
		fmt.Println("Loading from database...")
		return Product{ID: 1, Name: "Laptop", Price: 999.99}, nil
	}

	// 第一次调用：缓存未命中，从数据库加载
	product, _ := typed.GetOrSet(ctx, "product:1", 5*time.Minute, loadFromDB)
	fmt.Printf("%s: $%.2f\n", product.Name, product.Price)

	// 第二次调用：缓存命中，不会调用 loadFromDB
	product, _ = typed.GetOrSet(ctx, "product:1", 5*time.Minute, loadFromDB)
	fmt.Printf("%s: $%.2f\n", product.Name, product.Price)

	// Output:
	// Loading from database...
	// Laptop: $999.99
	// Laptop: $999.99
}

// ExampleCache_batch 演示批量操作
func Example_batch() {
	// 创建 Redis 缓存（需要 Redis 服务器运行）
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	c, _ := rediscache.New(rediscache.WithClient(client))
	defer c.(cache.Closer).Close()

	// 类型断言为 MultiCache
	mc, ok := c.(cache.MultiCache)
	if !ok {
		fmt.Println("Cache does not support batch operations")
		return
	}

	ctx := context.Background()

	// 批量设置
	items := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}
	err := mc.SetMulti(ctx, items, 5*time.Minute)
	if err != nil {
		fmt.Println("Redis not available, skipping example")
		return
	}

	// 批量获取
	keys := []string{"key1", "key2", "key3"}
	results, err := mc.GetMulti(ctx, keys)
	if err != nil {
		fmt.Println("Redis not available, skipping example")
		return
	}

	for _, key := range keys {
		if value, ok := results[key]; ok {
			fmt.Printf("%s: %s\n", key, value)
		}
	}

	// 批量删除
	_ = mc.DeleteMulti(ctx, keys)
}

// ExampleLocker_mem 演示内存缓存的锁
func Example_lock_mem() {
	c, _ := mem.New()
	defer c.(cache.Closer).Close()

	locker, ok := c.(cache.Locker)
	if !ok {
		fmt.Println("Cache does not support locking")
		return
	}

	ctx := context.Background()

	// 尝试获取锁
	unlock, err := locker.TryLock(ctx, "resource:1", 10*time.Second)
	if err != nil {
		fmt.Println("Failed to acquire lock")
		return
	}
	defer unlock()

	// 执行需要保护的操作
	fmt.Println("Lock acquired, performing critical operation...")

	// Output:
	// Lock acquired, performing critical operation...
}

// ExampleLocker_redis 演示 Redis 分布式锁
func Example_lock_redis() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	c, _ := rediscache.New(rediscache.WithClient(client))
	defer c.(cache.Closer).Close()

	locker, ok := c.(cache.Locker)
	if !ok {
		fmt.Println("Cache does not support locking")
		return
	}

	ctx := context.Background()

	// 阻塞式获取锁
	unlock, err := locker.Lock(ctx, "distributed:lock:1", 30*time.Second)
	if err != nil {
		fmt.Println("Redis not available, skipping example")
		return
	}
	defer unlock()

	// 执行分布式环境下需要保护的操作
	fmt.Println("Distributed lock acquired")
}

// ExampleCache_serializer 演示自定义序列化器
func Example_serializer() {
	type Config struct {
		Host string
		Port int
	}

	c, _ := mem.New()
	defer c.(cache.Closer).Close()

	// 使用 Gob 序列化器（更快，但仅限 Go）
	typed := cache.NewTyped[Config](c, cache.WithSerializer(cache.GobSerializer))

	ctx := context.Background()

	config := Config{Host: "localhost", Port: 8080}
	_ = typed.Set(ctx, "config", config, 0)

	retrieved, _ := typed.Get(ctx, "config")
	fmt.Printf("%s:%d\n", retrieved.Host, retrieved.Port)

	// Output:
	// localhost:8080
}
