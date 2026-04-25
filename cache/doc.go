/*
Package cache 提供统一的缓存能力接口，支持内存缓存、Redis 缓存和类型安全包装器。

接口按能力拆分，调用方可以按需断言：

  - Store：基础 Get、Set、Delete、Exists。
  - BatchStore：GetMany、SetMany、DeleteMany、ExistsMany。
  - Expirer：TTL、Expire、Persist。
  - ConditionalStore：SetNX、SetXX、Swap。
  - Counter：整数和浮点原子计数。
  - Scanner：游标式 key 扫描。
  - Locker：锁能力。
  - Closer：资源释放。

# TTL 语义

NoExpiration 表示永不过期，KeepTTL 表示在支持的条件更新中保留原 TTL：

	c.Set(ctx, "key", []byte("value"), cache.NoExpiration)

	if expirer, ok := c.(cache.Expirer); ok {
		_ = expirer.Expire(ctx, "key", time.Minute)
		_ = expirer.Persist(ctx, "key")
	}

TTL 在 key 不存在时返回 ErrNotFound；key 存在但永不过期时返回 NoExpiration, nil。

# 快速开始

	import (
		"context"
		"time"

		"github.com/f2xme/gox/cache"
		"github.com/f2xme/gox/cache/adapter/memory"
	)

	func main() {
		c, _ := memory.New()
		defer c.(cache.Closer).Close()

		ctx := context.Background()

		_ = c.Set(ctx, "user:1", []byte("Alice"), 5*time.Minute)

		value, err := c.Get(ctx, "user:1")
		if err == cache.ErrNotFound {
			// 键不存在
		}
		_ = value

		_ = c.Delete(ctx, "user:1")
	}

# 批量操作

	if batch, ok := c.(cache.BatchStore); ok {
		_ = batch.SetMany(ctx, map[string][]byte{
			"user:1": []byte("Alice"),
			"user:2": []byte("Bob"),
		}, time.Minute)

		users, _ := batch.GetMany(ctx, []string{"user:1", "user:2"})
		exists, _ := batch.ExistsMany(ctx, []string{"user:1", "user:2"})
		_ = users
		_ = exists
	}

# 条件写入

	if conditional, ok := c.(cache.ConditionalStore); ok {
		ok, _ := conditional.SetNX(ctx, "job:1", []byte("queued"), time.Minute)
		_ = ok

		old, _ := conditional.Swap(ctx, "job:1", []byte("done"), cache.KeepTTL)
		_ = old
	}

# 类型安全包装器

Typed 自动处理序列化、反序列化和 cache-aside 加载：

	type User struct {
		ID   int
		Name string
	}

	typed := cache.NewTyped[User](c)

	_ = typed.Set(ctx, "user:1", User{ID: 1, Name: "Alice"}, time.Minute)

	user, err := typed.GetOrLoad(ctx, "user:1", time.Minute, func(ctx context.Context) (User, error) {
		return loadUser(ctx, 1)
	})
	_ = user
	_ = err

默认序列化器是 JSON。可以用 WithSerializer 配置 Gob 或自定义序列化器。GetOrLoad 默认会返回缓存写入错误；如果应用更重视可用性，可以使用 WithIgnoreSetErrors。

# 适配器

memory 适合单进程本地缓存，支持 LRU/LFU 淘汰、过期清理、扫描、计数和进程内锁。

redis 适合分布式缓存，支持批量操作、条件写入、扫描、计数和基于 Redis 的锁。

# 错误

包定义了小型错误集合：

  - ErrNotFound：键不存在。
  - ErrLocked：锁已被占用。
  - ErrInvalidTTL：TTL 参数不适用于当前操作。

旧版兼容接口和方法已移除，新代码应使用 Store、BatchStore、Expirer、ConditionalStore、Many、Swap 和 GetOrLoad。
*/
package cache
