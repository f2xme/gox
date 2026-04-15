package jwt

import (
	"context"
	"time"
)

// Revoker 定义令牌撤销接口
type Revoker interface {
	// Revoke 将令牌标记为已撤销
	Revoke(ctx context.Context, tokenID string, expiration time.Duration) error

	// IsRevoked 检查令牌是否已被撤销
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
}
