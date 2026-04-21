package auth

import (
	"strings"

	"github.com/f2xme/gox/httpx"
)

// Options 定义认证中间件配置
type Options struct {
	validator       TokenValidator
	statusChecker   UserChecker
	checkHandler    func(httpx.Context, error)
	tokenExtractor  func(httpx.Context) string
	skipPaths       map[string]bool // 精确匹配
	skipPatterns    []string        // 通配符模式
	optionalPaths   map[string]bool // 可选认证：有 token 则验证，无 token 则放行
	optionalPatterns []string        // 可选认证通配符模式
	errorHandler    func(httpx.Context)
}

func defaultOptions() Options {
	return Options{
		tokenExtractor: defaultTokenExtractor,
		errorHandler:   defaultErrorHandler,
		checkHandler:   defaultCheckerHandler,
	}
}

func defaultTokenExtractor(ctx httpx.Context) string {
	authHeader := ctx.Header("Authorization").String()
	if len(authHeader) <= len(BearerPrefix) || authHeader[:len(BearerPrefix)] != BearerPrefix {
		return ""
	}
	return authHeader[len(BearerPrefix):]
}

func defaultErrorHandler(ctx httpx.Context) {
	ctx.Unauthorized("Authentication required")
}

func defaultCheckerHandler(ctx httpx.Context, _ error) {
	ctx.Forbidden()
}

// Option 配置认证中间件
type Option func(*Options)

// WithValidator 设置 token 验证器
func WithValidator(v TokenValidator) Option {
	return func(o *Options) {
		o.validator = v
	}
}

// WithUserChecker 设置实时用户状态检查器
// 此检查器在每次请求时调用，以确保用户状态变更立即生效
func WithUserChecker(checker UserChecker) Option {
	return func(o *Options) {
		o.statusChecker = checker
	}
}

// WithCheckHandler 设置用户检查失败时的处理函数
// 若设置了 WithUserChecker，则此选项必须同时设置
func WithCheckHandler(fn func(httpx.Context, error)) Option {
	return func(o *Options) {
		o.checkHandler = fn
	}
}

// WithTokenExtractor 设置 token 提取函数
func WithTokenExtractor(fn func(httpx.Context) string) Option {
	return func(o *Options) {
		o.tokenExtractor = fn
	}
}

func parsePaths(paths []string) (map[string]bool, []string) {
	exact := make(map[string]bool)
	var patterns []string
	for _, p := range paths {
		if strings.HasSuffix(p, "/*") {
			patterns = append(patterns, p)
		} else {
			exact[p] = true
		}
	}
	return exact, patterns
}

// WithSkipPaths 设置跳过认证的路径
// 支持精确匹配和通配符模式（如 "/public/*"）
func WithSkipPaths(paths ...string) Option {
	return func(o *Options) {
		o.skipPaths, o.skipPatterns = parsePaths(paths)
	}
}

// WithOptionalPaths 设置可选认证路径
// 有 token 时尝试验证并注入 claims；token 缺失或无效时均直接放行（claims 为 nil）
// 支持精确匹配和通配符模式（如 "/api/*"）
func WithOptionalPaths(paths ...string) Option {
	return func(o *Options) {
		o.optionalPaths, o.optionalPatterns = parsePaths(paths)
	}
}

// WithErrorHandler 设置未授权错误处理函数
func WithErrorHandler(fn func(httpx.Context)) Option {
	return func(o *Options) {
		o.errorHandler = fn
	}
}
