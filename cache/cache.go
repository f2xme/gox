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

// Closer 为缓存实现提供资源清理功能。
type Closer interface {
	// Close 释放缓存持有的所有资源。
	Close() error
}
