package cache

import (
	"context"
	"time"

	"github.com/f2xme/gox/cache"
	"github.com/f2xme/gox/captcha"
)

// cacheStore 基于 gox/cache 包的存储实现。
type cacheStore struct {
	cache Backend
	opts  Options
}

// Backend 定义 cache 适配器需要的最小缓存后端能力。
type Backend interface {
	// Get 获取指定键的值。
	Get(ctx context.Context, key string) ([]byte, error)

	// Set 使用指定的键和 TTL 存储值。
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete 从缓存中删除键。
	Delete(ctx context.Context, key string) error

	// Exists 检查键是否存在。
	Exists(ctx context.Context, key string) (bool, error)
}

// Set 存储验证码答案。
func (s *cacheStore) Set(ctx context.Context, id string, answer string, ttl time.Duration) error {
	key := s.opts.Prefix + id
	if ttl == 0 {
		ttl = s.opts.TTL
	}
	return s.cache.Set(ctx, key, []byte(answer), ttl)
}

// Get 获取验证码答案。
func (s *cacheStore) Get(ctx context.Context, id string) (string, error) {
	key := s.opts.Prefix + id
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		if err == cache.ErrNotFound {
			return "", captcha.ErrNotFound
		}
		return "", err
	}
	return string(data), nil
}

// Delete 删除验证码。
func (s *cacheStore) Delete(ctx context.Context, id string) error {
	key := s.opts.Prefix + id
	return s.cache.Delete(ctx, key)
}

// Exists 检查验证码是否存在。
func (s *cacheStore) Exists(ctx context.Context, id string) (bool, error) {
	key := s.opts.Prefix + id
	return s.cache.Exists(ctx, key)
}
