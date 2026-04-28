package memory

import (
	"context"
	"sync"
	"time"

	"github.com/f2xme/gox/session"
)

// Store 是基于内存的会话存储。
type Store struct {
	mu      sync.RWMutex
	items   map[string]*entry
	opts    Options
	stopCh  chan struct{}
	stopped bool
}

type entry struct {
	session    *session.Session
	expiration time.Time
}

// New 创建内存会话存储。
func New(opts ...Option) (*Store, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	if err := options.validate(); err != nil {
		return nil, err
	}

	store := &Store{
		items:  make(map[string]*entry),
		opts:   options,
		stopCh: make(chan struct{}),
	}
	go store.cleanupLoop()

	return store, nil
}

// Get 根据 ID 获取会话。
func (s *Store) Get(ctx context.Context, id string) (*session.Session, error) {
	s.mu.RLock()
	item, ok := s.items[id]
	if !ok {
		s.mu.RUnlock()
		return nil, session.ErrNotFound
	}
	if item.isExpired(time.Now()) {
		s.mu.RUnlock()
		s.mu.Lock()
		if current, exists := s.items[id]; exists && current.isExpired(time.Now()) {
			delete(s.items, id)
		}
		s.mu.Unlock()
		return nil, session.ErrNotFound
	}
	cloned := cloneSession(item.session)
	s.mu.RUnlock()
	return cloned, nil
}

// Set 保存会话并设置存储层 TTL。
func (s *Store) Set(ctx context.Context, sess *session.Session, ttl time.Duration) error {
	if sess == nil || sess.ID == "" {
		return session.ErrInvalidID
	}
	if ttl <= 0 {
		return session.ErrInvalidTTL
	}

	cloned := cloneSession(sess)

	s.mu.Lock()
	s.items[cloned.ID] = &entry{
		session:    cloned,
		expiration: time.Now().Add(ttl),
	}
	s.mu.Unlock()
	return nil
}

// Delete 删除会话。
func (s *Store) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	delete(s.items, id)
	s.mu.Unlock()
	return nil
}

// Close 停止清理 goroutine。
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		return nil
	}
	close(s.stopCh)
	s.stopped = true
	return nil
}

func (s *Store) cleanupLoop() {
	ticker := time.NewTicker(s.opts.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupExpired()
		case <-s.stopCh:
			return
		}
	}
}

func (s *Store) cleanupExpired() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, item := range s.items {
		if item.isExpired(now) {
			delete(s.items, id)
		}
	}
}

func cloneSession(sess *session.Session) *session.Session {
	if sess == nil {
		return nil
	}

	values := make(map[string]any, len(sess.Values))
	for k, v := range sess.Values {
		values[k] = v
	}

	return &session.Session{
		ID:        sess.ID,
		Values:    values,
		CreatedAt: sess.CreatedAt,
		UpdatedAt: sess.UpdatedAt,
		ExpiresAt: sess.ExpiresAt,
	}
}

func (e *entry) isExpired(now time.Time) bool {
	return e == nil || !e.expiration.After(now)
}
