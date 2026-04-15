// Package redis 提供基于 Redis 的缓存实现。
package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/f2xme/gox/cache"
	"github.com/redis/go-redis/v9"
)

// unlockScript 是一个 Lua 脚本，在删除锁之前原子性地检查锁值是否匹配。
// 这可以防止删除其他人的锁。
const unlockScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
`

// redisCache 使用 Redis 实现 cache.Cache、cache.MultiCache、cache.Locker 和 cache.Closer。
type redisCache struct {
	client redis.UniversalClient
}

// Get 获取给定键的值。
// 如果键不存在则返回 cache.ErrNotFound。
func (r *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, cache.ErrNotFound
		}
		return nil, err
	}
	return []byte(val), nil
}

// Set 使用给定的键和 TTL 存储值。
// TTL 为 0 表示无过期时间。
func (r *redisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete 从缓存中删除键。
// 如果键不存在不会返回错误。
func (r *redisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists 检查键是否存在于缓存中。
func (r *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// GetMulti 在单次操作中获取多个键。
// 不存在的键不会包含在返回的 map 中。
func (r *redisCache) GetMulti(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return make(map[string][]byte), nil
	}

	pipe := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	result := make(map[string][]byte)
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue // 跳过不存在的键
			}
			return nil, err
		}
		result[keys[i]] = []byte(val)
	}

	return result, nil
}

// SetMulti 使用相同的 TTL 存储多个键值对。
// TTL 为 0 表示无过期时间。
func (r *redisCache) SetMulti(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()

	for key, value := range items {
		pipe.Set(ctx, key, value, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// DeleteMulti 在单次操作中删除多个键。
func (r *redisCache) DeleteMulti(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()

	for _, key := range keys {
		pipe.Del(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Close 释放缓存持有的所有资源。
func (r *redisCache) Close() error {
	return r.client.Close()
}

// TryLock 尝试为给定的键获取锁，使用指定的 TTL。
// 如果锁已被持有则立即返回 cache.ErrLockFailed。
//
// 重要：锁将在 TTL 持续时间后自动过期。
// 对于长时间运行的任务，确保 TTL 长于任务持续时间，
// 或单独实现锁续期机制。
//
// 返回一个必须调用以释放锁的 unlock 函数。
func (r *redisCache) TryLock(ctx context.Context, key string, ttl time.Duration) (func() error, error) {
	// 为此锁生成唯一令牌
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	// 使用 SET NX（如果不存在则设置）尝试获取锁
	ok, err := r.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, cache.ErrLockFailed
	}

	// 返回 unlock 函数
	unlock := func() error {
		// 使用 Lua 脚本原子性地检查和删除
		result, err := r.client.Eval(ctx, unlockScript, []string{key}, token).Result()
		if err != nil {
			return err
		}
		// result 为 1 表示已删除，0 表示未删除（已过期或不同令牌）
		_ = result
		return nil
	}

	return unlock, nil
}

const (
	lockRetryInitialBackoff = 10 * time.Millisecond
	lockRetryMaxBackoff     = 100 * time.Millisecond
)

// Lock 为给定的键获取锁，使用指定的 TTL。
// 阻塞直到获取锁或 context 被取消。
//
// 重要：锁将在 TTL 持续时间后自动过期。
// 对于长时间运行的任务，确保 TTL 长于任务持续时间，
// 或单独实现锁续期机制。
//
// 返回一个必须调用以释放锁的 unlock 函数。
func (r *redisCache) Lock(ctx context.Context, key string, ttl time.Duration) (func() error, error) {
	backoff := lockRetryInitialBackoff

	for {
		unlock, err := r.TryLock(ctx, key, ttl)
		if err == nil {
			return unlock, nil
		}

		if err != cache.ErrLockFailed {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
			if backoff > lockRetryMaxBackoff {
				backoff = lockRetryMaxBackoff
			}
		}
	}
}

// generateToken 为锁所有权生成一个唯一的随机令牌。
func generateToken() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
