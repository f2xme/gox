// Package redis 提供基于 Redis 的缓存实现。
package redis

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
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
	lockRetryBaseBackoff = 10 * time.Millisecond
	lockRetryMaxBackoff  = 100 * time.Millisecond
)

// Lock 为给定的键获取锁，使用指定的 TTL。
// 阻塞直到获取锁或 context 被取消。
//
// 使用 decorrelated jitter 退避策略减少惊群效应。
//
// 重要：锁将在 TTL 持续时间后自动过期。
// 对于长时间运行的任务，确保 TTL 长于任务持续时间，
// 或使用 LockWithRenewal 方法。
//
// 返回一个必须调用以释放锁的 unlock 函数。
func (r *redisCache) Lock(ctx context.Context, key string, ttl time.Duration) (func() error, error) {
	backoff := lockRetryBaseBackoff

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
			// Decorrelated jitter: sleep = min(max, random_between(base, sleep * 3))
			nextBackoff := time.Duration(rand.Int63n(int64(backoff*3-lockRetryBaseBackoff))) + lockRetryBaseBackoff
			backoff = min(nextBackoff, lockRetryMaxBackoff)
		}
	}
}

// Increment 原子性地增加键的值，并返回增加后的值。
// 如果键不存在，则初始化为 0 后再增加。
func (r *redisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return r.client.IncrBy(ctx, key, delta).Result()
}

// IncrementFloat 原子性地增加键的浮点值，并返回增加后的值。
// 如果键不存在，则初始化为 0.0 后再增加。
func (r *redisCache) IncrementFloat(ctx context.Context, key string, delta float64) (float64, error) {
	return r.client.IncrByFloat(ctx, key, delta).Result()
}

// generateToken 为锁所有权生成一个唯一的随机令牌。
func generateToken() (string, error) {
	b := make([]byte, 16)
	_, err := cryptorand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// TTL 实现 cache.Advanced 接口。
func (r *redisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if ttl == -2 { // key 不存在
		return 0, cache.ErrNotFound
	}
	if ttl == -1 { // key 存在但没有过期时间
		return 0, nil
	}
	return ttl, nil
}

// SetNX 实现 cache.Advanced 接口。
func (r *redisCache) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, ttl).Result()
}

// SetXX 实现 cache.Advanced 接口。
func (r *redisCache) SetXX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	return r.client.SetXX(ctx, key, value, ttl).Result()
}

// GetSet 实现 cache.Advanced 接口。
func (r *redisCache) GetSet(ctx context.Context, key string, value []byte, ttl time.Duration) ([]byte, error) {
	// 使用 Lua 脚本实现原子操作
	script := `
		local old = redis.call("GET", KEYS[1])
		if old == false then
			return nil
		end
		if tonumber(ARGV[2]) > 0 then
			redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
		elseif tonumber(ARGV[2]) == -1 then
			redis.call("SET", KEYS[1], ARGV[1])
		else
			redis.call("SET", KEYS[1], ARGV[1], "KEEPTTL")
		end
		return old
	`

	ttlMs := int64(ttl / time.Millisecond)
	result, err := r.client.Eval(ctx, script, []string{key}, value, ttlMs).Result()
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, cache.ErrNotFound
	}
	return []byte(result.(string)), nil
}

// Expire 实现 cache.Advanced 接口。
func (r *redisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	ok, err := r.client.Expire(ctx, key, ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return cache.ErrNotFound
	}
	return nil
}

// ExistsMulti 实现 cache.MultiCacheV2 接口。
func (r *redisCache) ExistsMulti(ctx context.Context, keys []string) (map[string]bool, error) {
	if len(keys) == 0 {
		return make(map[string]bool), nil
	}

	pipe := r.client.Pipeline()
	cmds := make([]*redis.IntCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.Exists(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool, len(keys))
	for i, cmd := range cmds {
		n, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		result[keys[i]] = n > 0
	}

	return result, nil
}

// Scan 实现 cache.Scanner 接口。
func (r *redisCache) Scan(ctx context.Context, pattern string, cursor uint64, count int64) ([]string, uint64, error) {
	keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, count).Result()
	return keys, nextCursor, err
}

// LockWithRenewal 实现 cache.LockerV2 接口。
func (r *redisCache) LockWithRenewal(ctx context.Context, key string, ttl, renewInterval time.Duration) (func() error, error) {
	// 生成唯一 token
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	// 获取锁
	ok, err := r.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, cache.ErrLockFailed
	}

	// 启动续期 goroutine
	stopCh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(renewInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 续期锁
				r.client.Expire(ctx, key, ttl)
			case <-stopCh:
				return
			}
		}
	}()

	// 包装 unlock 函数
	wrappedUnlock := func() error {
		close(stopCh)
		// 使用 Lua 脚本原子性地检查和删除
		_, err := r.client.Eval(ctx, unlockScript, []string{key}, token).Result()
		return err
	}

	return wrappedUnlock, nil
}

