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
	mu     sync.RWMutex
	items  map[string]*item
	opts   Options
	stopCh chan struct{}
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

	s.items[id] = &item{
		answer:     answer,
		expiration: expiration,
	}
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

// Delete 删除验证码。
func (s *memoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, id)
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
	close(s.stopCh)
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
			delete(s.items, key)
		}
	}
}
