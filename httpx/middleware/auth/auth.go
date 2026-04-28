package auth

import (
	"context"
	"strings"

	"github.com/f2xme/gox/httpx"
)

const (
	// ClaimsContextKey 是存储认证声明的上下文键
	ClaimsContextKey = "auth.claims"
	// BearerPrefix 是 Bearer token 认证的前缀
	BearerPrefix = "Bearer "
)

// Validator 验证 token 并返回声明。
type Validator interface {
	Validate(ctx context.Context, token string) (Claims, error)
}

// ValidatorFunc 是实现 Validator 的函数类型，便于以内联函数直接作为验证器使用。
type ValidatorFunc func(ctx context.Context, token string) (Claims, error)

// Validate 调用底层函数完成 token 验证。
func (f ValidatorFunc) Validate(ctx context.Context, token string) (Claims, error) {
	return f(ctx, token)
}

// Claims 表示已认证的 token 声明
type Claims interface {
	GetUID() int64
	Get(key string) (any, bool)
}

// UserChecker 实时检查用户状态（如封禁、禁用）
// 在每次请求的 token 验证后调用，以确保立即生效
type UserChecker interface {
	CheckUser(uid int64) error
}

// New 创建认证中间件
func New(opts ...Option) httpx.Middleware {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	return func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			if shouldSkip(ctx.Path(), o.skipPaths, o.skipPatterns) {
				return next(ctx)
			}

			optional := (len(o.optionalPaths) > 0 || len(o.optionalPatterns) > 0) &&
				shouldSkip(ctx.Path(), o.optionalPaths, o.optionalPatterns)

			token := o.tokenExtractor(ctx)
			if token == "" || o.validator == nil {
				if optional {
					return next(ctx)
				}
				o.errorHandler(ctx)
				return nil
			}

			claims, err := o.validator.Validate(ctx.Request().Context(), token)
			if err != nil {
				if optional {
					return next(ctx)
				}
				o.errorHandler(ctx)
				return nil
			}

			ctx.Set(ClaimsContextKey, claims)

			if o.statusChecker != nil {
				if err = o.statusChecker.CheckUser(claims.GetUID()); err != nil {
					o.checkHandler(ctx, err)
					return nil
				}
			}

			return next(ctx)
		}
	}
}

func shouldSkip(path string, skipPaths map[string]bool, skipPatterns []string) bool {
	if skipPaths[path] {
		return true
	}
	for _, pattern := range skipPatterns {
		if matchPath(pattern, path) {
			return true
		}
	}
	return false
}

func matchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return false
}

// GetClaims 从上下文中获取认证声明
func GetClaims(ctx httpx.Context) Claims {
	if claims, ok := ctx.Get(ClaimsContextKey); ok {
		if typed, ok := claims.(Claims); ok {
			return typed
		}
	}
	return nil
}

// GetUID 从上下文中获取当前用户 ID，未认证时返回 0
func GetUID(ctx httpx.Context) int64 {
	if c := GetClaims(ctx); c != nil {
		return c.GetUID()
	}
	return 0
}
