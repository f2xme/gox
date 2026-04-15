package tracing

import (
	"context"

	"github.com/f2xme/gox/httpx"
)

// Options 定义追踪中间件配置
type Options struct {
	Tracer        Tracer
	OperationName func(ctx httpx.Context) string
	Handler       func(ctx httpx.Context, span Span)
}

// Option 配置追踪中间件
type Option func(*Options)

// defaultOptions 返回默认配置
func defaultOptions() *Options {
	return &Options{
		Tracer: nil,
		OperationName: func(ctx httpx.Context) string {
			return ctx.Method() + " " + ctx.Path()
		},
		Handler: nil,
	}
}

// WithTracer 设置追踪器实现
// 如果未设置，中间件将不执行任何操作
func WithTracer(t Tracer) Option {
	return func(c *Options) {
		c.Tracer = t
	}
}

// WithOperationName 设置自定义操作名生成函数
// 默认为 "METHOD PATH"（如 "GET /api/users"）
func WithOperationName(fn func(ctx httpx.Context) string) Option {
	return func(c *Options) {
		c.OperationName = fn
	}
}

// WithHandler 设置请求完成后调用的自定义处理函数
// 用于添加自定义标签或日志记录
func WithHandler(fn func(ctx httpx.Context, span Span)) Option {
	return func(c *Options) {
		c.Handler = fn
	}
}

// noopTracer 是用于测试或禁用追踪时的空操作追踪器
type noopTracer struct{}

func (noopTracer) Extract(ctx httpx.Context) context.Context {
	return ctx.Request().Context()
}

func (noopTracer) StartSpan(ctx context.Context, operationName string) Span {
	return &noopSpan{}
}

func (noopTracer) Inject(ctx context.Context, span Span) context.Context {
	return ctx
}

// noopSpan 是空操作 span 实现
type noopSpan struct{}

func (noopSpan) SetTag(key string, value any)       {}
func (noopSpan) SetBaggageItem(key, value string)   {}
func (noopSpan) Finish()                            {}
func (noopSpan) Context() context.Context           { return context.Background() }
