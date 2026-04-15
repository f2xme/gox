package mem

import (
	"context"
	"sync"
	"time"

	"github.com/f2xme/gox/cache"
)

// lockEntry 表示一个锁及其状态和过期时间。
type lockEntry struct {
	mu         sync.Mutex
	expiration int64 // Unix 纳秒，0 表示无过期时间
	held       bool  // 锁是否当前被持有
}

// isExpired 检查锁是否已过期。
func (l *lockEntry) isExpired() bool {
	if l.expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > l.expiration
}

// Lock 阻塞直到获取锁或 context 被取消。
// 返回一个释放锁的 unlock 函数。
func (c *memCache) Lock(ctx context.Context, key string, ttl time.Duration) (func() error, error) {
	for {
		// 尝试获取锁
		unlock, err := c.TryLock(ctx, key, ttl)
		if err == nil {
			return unlock, nil
		}

		// 如果错误不是 ErrLockFailed，返回错误
		if err != cache.ErrLockFailed {
			return nil, err
		}

		// 检查 context 是否被取消
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// 重试前休眠
		select {
		case <-time.After(10 * time.Millisecond):
			// 继续下一次迭代
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// TryLock 尝试立即获取锁而不阻塞。
// 如果锁已被持有则返回 ErrLockFailed。
func (c *memCache) TryLock(ctx context.Context, key string, ttl time.Duration) (func() error, error) {
	c.mu.Lock()

	// 如果需要，初始化 locks map
	if c.locks == nil {
		c.locks = make(map[string]*lockEntry)
	}

	// 检查锁是否存在
	entry, exists := c.locks[key]
	if !exists {
		// 创建新的锁条目
		entry = &lockEntry{
			held: true,
		}
		if ttl > 0 {
			entry.expiration = time.Now().Add(ttl).UnixNano()
		}
		c.locks[key] = entry
		c.mu.Unlock()

		// 返回只捕获 key 的 unlock 函数
		unlock := func() error {
			c.mu.Lock()
			defer c.mu.Unlock()

			if e, exists := c.locks[key]; exists {
				e.mu.Lock()
				if e.held {
					e.held = false
				}
				e.mu.Unlock()
			}
			return nil
		}

		return unlock, nil
	}

	// 锁存在，检查其状态
	entry.mu.Lock()

	if entry.isExpired() {
		// 锁已过期，重置状态
		entry.held = false
	}

	if entry.held {
		// 锁被持有，无法获取
		entry.mu.Unlock()
		c.mu.Unlock()
		return nil, cache.ErrLockFailed
	}

	// 获取锁
	entry.held = true
	if ttl > 0 {
		entry.expiration = time.Now().Add(ttl).UnixNano()
	} else {
		entry.expiration = 0
	}
	entry.mu.Unlock()
	c.mu.Unlock()

	// 返回只捕获 key 的 unlock 函数
	unlock := func() error {
		c.mu.Lock()
		defer c.mu.Unlock()

		if e, exists := c.locks[key]; exists {
			e.mu.Lock()
			if e.held {
				e.held = false
			}
			e.mu.Unlock()
		}
		return nil
	}

	return unlock, nil
}
