// Package cache 提供统一的缓存操作抽象层。
// 支持内存缓存和分布式缓存，提供类型安全的操作。
package cache

import (
	"context"
	"time"
)

// Cache 定义基础缓存操作接口。
// 所有实现必须支持字节级操作并支持 context。
type Cache interface {
	// Get 获取指定键的值。
	// 如果键不存在则返回 ErrNotFound。
	Get(ctx context.Context, key string) ([]byte, error)

	// Set 使用指定的键和 TTL 存储值。
	// TTL 为 0 表示永不过期。
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete 从缓存中删除键。
	// 如果键不存在不会返回错误。
	Delete(ctx context.Context, key string) error

	// Exists 检查键是否存在于缓存中。
	Exists(ctx context.Context, key string) (bool, error)
}

// MultiCache 扩展 Cache 接口，提供批量操作以提高性能。
type MultiCache interface {
	Cache

	// GetMulti 在单次操作中获取多个键。
	// 不存在的键不会包含在返回的 map 中。
	GetMulti(ctx context.Context, keys []string) (map[string][]byte, error)

	// SetMulti 使用相同的 TTL 存储多个键值对。
	// TTL 为 0 表示永不过期。
	SetMulti(ctx context.Context, items map[string][]byte, ttl time.Duration) error

	// DeleteMulti 在单次操作中删除多个键。
	DeleteMulti(ctx context.Context, keys []string) error
}

// MultiCacheV2 扩展 MultiCache 接口，提供更多批量操作。
type MultiCacheV2 interface {
	MultiCache

	// ExistsMulti 批量检查键是否存在。
	// 返回 map 中每个键对应其存在状态。
	ExistsMulti(ctx context.Context, keys []string) (map[string]bool, error)
}

// Locker 提供分布式锁功能。
type Locker interface {
	// Lock 为指定的键获取锁，使用指定的 TTL。
	// 阻塞直到获取锁或 context 被取消。
	// 返回一个必须调用以释放锁的 unlock 函数。
	Lock(ctx context.Context, key string, ttl time.Duration) (unlock func() error, err error)

	// TryLock 尝试为指定的键获取锁，使用指定的 TTL。
	// 如果锁已被持有则立即返回 ErrLockFailed。
	// 返回一个必须调用以释放锁的 unlock 函数。
	TryLock(ctx context.Context, key string, ttl time.Duration) (unlock func() error, err error)
}

// LockerV2 扩展 Locker 接口，提供自动续期和可重入锁。
type LockerV2 interface {
	Locker

	// LockWithRenewal 获取带自动续期的锁。
	// renewInterval 指定续期间隔，通常设置为 ttl 的 1/3 到 1/2。
	// 返回的 unlock 函数会自动停止续期并释放锁。
	LockWithRenewal(ctx context.Context, key string, ttl, renewInterval time.Duration) (unlock func() error, err error)

	// TryLockWithRenewal 非阻塞版本的 LockWithRenewal。
	// 如果锁已被持有则立即返回 ErrLockFailed。
	TryLockWithRenewal(ctx context.Context, key string, ttl, renewInterval time.Duration) (unlock func() error, err error)

	// LockReentrant 获取可重入锁。
	// ownerID 是调用者的唯一标识（如 requestID, traceID 等）。
	// 同一个 ownerID 可以多次获取同一把锁，每次获取会增加重入计数。
	// unlock 时会减少计数，计数归零时才真正释放锁。
	LockReentrant(ctx context.Context, key string, ownerID string, ttl time.Duration) (unlock func() error, err error)

	// TryLockReentrant 非阻塞版本的 LockReentrant。
	TryLockReentrant(ctx context.Context, key string, ownerID string, ttl time.Duration) (unlock func() error, err error)
}

// LockInfo 包含锁的元数据。
type LockInfo struct {
	Owner      string        // 锁持有者的标识
	AcquiredAt time.Time     // 锁获取时间
	TTL        time.Duration // 锁的剩余有效期
	Reentrant  bool          // 是否为可重入锁
	Count      int           // 重入计数（仅可重入锁有效）
}

// LockMetadata 提供锁元数据查询功能。
type LockMetadata interface {
	// GetLockInfo 查询锁的当前状态。
	// 如果锁不存在或已过期返回 ErrNotFound。
	GetLockInfo(ctx context.Context, key string) (LockInfo, error)
}

// Closer 为缓存实现提供资源清理功能。
type Closer interface {
	// Close 释放缓存持有的所有资源。
	Close() error
}

// Counter 提供原子计数器操作。
// 适用于需要原子递增/递减的场景，如计数器、限流器等。
type Counter interface {
	// Increment 原子性地增加键的值，并返回增加后的值。
	// 如果键不存在，则初始化为 0 后再增加。
	// delta 可以是正数（递增）或负数（递减）。
	Increment(ctx context.Context, key string, delta int64) (int64, error)

	// IncrementFloat 原子性地增加键的浮点值，并返回增加后的值。
	// 如果键不存在，则初始化为 0.0 后再增加。
	// delta 可以是正数（递增）或负数（递减）。
	IncrementFloat(ctx context.Context, key string, delta float64) (float64, error)
}

// Advanced 提供高级缓存操作。
type Advanced interface {
	// TTL 返回键的剩余过期时间。
	// 如果键不存在返回 ErrNotFound。
	// 如果键没有设置过期时间返回 0, nil。
	TTL(ctx context.Context, key string) (time.Duration, error)

	// SetNX 仅当键不存在时设置值。
	// 返回 true 表示设置成功，false 表示键已存在。
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)

	// SetXX 仅当键存在时更新值。
	// 返回 true 表示更新成功，false 表示键不存在。
	SetXX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)

	// GetSet 原子性地获取旧值并设置新值。
	// 如果键不存在返回 nil, ErrNotFound。
	// ttl 为 0 表示不改变原有过期时间（如果有），-1 表示移除过期时间。
	GetSet(ctx context.Context, key string, value []byte, ttl time.Duration) ([]byte, error)

	// Expire 更新键的过期时间而不修改值。
	// 如果键不存在返回 ErrNotFound。
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

// Scanner 提供键遍历功能。
type Scanner interface {
	// Scan 使用游标迭代匹配 pattern 的键。
	// pattern 支持 glob 模式：* 匹配任意字符，? 匹配单个字符，[abc] 匹配字符集。
	// cursor 为 0 表示开始新的迭代，返回的 cursor 为 0 表示迭代结束。
	// count 是每次迭代的建议返回数量（实际可能更多或更少）。
	Scan(ctx context.Context, pattern string, cursor uint64, count int64) (keys []string, nextCursor uint64, err error)
}
