package auth

import (
	"strings"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/middleware/internal"
)

// Options 定义认证中间件配置
type Options struct {
	validator      TokenValidator
	statusChecker  UserStatusChecker
	tokenExtractor func(httpx.Context) string
	skipPaths      map[string]bool // 精确匹配
	skipPatterns   []string        // 通配符模式
	errorHandler   func(httpx.Context)
	banHandler     func(httpx.Context)
}

func defaultOptions() Options {
	return Options{
		tokenExtractor: defaultTokenExtractor,
		errorHandler:   defaultUnauthorizedHandler,
		banHandler:     defaultBannedHandler,
	}
}

func defaultTokenExtractor(ctx httpx.Context) string {
	authHeader := ctx.Header("Authorization")
	if len(authHeader) <= len(BearerPrefix) || authHeader[:len(BearerPrefix)] != BearerPrefix {
		return ""
	}
	return authHeader[len(BearerPrefix):]
}

func defaultUnauthorizedHandler(ctx httpx.Context) {
	internal.JSONError(ctx, 401, "Authentication required")
}

func defaultBannedHandler(ctx httpx.Context) {
	internal.JSONError(ctx, 403, "User is banned")
}

// Option 配置认证中间件
type Option func(*Options)

// WithValidator 设置 token 验证器
func WithValidator(v TokenValidator) Option {
	return func(o *Options) {
		o.validator = v
	}
}

// WithUserStatusChecker 设置实时用户状态检查器
// 此检查器在每次请求时调用，以确保封禁状态立即生效
func WithUserStatusChecker(checker UserStatusChecker) Option {
	return func(o *Options) {
		o.statusChecker = checker
	}
}

// WithTokenExtractor 设置 token 提取函数
func WithTokenExtractor(fn func(httpx.Context) string) Option {
	return func(o *Options) {
		o.tokenExtractor = fn
	}
}

// WithSkipPaths 设置跳过认证的路径
// 支持精确匹配和通配符模式（如 "/public/*"）
func WithSkipPaths(paths ...string) Option {
	return func(o *Options) {
		o.skipPaths = make(map[string]bool)
		o.skipPatterns = nil
		for _, path := range paths {
			if strings.HasSuffix(path, "/*") {
				o.skipPatterns = append(o.skipPatterns, path)
			} else {
				o.skipPaths[path] = true
			}
		}
	}
}

// WithErrorHandler 设置未授权错误处理函数
func WithErrorHandler(fn func(httpx.Context)) Option {
	return func(o *Options) {
		o.errorHandler = fn
	}
}

// WithBanHandler 设置封禁用户错误处理函数
func WithBanHandler(fn func(httpx.Context)) Option {
	return func(o *Options) {
		o.banHandler = fn
	}
}