// TryLockWithRenewal 实现 cache.LockerV2 接口。
func (r *redisCache) TryLockWithRenewal(ctx context.Context, key string, ttl, renewInterval time.Duration) (func() error, error) {
	return r.LockWithRenewal(ctx, key, ttl, renewInterval)
}

// LockReentrant 实现 cache.LockerV2 接口。
func (r *redisCache) LockReentrant(ctx context.Context, key string, ownerID string, ttl time.Duration) (func() error, error) {
	script := `
		local current = redis.call("HGET", KEYS[1], "owner")
		if current == false then
			redis.call("HSET", KEYS[1], "owner", ARGV[1], "count", 1, "acquired_at", ARGV[3])
			redis.call("PEXPIRE", KEYS[1], ARGV[2])
			return 1
		elseif current == ARGV[1] then
			redis.call("HINCRBY", KEYS[1], "count", 1)
			redis.call("PEXPIRE", KEYS[1], ARGV[2])
			return 1
		else
			return 0
		end
	`

	ttlMs := ttl.Milliseconds()
	acquiredAt := time.Now().Unix()

	for {
		result, err := r.client.Eval(ctx, script, []string{key}, ownerID, ttlMs, acquiredAt).Result()
		if err != nil {
			return nil, err
		}
		if result.(int64) == 1 {
			break
		}

		// 使用 decorrelated jitter 退避
		sleep := time.Duration(lockRetryBaseBackoff.Nanoseconds() +
			int64(float64(lockRetryMaxBackoff.Nanoseconds()-lockRetryBaseBackoff.Nanoseconds()) *
			float64(time.Now().UnixNano()%1000) / 1000.0))

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(sleep):
		}
	}

	unlock := func() error {
		unlockScript := `
			local current = redis.call("HGET", KEYS[1], "owner")
			if current ~= ARGV[1] then
				return 0
			end
			local count = redis.call("HINCRBY", KEYS[1], "count", -1)
			if count <= 0 then
				redis.call("DEL", KEYS[1])
			end
			return 1
		`
		_, err := r.client.Eval(ctx, unlockScript, []string{key}, ownerID).Result()
		return err
	}

	return unlock, nil
}

// TryLockReentrant 实现 cache.LockerV2 接口。
func (r *redisCache) TryLockReentrant(ctx context.Context, key string, ownerID string, ttl time.Duration) (func() error, error) {
	script := `
		local current = redis.call("HGET", KEYS[1], "owner")
		if current == false then
			redis.call("HSET", KEYS[1], "owner", ARGV[1], "count", 1, "acquired_at", ARGV[3])
			redis.call("PEXPIRE", KEYS[1], ARGV[2])
			return 1
		elseif current == ARGV[1] then
			redis.call("HINCRBY", KEYS[1], "count", 1)
			redis.call("PEXPIRE", KEYS[1], ARGV[2])
			return 1
		else
			return 0
		end
	`

	ttlMs := ttl.Milliseconds()
	acquiredAt := time.Now().Unix()

	result, err := r.client.Eval(ctx, script, []string{key}, ownerID, ttlMs, acquiredAt).Result()
	if err != nil {
		return nil, err
	}
	if result.(int64) == 0 {
		return nil, cache.ErrLockFailed
	}

	unlock := func() error {
		unlockScript := `
			local current = redis.call("HGET", KEYS[1], "owner")
			if current ~= ARGV[1] then
				return 0
			end
			local count = redis.call("HINCRBY", KEYS[1], "count", -1)
			if count <= 0 then
				redis.call("DEL", KEYS[1])
			end
			return 1
		`
		_, err := r.client.Eval(ctx, unlockScript, []string{key}, ownerID).Result()
		return err
	}

	return unlock, nil
}

// GetLockInfo 实现 cache.LockMetadata 接口。
func (r *redisCache) GetLockInfo(ctx context.Context, key string) (cache.LockInfo, error) {
	result, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return cache.LockInfo{}, err
	}
	if len(result) == 0 {
		return cache.LockInfo{}, cache.ErrNotFound
	}

	info := cache.LockInfo{
		Owner:     result["owner"],
		Reentrant: result["count"] != "",
	}

	if acquiredAtStr, ok := result["acquired_at"]; ok {
		var acquiredAt int64
		if _, err := fmt.Sscanf(acquiredAtStr, "%d", &acquiredAt); err == nil {
			info.AcquiredAt = time.Unix(acquiredAt, 0)
		}
	}

	if countStr, ok := result["count"]; ok {
		fmt.Sscanf(countStr, "%d", &info.Count)
	}

	// 获取 TTL
	ttl, err := r.client.TTL(ctx, key).Result()
	if err == nil && ttl > 0 {
		info.TTL = ttl
	}

	return info, nil
}
