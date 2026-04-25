package memory

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/f2xme/gox/cache"
)

var lockOwnerSeq uint64

// lockEntry 表示一个锁及其状态和过期时间。
type lockEntry struct {
	mu         sync.Mutex
	expiration int64 // Unix 纳秒，0 表示无过期时间
	held       bool  // 锁是否当前被持有
	owner      string
	acquiredAt time.Time
	reentrant  bool
	count      int
}

// isExpired 检查锁是否已过期。
func (l *lockEntry) isExpired() bool {
	if l.expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > l.expiration
}

func nextLockOwner() string {
	return "memory-lock-" + strconv.FormatUint(atomic.AddUint64(&lockOwnerSeq, 1), 10)
}

func lockExpiration(ttl time.Duration) int64 {
	if ttl <= 0 {
		return 0
	}
	return time.Now().Add(ttl).UnixNano()
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

		// 如果错误不是 ErrLocked，返回错误
		if err != cache.ErrLocked {
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
// 如果锁已被持有则返回 ErrLocked。
func (c *memCache) TryLock(ctx context.Context, key string, ttl time.Duration) (func() error, error) {
	unlock, _, err := c.tryLock(ctx, key, nextLockOwner(), ttl, false)
	return unlock, err
}

func (c *memCache) tryLock(ctx context.Context, key, owner string, ttl time.Duration, reentrant bool) (func() error, string, error) {
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
			held:       true,
			owner:      owner,
			acquiredAt: time.Now(),
			reentrant:  reentrant,
			count:      1,
			expiration: lockExpiration(ttl),
		}
		c.locks[key] = entry
		c.mu.Unlock()

		return c.unlockFunc(key, owner), owner, nil
	}

	// 锁存在，检查其状态
	entry.mu.Lock()

	if entry.isExpired() {
		// 锁已过期，重置状态
		entry.held = false
		entry.owner = ""
		entry.count = 0
		entry.reentrant = false
	}

	if entry.held {
		if reentrant && entry.reentrant && entry.owner == owner {
			entry.count++
			entry.expiration = lockExpiration(ttl)
			entry.mu.Unlock()
			c.mu.Unlock()
			return c.unlockFunc(key, owner), owner, nil
		}

		// 锁被持有，无法获取
		entry.mu.Unlock()
		c.mu.Unlock()
		return nil, "", cache.ErrLocked
	}

	// 获取锁
	entry.held = true
	entry.owner = owner
	entry.acquiredAt = time.Now()
	entry.reentrant = reentrant
	entry.count = 1
	entry.expiration = lockExpiration(ttl)
	entry.mu.Unlock()
	c.mu.Unlock()

	return c.unlockFunc(key, owner), owner, nil
}

func (c *memCache) unlockFunc(key, owner string) func() error {
	var once sync.Once
	return func() error {
		once.Do(func() {
			c.unlockOwner(key, owner)
		})
		return nil
	}
}

func (c *memCache) unlockOwner(key, owner string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, exists := c.locks[key]; exists {
		e.mu.Lock()
		defer e.mu.Unlock()
		if e.held && e.owner == owner {
			e.count--
			if e.count <= 0 {
				e.held = false
				e.owner = ""
				e.reentrant = false
				e.count = 0
			}
		}
	}
}

// LockWithRenewal 获取带自动续期的锁。
func (c *memCache) LockWithRenewal(ctx context.Context, key string, ttl, renewInterval time.Duration) (func() error, error) {
	for {
		unlock, err := c.TryLockWithRenewal(ctx, key, ttl, renewInterval)
		if err == nil {
			return unlock, nil
		}
		if err != cache.ErrLocked {
			return nil, err
		}

		select {
		case <-time.After(10 * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// TryLockWithRenewal 尝试获取带自动续期的锁。
func (c *memCache) TryLockWithRenewal(ctx context.Context, key string, ttl, renewInterval time.Duration) (func() error, error) {
	unlock, owner, err := c.tryLock(ctx, key, nextLockOwner(), ttl, false)
	if err != nil {
		return nil, err
	}

	if renewInterval <= 0 {
		return unlock, nil
	}

	done := make(chan struct{})
	var once sync.Once
	go func() {
		ticker := time.NewTicker(renewInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.renewLock(key, owner, ttl)
			case <-done:
				return
			}
		}
	}()

	return func() error {
		once.Do(func() {
			close(done)
			_ = unlock()
		})
		return nil
	}, nil
}

func (c *memCache) renewLock(key, owner string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.locks[key]
	if !exists {
		return
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()
	if entry.held && entry.owner == owner {
		entry.expiration = lockExpiration(ttl)
	}
}

// LockReentrant 获取可重入锁。
func (c *memCache) LockReentrant(ctx context.Context, key string, ownerID string, ttl time.Duration) (func() error, error) {
	for {
		unlock, err := c.TryLockReentrant(ctx, key, ownerID, ttl)
		if err == nil {
			return unlock, nil
		}
		if err != cache.ErrLocked {
			return nil, err
		}

		select {
		case <-time.After(10 * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// TryLockReentrant 尝试获取可重入锁。
func (c *memCache) TryLockReentrant(ctx context.Context, key string, ownerID string, ttl time.Duration) (func() error, error) {
	if ownerID == "" {
		ownerID = nextLockOwner()
	}
	unlock, _, err := c.tryLock(ctx, key, ownerID, ttl, true)
	return unlock, err
}

// GetLockInfo 查询锁的当前状态。
func (c *memCache) GetLockInfo(ctx context.Context, key string) (cache.LockInfo, error) {
	c.mu.RLock()
	entry, exists := c.locks[key]
	c.mu.RUnlock()
	if !exists {
		return cache.LockInfo{}, cache.ErrNotFound
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()
	if !entry.held || entry.isExpired() {
		return cache.LockInfo{}, cache.ErrNotFound
	}

	var ttl time.Duration
	if entry.expiration > 0 {
		ttl = time.Duration(entry.expiration - time.Now().UnixNano())
		if ttl < 0 {
			return cache.LockInfo{}, cache.ErrNotFound
		}
	}

	return cache.LockInfo{
		Owner:      entry.owner,
		AcquiredAt: entry.acquiredAt,
		TTL:        ttl,
		Reentrant:  entry.reentrant,
		Count:      entry.count,
	}, nil
}
