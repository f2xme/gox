package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const revokedKeySuffix = "revoked:"

// redisRevoker 使用 Redis 实现 Revoker 接口
type redisRevoker struct {
	client *redis.Client
	prefix string
}

// RevokerOption 定义配置 redisRevoker 的函数
type RevokerOption func(*redisRevoker)

// WithPrefix 设置已撤销令牌的键前缀
func WithPrefix(prefix string) RevokerOption {
	return func(r *redisRevoker) {
		r.prefix = prefix
	}
}

// NewRedisRevoker 创建基于 Redis 的令牌撤销器
func NewRedisRevoker(client *redis.Client, opts ...RevokerOption) Revoker {
	r := &redisRevoker{
		client: client,
		prefix: "jwt:",
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *redisRevoker) revokedKey(tokenID string) string {
	return r.prefix + revokedKeySuffix + tokenID
}

// Revoke 将令牌标记为已撤销
func (r *redisRevoker) Revoke(ctx context.Context, tokenID string, expiration time.Duration) error {
	if tokenID == "" {
		return ErrEmptyTokenID
	}

	key := r.revokedKey(tokenID)

	err := r.client.Set(ctx, key, "1", expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

// IsRevoked 检查令牌是否已被撤销
func (r *redisRevoker) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	if tokenID == "" {
		return false, nil
	}

	key := r.revokedKey(tokenID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check revocation: %w", err)
	}

	return exists > 0, nil
}
