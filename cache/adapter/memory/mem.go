package memory

import (
	"context"
	"encoding/binary"
	"math"
	"path"
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

func expirationFromTTL(ttl time.Duration) (int64, error) {
	if ttl == cache.KeepTTL {
		return 0, cache.ErrInvalidTTL
	}
	if ttl < cache.KeepTTL {
		return 0, cache.ErrInvalidTTL
	}
	if ttl == cache.NoExpiration {
		return 0, nil
	}
	return time.Now().Add(ttl).UnixNano(), nil
}

func copyBytes(value []byte) []byte {
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	return valueCopy
}

func (c *memCache) evictBeforeNewKeyLocked(key string) {
	if c.cfg.MaxSize <= 0 || c.eviction == nil {
		return
	}

	if itm, exists := c.items[key]; exists {
		if !itm.isExpired() {
			return
		}
		delete(c.items, key)
		c.eviction.remove(key)
	}

	for k, itm := range c.items {
		if itm.isExpired() {
			delete(c.items, k)
			c.eviction.remove(k)
		}
	}

	if len(c.items) < c.cfg.MaxSize {
		return
	}

	victim := c.eviction.selectVictim()
	if victim != "" {
		delete(c.items, victim)
		c.eviction.remove(victim)
	}
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
	expiration, err := expirationFromTTL(ttl)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.evictBeforeNewKeyLocked(key)

	c.items[key] = &item{
		value:      copyBytes(value),
		expiration: expiration,
	}

	// 更新淘汰策略
	if c.eviction != nil {
		c.eviction.onSet(key)
	}

	return nil
}

// GetMany 批量获取多个键。
func (c *memCache) GetMany(ctx context.Context, keys []string) (map[string][]byte, error) {
	result := make(map[string][]byte, len(keys))
	for _, key := range keys {
		value, err := c.Get(ctx, key)
		if err != nil {
			if err == cache.ErrNotFound {
				continue
			}
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

// SetMany 使用相同 TTL 批量设置多个键。
func (c *memCache) SetMany(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	for key, value := range items {
		if err := c.Set(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

// DeleteMany 批量删除多个键。
func (c *memCache) DeleteMany(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := c.Delete(ctx, key); err != nil {
			return err
		}
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

// Incr 原子性地增加键的值，并返回增加后的值。
// 如果键不存在，则初始化为 0 后再增加。
func (c *memCache) Incr(ctx context.Context, key string, delta int64) (int64, error) {
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

	if !isValid {
		c.evictBeforeNewKeyLocked(key)
	}

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

// IncrFloat 原子性地增加键的浮点值，并返回增加后的值。
// 如果键不存在，则初始化为 0.0 后再增加。
func (c *memCache) IncrFloat(ctx context.Context, key string, delta float64) (float64, error) {
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

	if !isValid {
		c.evictBeforeNewKeyLocked(key)
	}

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

// TTL 实现对应能力接口。
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

// SetNX 实现对应能力接口。
func (c *memCache) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	expiration, err := expirationFromTTL(ttl)
	if err != nil {
		return false, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	itm, exists := c.items[key]
	if exists && !itm.isExpired() {
		return false, nil // 键已存在
	}

	c.evictBeforeNewKeyLocked(key)

	c.items[key] = &item{
		value:      copyBytes(value),
		expiration: expiration,
	}

	if c.eviction != nil {
		c.eviction.onSet(key)
	}

	return true, nil
}

// SetXX 实现对应能力接口。
func (c *memCache) SetXX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	itm, exists := c.items[key]
	if !exists || itm.isExpired() {
		return false, nil // 键不存在
	}

	expiration := itm.expiration
	if ttl != cache.KeepTTL {
		var err error
		expiration, err = expirationFromTTL(ttl)
		if err != nil {
			return false, err
		}
	}

	c.items[key] = &item{
		value:      copyBytes(value),
		expiration: expiration,
	}

	if c.eviction != nil {
		c.eviction.onAccess(key)
	}

	return true, nil
}

// Swap 原子性地获取旧值并设置新值。
func (c *memCache) Swap(ctx context.Context, key string, value []byte, ttl time.Duration) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	itm, exists := c.items[key]
	if !exists || itm.isExpired() {
		return nil, cache.ErrNotFound
	}

	oldValue := make([]byte, len(itm.value))
	copy(oldValue, itm.value)

	expiration := itm.expiration
	if ttl != cache.KeepTTL {
		var err error
		expiration, err = expirationFromTTL(ttl)
		if err != nil {
			return nil, err
		}
	}

	c.items[key] = &item{
		value:      copyBytes(value),
		expiration: expiration,
	}

	if c.eviction != nil {
		c.eviction.onAccess(key)
	}

	return oldValue, nil
}

// Expire 实现对应能力接口。
func (c *memCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if ttl == cache.KeepTTL || ttl < cache.KeepTTL {
		return cache.ErrInvalidTTL
	}

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

// Persist 移除键的过期时间。
func (c *memCache) Persist(ctx context.Context, key string) error {
	return c.Expire(ctx, key, cache.NoExpiration)
}

// ExistsMany 批量检查键是否存在。
func (c *memCache) ExistsMany(ctx context.Context, keys []string) (map[string]bool, error) {
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
		matched, err := path.Match(pattern, key)
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
