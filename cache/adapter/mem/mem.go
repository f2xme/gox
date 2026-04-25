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

// TTL 实现 cache.Advanced 接口。
func (c *memCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	itm, exists := c.items[key]
	if !exists || itm.isExpired() {
		return 0, cache.ErrNotFound
	}

	if itm.expiration == 0 {
		return 0, nil // 无过期时间
	}

	remaining := time.Duration(itm.expiration - time.Now().UnixNano())
	if remaining < 0 {
		return 0, cache.ErrNotFound
	}
	return remaining, nil
}

// SetNX 实现 cache.Advanced 接口。
func (c *memCache) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	itm, exists := c.items[key]
	if exists && !itm.isExpired() {
		return false, nil // 键已存在
	}

	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	c.items[key] = &item{
		value:      valueCopy,
		expiration: expiration,
	}

	if c.eviction != nil {
		c.eviction.onSet(key)
	}

	return true, nil
}

// SetXX 实现 cache.Advanced 接口。
func (c *memCache) SetXX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	itm, exists := c.items[key]
	if !exists || itm.isExpired() {
		return false, nil // 键不存在
	}

	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	} else if ttl == 0 {
		expiration = itm.expiration // 保持原有 TTL
	}
	// ttl < 0 表示无过期时间

	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	c.items[key] = &item{
		value:      valueCopy,
		expiration: expiration,
	}

	if c.eviction != nil {
		c.eviction.onAccess(key)
	}

	return true, nil
}

// GetSet 实现 cache.Advanced 接口。
func (c *memCache) GetSet(ctx context.Context, key string, value []byte, ttl time.Duration) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	itm, exists := c.items[key]
	if !exists || itm.isExpired() {
		return nil, cache.ErrNotFound
	}

	oldValue := make([]byte, len(itm.value))
	copy(oldValue, itm.value)

	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	} else if ttl == 0 {
		expiration = itm.expiration // 保持原有 TTL
	}
	// ttl < 0 表示无过期时间

	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	c.items[key] = &item{
		value:      valueCopy,
		expiration: expiration,
	}

	if c.eviction != nil {
		c.eviction.onAccess(key)
	}

	return oldValue, nil
}

// Expire 实现 cache.Advanced 接口。
func (c *memCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	itm, exists := c.items[key]
	if !exists || itm.isExpired() {
		return cache.ErrNotFound
	}

	if ttl > 0 {
		itm.expiration = time.Now().Add(ttl).UnixNano()
	} else {
		itm.expiration = 0 // 移除过期时间
	}

	return nil
}

// ExistsMulti 实现 cache.MultiCacheV2 接口。
func (c *memCache) ExistsMulti(ctx context.Context, keys []string) (map[string]bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]bool, len(keys))
	for _, key := range keys {
		itm, exists := c.items[key]
		result[key] = exists && !itm.isExpired()
	}
	return result, nil
}

// Scan 实现 cache.Scanner 接口。
func (c *memCache) Scan(ctx context.Context, pattern string, cursor uint64, count int64) ([]string, uint64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 内存实现简化：一次返回所有匹配的键
	// cursor 只用于标识是否已完成（0=开始/结束）
	if cursor != 0 {
		return nil, 0, nil // 已经完成迭代
	}

	var matches []string
	for key := range c.items {
		itm := c.items[key]
		if itm.isExpired() {
			continue
		}
		matched, err := matchGlob(pattern, key)
		if err != nil {
			return nil, 0, err
		}
		if matched {
			matches = append(matches, key)
		}
		if count > 0 && int64(len(matches)) >= count {
			break
		}
	}
	return matches, 0, nil // cursor=0 表示迭代完成
}

// matchGlob 简单的 glob 模式匹配实现。
func matchGlob(pattern, str string) (bool, error) {
	if pattern == "*" {
		return true, nil
	}
	// 简化实现：支持 * 通配符
	// 生产环境建议使用更完善的 glob 库
	if !containsWildcard(pattern) {
		return pattern == str, nil
	}
	// 基本的前缀/后缀匹配
	if pattern[0] == '*' && pattern[len(pattern)-1] == '*' {
		return containsSubstring(str, pattern[1:len(pattern)-1]), nil
	}
	if pattern[0] == '*' {
		return hasSuffix(str, pattern[1:]), nil
	}
	if pattern[len(pattern)-1] == '*' {
		return hasPrefix(str, pattern[:len(pattern)-1]), nil
	}
	return pattern == str, nil
}
}

func containsWildcard(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '*' || s[i] == '?' {
			return true
		}
	}
	return false
}

func containsSubstring(s, substr string) bool {
	return len(substr) == 0 || indexOfSubstring(s, substr) >= 0
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
