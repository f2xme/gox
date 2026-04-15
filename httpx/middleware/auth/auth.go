package auth

import (
	"strings"

	"github.com/f2xme/gox/httpx"
)

const (
	// ClaimsContextKey 是存储认证声明的上下文键
	ClaimsContextKey = "auth.claims"
	// BearerPrefix 是 Bearer token 认证的前缀
	BearerPrefix = "Bearer "
)

// TokenValidator 验证 bearer token 并返回声明
type TokenValidator interface {
	Validate(token string) (Claims, error)
}

// Claims 表示已认证的 token 声明
type Claims interface {
	GetSubject() string
	Get(key string) (any, bool)
}

// UserStatusChecker 实时检查用户状态（如封禁、禁用）
// 在每次请求的 token 验证后调用，以确保立即生效
type UserStatusChecker interface {
	IsBanned(userID string) (bool, error)
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

			token := o.tokenExtractor(ctx)
			if token == "" || o.validator == nil {
				o.errorHandler(ctx)
				return nil
			}

			claims, err := o.validator.Validate(token)
			if err != nil {
				o.errorHandler(ctx)
				return nil
			}

			ctx.Set(ClaimsContextKey, claims)

			// Real-time user status check (e.g., banned, disabled)
			if o.statusChecker != nil {
				userID := claims.GetSubject()
				banned, err := o.statusChecker.IsBanned(userID)
				if err != nil {
					// Fail-open: allow request if checker unavailable to prevent cascading failures
				} else if banned {
					o.banHandler(ctx)
					return nil
				}
			}

			return next(ctx)
		}
	}
}

func shouldSkip(path string, skipPaths map[string]bool, skipPatterns []string) bool {
	// Fast O(1) exact match check
	if skipPaths[path] {
		return true
	}
	// Fallback to pattern matching for wildcards
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
