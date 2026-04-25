package memory

import (
	"context"
	"sync"
	"time"

	"github.com/f2xme/gox/captcha"
)

// item 表示一个缓存条目及其答案和过期时间。
type item struct {
	answer     string
	expiration int64 // Unix 纳秒，0 表示无过期时间
}

// isExpired 检查条目是否已过期。
func (i *item) isExpired() bool {
	if i.expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > i.expiration
}

// memoryStore 是内存存储实现。
type memoryStore struct {
	mu        sync.RWMutex
	items     map[string]*item
	order     []string
	opts      Options
	stopCh    chan struct{}
	closeOnce sync.Once
}

// Set 存储验证码答案。
func (s *memoryStore) Set(ctx context.Context, id string, answer string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 使用传入的 TTL，如果为 0 则使用默认值
	if ttl == 0 {
		ttl = s.opts.TTL
	}

	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	if _, exists := s.items[id]; !exists {
		s.order = append(s.order, id)
	}

	s.items[id] = &item{
		answer:     answer,
		expiration: expiration,
	}
	s.enforceMaxSize()
	return nil
}

// Get 获取验证码答案。
func (s *memoryStore) Get(ctx context.Context, id string) (string, error) {
	s.mu.RLock()
	item, exists := s.items[id]
	if !exists {
		s.mu.RUnlock()
		return "", captcha.ErrNotFound
	}

	if item.isExpired() {
		s.mu.RUnlock()
		// 升级为写锁
		s.mu.Lock()
		// 获取写锁后重新检查（双重检查模式）
		if item, exists := s.items[id]; exists && item.isExpired() {
			delete(s.items, id)
		}
		s.mu.Unlock()
		return "", captcha.ErrNotFound
	}

	answer := item.answer
	s.mu.RUnlock()
	return answer, nil
}

// Take 获取并删除验证码答案。
func (s *memoryStore) Take(ctx context.Context, id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.items[id]
	if !exists {
		return "", captcha.ErrNotFound
	}
	if item.isExpired() {
		s.deleteLocked(id)
		return "", captcha.ErrNotFound
	}

	answer := item.answer
	s.deleteLocked(id)
	return answer, nil
}

// Delete 删除验证码。
func (s *memoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteLocked(id)
	return nil
}

// Exists 检查验证码是否存在且未过期。
func (s *memoryStore) Exists(ctx context.Context, id string) (bool, error) {
	s.mu.RLock()
	item, exists := s.items[id]
	if !exists {
		s.mu.RUnlock()
		return false, nil
	}

	if item.isExpired() {
		s.mu.RUnlock()
		// 升级为写锁以删除过期条目
		s.mu.Lock()
		// 获取写锁后重新检查条目以避免竞态条件
		if item, exists := s.items[id]; exists && item.isExpired() {
			delete(s.items, id)
		}
		s.mu.Unlock()
		return false, nil
	}

	s.mu.RUnlock()
	return true, nil
}

// Close 停止清理 goroutine 并释放资源。
func (s *memoryStore) Close() error {
	s.closeOnce.Do(func() {
		close(s.stopCh)
	})
	return nil
}

// cleanupLoop 定期运行以删除过期条目。
func (s *memoryStore) cleanupLoop() {
	ticker := time.NewTicker(s.opts.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.stopCh:
			return
		}
	}
}

// cleanup 从缓存中删除所有过期条目。
func (s *memoryStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, item := range s.items {
		if item.isExpired() {
			s.deleteLocked(key)
		}
	}
}

// deleteLocked 删除条目，调用方必须持有写锁。
func (s *memoryStore) deleteLocked(id string) {
	delete(s.items, id)
	for i, key := range s.order {
		if key == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			return
		}
	}
}

// enforceMaxSize 按写入顺序淘汰超出容量限制的条目，调用方必须持有写锁。
func (s *memoryStore) enforceMaxSize() {
	if s.opts.MaxSize <= 0 {
		return
	}
	for len(s.items) > s.opts.MaxSize && len(s.order) > 0 {
		oldest := s.order[0]
		s.order = s.order[1:]
		delete(s.items, oldest)
	}
}
