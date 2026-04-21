package validator

import (
	"github.com/f2xme/gox/httpx"
)

// Options 定义验证中间件配置
type Options struct {
	MaxBodySize      int64
	AllowedTypes     map[string]bool
	RequiredHeaders  []string
	CustomValidators []func(ctx httpx.Context) error
	ErrorHandler     func(ctx httpx.Context, code int, message string)
}

// Option 配置验证中间件
type Option func(*Options)

func defaultOptions() *Options {
	return &Options{
		MaxBodySize:  0, // 0 表示无限制
		ErrorHandler: defaultErrorHandler,
	}
}

func defaultErrorHandler(ctx httpx.Context, code int, message string) {
	ctx.JSON(code, httpx.NewFailResponse(message))
}

// WithMaxBodySize 设置允许的最大请求体大小（字节）
// 如果 Content-Length 头超过此限制，请求将被拒绝并返回 413
// 值为 0 表示无限制
func WithMaxBodySize(size int64) Option {
	return func(c *Options) {
		c.MaxBodySize = size
	}
}

// WithAllowedContentTypes 限制允许的 Content-Type 值
// 如果请求的 Content-Type 不匹配任何允许的类型，将被拒绝并返回 415
func WithAllowedContentTypes(types ...string) Option {
	return func(c *Options) {
		c.AllowedTypes = make(map[string]bool, len(types))
		for _, t := range types {
			c.AllowedTypes[t] = true
		}
	}
}

// WithRequiredHeaders 指定请求中必须存在的请求头
// 如果缺少任何必需的请求头，请求将被拒绝并返回 400
func WithRequiredHeaders(headers ...string) Option {
	return func(c *Options) {
		c.RequiredHeaders = headers
	}
}

// WithCustomValidator 添加自定义验证函数
// 如果验证失败，函数应返回错误
// 可以添加多个自定义验证器，它们将按顺序执行
func WithCustomValidator(fn func(ctx httpx.Context) error) Option {
	return func(c *Options) {
		c.CustomValidators = append(c.CustomValidators, fn)
	}
}

// WithErrorHandler 设置验证失败时的自定义错误处理函数
func WithErrorHandler(fn func(ctx httpx.Context, code int, message string)) Option {
	return func(c *Options) {
		c.ErrorHandler = fn
	}
}
