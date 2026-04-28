package redis

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/f2xme/gox/session"
	goredis "github.com/redis/go-redis/v9"
)

// Store 是基于 Redis 的会话存储。
type Store struct {
	client     goredis.UniversalClient
	keyPrefix  string
	ownsClient bool
	mu         sync.Mutex
	closed     bool
}

// Get 根据 ID 获取会话。
func (s *Store) Get(ctx context.Context, id string) (*session.Session, error) {
	data, err := s.client.Get(ctx, s.key(id)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, session.ErrNotFound
		}
		return nil, err
	}

	var sess session.Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	if sess.Values == nil {
		sess.Values = make(map[string]any)
	}
	return &sess, nil
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

	data, err := json.Marshal(cloned)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, s.key(cloned.ID), data, ttl).Err()
}

// Delete 删除会话。
func (s *Store) Delete(ctx context.Context, id string) error {
	return s.client.Del(ctx, s.key(id)).Err()
}

// Close 关闭由 Store 内部创建的 Redis 客户端。
// 如果客户端由 WithClient 传入，则 Close 不会关闭调用方拥有的客户端。
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	if !s.ownsClient {
		s.closed = true
		return nil
	}
	if err := s.client.Close(); err != nil {
		return err
	}
	s.closed = true
	return nil
}

func (s *Store) key(id string) string {
	return s.keyPrefix + id
}

func cloneSession(sess *session.Session) *session.Session {
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
