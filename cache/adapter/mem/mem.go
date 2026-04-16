package mem

import (
	"context"
	"encoding/binary"
	"math"
	"sync"
	"time"

	"github.com/f2xme/gox/cache"
)

// item 表示一个缓存条目及其值和过期时间。
type item struct {
	value      []byte
	expiration int64 // Unix 纳秒，0 表示无过期时间
}

// isExpired 检查条目是否已过期。
func (i *item) isExpired() bool {
	if i.expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > i.expiration
}

// memCache 是内存缓存实现。
type memCache struct {
	mu       sync.RWMutex
	items    map[string]*item
	locks    map[string]*lockEntry
	cfg      *Options
	eviction evictionPolicy
	stopCh   chan struct{}
	stopped  bool
}

// Get 从缓存中获取值。
// 如果键不存在或已过期则返回 cache.ErrNotFound。
func (c *memCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	item, exists := c.items[key]
	if !exists {
		c.mu.RUnlock()
		return nil, cache.ErrNotFound
	}

	if item.isExpired() {
		c.mu.RUnlock()
		// 升级为写锁
		c.mu.Lock()
		// 获取写锁后重新检查（双重检查模式）
		if item, exists := c.items[key]; exists && item.isExpired() {
			delete(c.items, key)
			if c.eviction != nil {
				c.eviction.remove(key)
			}
		}
		c.mu.Unlock()
		return nil, cache.ErrNotFound
	}

	// 返回副本以防止外部修改
	result := make([]byte, len(item.value))
	copy(result, item.value)

	// 成功获取后更新淘汰策略
	if c.eviction != nil {
		c.eviction.onAccess(key)
	}

	c.mu.RUnlock()
	return result, nil
}

// Set 使用给定的 TTL 在缓存中存储值。
// TTL 为 0 表示无过期时间。
func (c *memCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否需要淘汰
	if c.cfg.MaxSize > 0 && c.eviction != nil {
		_, exists := c.items[key]
		// 仅当键是新的且缓存已满时才淘汰
		if !exists {
			// 首先，清理过期条目以获得准确计数
			for k, item := range c.items {
				if item.isExpired() {
					delete(c.items, k)
					c.eviction.remove(k)
				}
			}

			// 现在检查是否仍需要淘汰
			if len(c.items) >= c.cfg.MaxSize {
				victim := c.eviction.selectVictim()
				if victim != "" {
					delete(c.items, victim)
					c.eviction.remove(victim)
				}
			}
		}
	}

	// 复制以防止外部修改
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	c.items[key] = &item{
		value:      valueCopy,
		expiration: expiration,
	}

	// 更新淘汰策略
	if c.eviction != nil {
		c.eviction.onSet(key)
	}

	return nil
}

// Delete 从缓存中删除键。
func (c *memCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	if c.eviction != nil {
		c.eviction.remove(key)
	}
	return nil
}

// Exists 检查键是否存在于缓存中且未过期。
func (c *memCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	item, exists := c.items[key]
	if !exists {
		c.mu.RUnlock()
		return false, nil
	}

	if item.isExpired() {
		c.mu.RUnlock()
		// 升级为写锁以删除过期条目
		c.mu.Lock()
		// 获取写锁后重新检查条目以避免竞态条件
		if item, exists := c.items[key]; exists && item.isExpired() {
			delete(c.items, key)
			if c.eviction != nil {
				c.eviction.remove(key)
			}
		}
		c.mu.Unlock()
		return false, nil
	}

	c.mu.RUnlock()
	return true, nil
}

// Close 停止清理 goroutine 并释放资源。
// Close 后，缓存数据仍可访问，但过期条目不会自动删除。
func (c *memCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return nil
	}

	c.stopped = true
	close(c.stopCh)
	return nil
}

// cleanupLoop 定期运行以删除过期条目。
func (c *memCache) cleanupLoop() {
	ticker := time.NewTicker(c.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCh:
			return
		}
	}
}


// cleanup 从缓存中删除所有过期条目。
func (c *memCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 清理过期的缓存条目
	for key, item := range c.items {
		if item.isExpired() {
			delete(c.items, key)
			if c.eviction != nil {
				c.eviction.remove(key)
			}
		}
	}

	// 清理过期且未持有的锁
	for key, entry := range c.locks {
		entry.mu.Lock()
		if !entry.held && entry.isExpired() {
			delete(c.locks, key)
		}
		entry.mu.Unlock()
	}
}

// Increment 原子性地增加键的值，并返回增加后的值。
// 如果键不存在，则初始化为 0 后再增加。
func (c *memCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	itm, exists := c.items[key]
	var currentValue int64
	var expiration int64

	isValid := exists && !itm.isExpired()
	if isValid {
		if len(itm.value) == 8 {
			currentValue = int64(binary.BigEndian.Uint64(itm.value))
		}
		expiration = itm.expiration
	}

	newValue := currentValue + delta

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(newValue))

	c.items[key] = &item{
		value:      buf,
		expiration: expiration,
	}

	if c.eviction != nil {
		if isValid {
			c.eviction.onAccess(key)
		} else {
			c.eviction.onSet(key)
		}
	}

	return newValue, nil
}

// IncrementFloat 原子性地增加键的浮点值，并返回增加后的值。
// 如果键不存在，则初始化为 0.0 后再增加。
func (c *memCache) IncrementFloat(ctx context.Context, key string, delta float64) (float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	itm, exists := c.items[key]
	var currentValue float64
	var expiration int64

	isValid := exists && !itm.isExpired()
	if isValid {
		if len(itm.value) == 8 {
			bits := binary.BigEndian.Uint64(itm.value)
			currentValue = math.Float64frombits(bits)
		}
		expiration = itm.expiration
	}

	newValue := currentValue + delta

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, math.Float64bits(newValue))

	c.items[key] = &item{
		value:      buf,
		expiration: expiration,
	}

	if c.eviction != nil {
		if isValid {
			c.eviction.onAccess(key)
		} else {
			c.eviction.onSet(key)
		}
	}

	return newValue, nil
}

