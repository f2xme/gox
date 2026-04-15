package jwt

import (
	"time"
)

// Options 定义 JWT 配置选项
type Options struct {
	// Issuer 令牌签发者
	Issuer string
	// Expiration 令牌默认过期时间
	Expiration time.Duration
	// Audience 令牌受众
	Audience []string
	// Revoker 令牌撤销器，用于检查令牌是否被撤销
	Revoker Revoker
}

// Option 定义配置选项函数
type Option func(*Options)

// defaultOptions 返回默认配置
func defaultOptions() Options {
	return Options{
		Expiration: 24 * time.Hour, // 默认 24 小时过期
	}
}

// WithIssuer 设置令牌签发者
//
// 示例：
//
//	NewHS256(secret, WithIssuer("my-app"))
func WithIssuer(issuer string) Option {
	return func(o *Options) {
		o.Issuer = issuer
	}
}

// WithExpiration 设置令牌默认过期时间
//
// 示例：
//
//	NewHS256(secret, WithExpiration(2*time.Hour))
func WithExpiration(expiration time.Duration) Option {
	return func(o *Options) {
		o.Expiration = expiration
	}
}

// WithAudience 设置令牌受众
//
// 示例：
//
//	NewHS256(secret, WithAudience("web", "mobile"))
func WithAudience(audience ...string) Option {
	return func(o *Options) {
		o.Audience = audience
	}
}

// WithRevoker 设置令牌撤销器
//
// 示例：
//
//	revoker := NewRedisRevoker(redisClient)
//	NewHS256(secret, WithRevoker(revoker))
func WithRevoker(revoker Revoker) Option {
	return func(o *Options) {
		o.Revoker = revoker
	}
}
