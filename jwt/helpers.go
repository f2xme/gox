package jwt

import (
	"context"
	"fmt"
	"time"
)

// applyDefaultClaims 将默认配置值应用到声明中（如果尚未设置）
func applyDefaultClaims(claims *Claims, opts *Options) {
	now := time.Now()

	if claims.Issuer == "" && opts.Issuer != "" {
		claims.Issuer = opts.Issuer
	}
	if len(claims.Audience) == 0 && len(opts.Audience) > 0 {
		claims.Audience = opts.Audience
	}
	if claims.IssuedAt.IsZero() {
		claims.IssuedAt = now
	}
	if claims.ExpiresAt.IsZero() && opts.Expiration > 0 {
		claims.ExpiresAt = now.Add(opts.Expiration)
	}
}

// checkRevocation 检查令牌是否已被撤销
func checkRevocation(ctx context.Context, revoker Revoker, tokenID string) error {
	if revoker != nil && tokenID != "" {
		revoked, err := revoker.IsRevoked(ctx, tokenID)
		if err != nil {
			return fmt.Errorf("failed to check revocation: %w", err)
		}
		if revoked {
			return ErrTokenRevoked
		}
	}
	return nil
}

// updateClaimsTimestamp 更新 IssuedAt 和 ExpiresAt 时间戳
func updateClaimsTimestamp(claims *Claims, expiration time.Duration) {
	now := time.Now()
	claims.IssuedAt = now
	claims.ExpiresAt = now.Add(expiration)
}
