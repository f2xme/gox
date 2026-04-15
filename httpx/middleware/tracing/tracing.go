package tracing

import (
	"context"
	"time"

	"github.com/f2xme/gox/httpx"
)

// New 创建追踪中间件
// 如果未提供追踪器，中间件将不执行任何操作
func New(opts ...Option) httpx.Middleware {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}

	return func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			if cfg.Tracer == nil {
				return next(ctx)
			}

			// Extract trace context from headers
			traceCtx := cfg.Tracer.Extract(ctx)

			// Start a new span
			span := cfg.Tracer.StartSpan(traceCtx, cfg.OperationName(ctx))
			defer span.Finish()

			// Set span tags
			span.SetTag("http.method", ctx.Method())
			span.SetTag("http.url", ctx.Path())
			span.SetTag("http.client_ip", ctx.ClientIP())

			// Inject trace context into request context
			reqCtx := cfg.Tracer.Inject(ctx.Request().Context(), span)
			req := ctx.Request()
			*req = *req.WithContext(reqCtx)

			// Store span in context for downstream use
			ctx.Set("tracing.span", span)

			// Execute handler
			start := time.Now()
			err := next(ctx)
			duration := time.Since(start)

			// Record response information
			span.SetTag("http.duration_ms", duration.Milliseconds())

			if err != nil {
				span.SetTag("error", true)
				span.SetTag("error.message", err.Error())
			}

			// Call custom handler if provided
			if cfg.Handler != nil {
				cfg.Handler(ctx, span)
			}

			return err
		}
	}
}

// Tracer 定义分布式追踪接口
// 实现可以集成 OpenTelemetry、Jaeger、Zipkin 等
type Tracer interface {
	// Extract 从 HTTP 头中提取追踪上下文
	Extract(ctx httpx.Context) context.Context

	// StartSpan 使用给定的上下文和操作名启动新的 span
	StartSpan(ctx context.Context, operationName string) Span

	// Inject 将追踪上下文注入到给定的上下文中
	Inject(ctx context.Context, span Span) context.Context
}

// Span 表示追踪中的单个 span
type Span interface {
	// SetTag 在 span 上设置标签
	SetTag(key string, value any)

	// SetBaggageItem 设置传播到子 span 的行李项
	SetBaggageItem(key, value string)

	// Finish 完成 span
	Finish()

	// Context 返回 span 的上下文用于传播
	Context() context.Context
}

// GetSpan 从上下文中获取当前 span
// 如果不存在 span 则返回 nil
func GetSpan(ctx httpx.Context) Span {
	if span, ok := ctx.Get("tracing.span"); ok {
		return span.(Span)
	}
	return nil
}
