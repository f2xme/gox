/*
Package cache 提供统一的缓存抽象层，支持内存缓存和分布式缓存。

cache 包定义了一组标准接口，用于缓存操作、批量操作、分布式锁和类型安全的缓存包装器。
通过这些接口，你可以轻松地在不同的缓存实现之间切换，而无需修改业务代码。

# 功能特性

  - 统一的缓存接口：Cache、MultiCache、MultiCacheV2
  - 高级缓存操作：TTL 查询、SetNX/SetXX、GetSet、Expire
  - 键扫描功能：支持 glob 模式的 Scan 操作
  - 分布式锁：基础锁、自动续期锁、可重入锁
  - 锁元数据查询：查询锁的持有者、获取时间、TTL 等信息
  - 原子计数器：支持整数和浮点数的原子递增/递减
  - 类型安全包装器：自动序列化/反序列化，防缓存击穿
  - 批量操作：GetMulti、SetMulti、DeleteMulti、ExistsMulti
  - 多种适配器：内存缓存（mem）、Redis 缓存
  - 灵活的序列化：JSON、Gob 或自定义序列化器

# 快速开始

基本用法：

	import (
		"context"
		"time"
		"github.com/f2xme/gox/cache"
		"github.com/f2xme/gox/cache/adapter/mem"
	)

	func main() {
		// 创建内存缓存
		c, _ := mem.New()
		defer c.(cache.Closer).Close()

		ctx := context.Background()

		// 设置值（TTL 为 5 分钟）
		c.Set(ctx, "user:1", []byte("Alice"), 5*time.Minute)

		// 获取值
		value, err := c.Get(ctx, "user:1")
		if err == cache.ErrNotFound {
			// 键不存在
		}

		// 删除键
		c.Delete(ctx, "user:1")
	}

# 可用适配器

## mem - 内存缓存

适用于单机应用，提供高性能的本地缓存：

	import "github.com/f2xme/gox/cache/adapter/mem"

	c, _ := mem.New(
		mem.WithMaxSize(1000),              // 最大条目数
		mem.WithEvictionPolicy("lru"),      // LRU 淘汰策略
		mem.WithCleanupInterval(time.Minute), // 清理间隔
	)

特性：
  - LRU/LFU 淘汰策略
  - 自动过期清理
  - 进程内锁
  - 零外部依赖

## redis - Redis 缓存

适用于分布式应用，提供跨进程的共享缓存：

	import (
		"github.com/f2xme/gox/cache/adapter/redis"
		goredis "github.com/redis/go-redis/v9"
	)

	client := goredis.NewClient(&goredis.Options{
		Addr: "localhost:6379",
	})

	c, _ := redis.New(
		redis.WithClient(client),
	)

特性：
  - 分布式缓存
  - 批量操作（MultiCache）
  - 分布式锁（基于 Redis）
  - 持久化支持

# 核心接口

## Cache - 基础缓存接口

所有缓存实现都必须实现此接口：

	type Cache interface {
		Get(ctx context.Context, key string) ([]byte, error)
		Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
		Delete(ctx context.Context, key string) error
		Exists(ctx context.Context, key string) (bool, error)
	}

## Advanced - 高级缓存操作

提供 TTL 查询、条件设置、原子更新等高级功能：

	type Advanced interface {
		TTL(ctx context.Context, key string) (time.Duration, error)
		SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
		SetXX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
		GetSet(ctx context.Context, key string, value []byte, ttl time.Duration) ([]byte, error)
		Expire(ctx context.Context, key string, ttl time.Duration) error
	}

使用示例：

	if adv, ok := c.(cache.Advanced); ok {
		// 查询剩余 TTL
		ttl, _ := adv.TTL(ctx, "key")

		// 仅当键不存在时设置
		ok, _ := adv.SetNX(ctx, "key", []byte("value"), time.Minute)

		// 原子性地获取旧值并设置新值
		oldValue, _ := adv.GetSet(ctx, "key", []byte("new"), time.Minute)
	}

## Scanner - 键扫描

支持 glob 模式的键遍历：

	type Scanner interface {
		Scan(ctx context.Context, pattern string, cursor uint64, count int64) ([]string, uint64, error)
	}

使用示例：

	if scanner, ok := c.(cache.Scanner); ok {
		cursor := uint64(0)
		for {
			keys, nextCursor, _ := scanner.Scan(ctx, "user:*", cursor, 100)
			// 处理 keys
			if nextCursor == 0 {
				break
			}
			cursor = nextCursor
		}
	}

## MultiCache - 批量操作接口

支持批量操作以提高性能：

	type MultiCache interface {
		Cache
		GetMulti(ctx context.Context, keys []string) (map[string][]byte, error)
		SetMulti(ctx context.Context, items map[string][]byte, ttl time.Duration) error
		DeleteMulti(ctx context.Context, keys []string) error
	}

## MultiCacheV2 - 扩展批量操作

新增批量存在性检查：

	type MultiCacheV2 interface {
		MultiCache
		ExistsMulti(ctx context.Context, keys []string) (map[string]bool, error)
	}

使用示例：

	if mc, ok := c.(cache.MultiCacheV2); ok {
		// 批量检查键是否存在
		results, _ := mc.ExistsMulti(ctx, []string{"key1", "key2", "key3"})
		for key, exists := range results {
			if exists {
				// 键存在
			}
		}
	}

## Locker - 分布式锁接口

提供锁机制以保护临界区：

	type Locker interface {
		Lock(ctx context.Context, key string, ttl time.Duration) (unlock func() error, err error)
		TryLock(ctx context.Context, key string, ttl time.Duration) (unlock func() error, err error)
	}

使用示例：

	locker := c.(cache.Locker)

	// 阻塞式获取锁
	unlock, err := locker.Lock(ctx, "resource:1", 30*time.Second)
	if err != nil {
		// 处理错误
	}
	defer unlock()

	// 执行需要保护的操作
	// ...

	// 非阻塞式尝试获取锁
	unlock, err = locker.TryLock(ctx, "resource:2", 30*time.Second)
	if err == cache.ErrLockFailed {
		// 锁已被占用
	}

## LockerV2 - 增强的分布式锁

提供自动续期和可重入锁功能：

	type LockerV2 interface {
		Locker
		LockWithRenewal(ctx context.Context, key string, ttl, renewInterval time.Duration) (unlock func() error, err error)
		TryLockWithRenewal(ctx context.Context, key string, ttl, renewInterval time.Duration) (unlock func() error, err error)
		LockReentrant(ctx context.Context, key string, ownerID string, ttl time.Duration) (unlock func() error, err error)
		TryLockReentrant(ctx context.Context, key string, ownerID string, ttl time.Duration) (unlock func() error, err error)
	}

使用示例：

	if lockerV2, ok := c.(cache.LockerV2); ok {
		// 自动续期锁（适用于长时间运行的任务）
		unlock, _ := lockerV2.LockWithRenewal(ctx, "long-task", 30*time.Second, 10*time.Second)
		defer unlock()

		// 可重入锁（同一 ownerID 可以多次获取）
		requestID := "req-123"
		unlock1, _ := lockerV2.LockReentrant(ctx, "resource", requestID, 30*time.Second)
		unlock2, _ := lockerV2.LockReentrant(ctx, "resource", requestID, 30*time.Second)
		defer unlock2()
		defer unlock1()
	}

## LockMetadata - 锁元数据查询

查询锁的状态信息：

	type LockMetadata interface {
		GetLockInfo(ctx context.Context, key string) (LockInfo, error)
	}

使用示例：

	if metadata, ok := c.(cache.LockMetadata); ok {
		info, _ := metadata.GetLockInfo(ctx, "lock-key")
		// info.Owner: 锁持有者
		// info.AcquiredAt: 获取时间
		// info.TTL: 剩余有效期
		// info.Reentrant: 是否可重入
		// info.Count: 重入计数
	}

## Counter - 原子计数器接口

提供原子递增/递减操作，适用于计数器、限流器等场景：

	type Counter interface {
		Increment(ctx context.Context, key string, delta int64) (int64, error)
		IncrementFloat(ctx context.Context, key string, delta float64) (float64, error)
	}

使用示例：

	counter := c.(cache.Counter)

	// 递增计数器
	newValue, _ := counter.Increment(ctx, "page:views", 1)

	// 递减计数器
	remaining, _ := counter.Increment(ctx, "quota:remaining", -1)

	// 浮点数计数器
	total, _ := counter.IncrementFloat(ctx, "balance", 10.5)

特性：
  - 原子操作，线程安全
  - 键不存在时自动初始化为 0
  - 支持正数（递增）和负数（递减）
  - 支持整数和浮点数

# 类型安全包装器

Typed 提供类型安全的缓存操作，自动处理序列化和反序列化：

	type User struct {
		ID   int
		Name string
	}

	// 创建类型安全的包装器
	typed := cache.NewTyped[User](c)

	// 存储结构体
	user := User{ID: 1, Name: "Alice"}
	typed.Set(ctx, "user:1", user, 5*time.Minute)

	// 获取结构体
	user, err := typed.Get(ctx, "user:1")

## Cache-Aside 模式

GetOrSet 方法实现了 cache-aside 模式，使用 singleflight 防止缓存击穿：

	product, err := typed.GetOrSet(ctx, "product:1", 5*time.Minute, func() (Product, error) {
		// 从数据库加载数据
		return loadProductFromDB(1)
	})

如果缓存命中，直接返回缓存值；如果缓存未命中，调用加载函数并将结果存入缓存。
多个并发请求同一个不存在的 key 时，只会执行一次加载函数。

## 批量操作

Typed 包装器支持批量操作，底层不支持时自动降级为循环调用：

	// 批量获取
	users, _ := typed.GetMulti(ctx, []string{"user:1", "user:2", "user:3"})

	// 批量设置
	items := map[string]User{
		"user:1": {ID: 1, Name: "Alice"},
		"user:2": {ID: 2, Name: "Bob"},
	}
	typed.SetMulti(ctx, items, 5*time.Minute)

	// 批量删除
	typed.DeleteMulti(ctx, []string{"user:1", "user:2"})

# 序列化器

支持两种内置序列化器：

## JSONSerializer（默认）

跨语言兼容，适用于需要与其他语言交互的场景：

	typed := cache.NewTyped[User](c) // 默认使用 JSON

## GobSerializer

Go 专用，性能更好，但仅限 Go 程序使用：

	typed := cache.NewTyped[User](c, cache.WithSerializer(cache.GobSerializer))

## 自定义序列化器

实现 Serializer 接口即可使用自定义序列化器：

	type Serializer interface {
		Marshal(v any) ([]byte, error)
		Unmarshal(data []byte, v any) error
	}

# 错误处理

包定义了标准错误：

  - ErrNotFound: 键不存在
  - ErrLockFailed: 无法获取锁（TryLock 专用）

示例：

	value, err := c.Get(ctx, "key")
	if err == cache.ErrNotFound {
		// 键不存在，可以设置默认值
	} else if err != nil {
		// 其他错误
	}

# 最佳实践

## 1. 使用 context 控制超时

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	value, err := c.Get(ctx, "key")

## 2. 合理设置 TTL

	// 热点数据：短 TTL
	c.Set(ctx, "trending:posts", data, 5*time.Minute)

	// 稳定数据：长 TTL
	c.Set(ctx, "config:app", data, 24*time.Hour)

	// 永不过期：TTL = 0
	c.Set(ctx, "static:data", data, 0)

## 3. 使用 Typed 包装器提高类型安全

	// 推荐：类型安全
	typed := cache.NewTyped[User](c)
	user, err := typed.Get(ctx, "user:1")

	// 不推荐：手动序列化
	data, err := c.Get(ctx, "user:1")
	json.Unmarshal(data, &user)

## 4. 长时间任务使用自动续期锁

	if lockerV2, ok := c.(cache.LockerV2); ok {
		// 锁会每 10 秒自动续期
		unlock, _ := lockerV2.LockWithRenewal(ctx, "task", 30*time.Second, 10*time.Second)
		defer unlock()

		// 执行可能需要几分钟的任务
		performLongRunningTask()
	}

## 5. 使用批量操作提高性能

	if mc, ok := c.(cache.MultiCacheV2); ok {
		// 一次性获取多个键，减少网络往返
		results, _ := mc.GetMulti(ctx, []string{"key1", "key2", "key3"})

		// 批量检查存在性
		exists, _ := mc.ExistsMulti(ctx, []string{"key1", "key2", "key3"})
	}

## 6. 优雅关闭

	c, _ := mem.New()
	defer func() {
		if closer, ok := c.(cache.Closer); ok {
			closer.Close()
		}
	}()

# 性能考虑

  - mem: 纳秒级延迟，适合高频访问的本地缓存
  - redis: 毫秒级延迟，适合分布式场景
  - 批量操作可显著减少网络开销（redis）
  - Gob 序列化比 JSON 快约 2-3 倍，但仅限 Go 使用
  - singleflight 可防止缓存击穿，减少 90%+ 的重复加载

# 线程安全

所有缓存实现都是线程安全的，可以在多个 goroutine 中并发使用。
*/
package cache
