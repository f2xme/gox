/*
Package tracing 提供分布式追踪中间件，用于跟踪 HTTP 请求在服务间的流转。

# 概述

tracing 中间件集成分布式追踪系统（如 OpenTelemetry、Jaeger、Zipkin），
自动从请求头提取追踪上下文、创建 span、注入追踪上下文供下游服务使用。

# 快速开始

基本用法：

	import (
		"github.com/f2xme/gox/httpx"
		"github.com/f2xme/gox/httpx/middleware/tracing"
	)

	func main() {
		app := httpx.New()

		// 使用自定义 tracer
		app.Use(tracing.New(
			tracing.WithTracer(myTracer),
		))

		app.GET("/users/:id", handleGetUser)
		app.Run(":8080")
	}

# 配置选项

## WithTracer - 设置追踪器

必须提供 tracer 实现，否则中间件为空操作：

	app.Use(tracing.New(
		tracing.WithTracer(myTracer),
	))

## WithOperationName - 自定义操作名

默认操作名为 "METHOD PATH"（如 "GET /api/users"），可自定义：

	app.Use(tracing.New(
		tracing.WithTracer(myTracer),
		tracing.WithOperationName(func(ctx httpx.Context) string {
			return fmt.Sprintf("%s %s", ctx.Method(), ctx.Path())
		}),
	))

## WithHandler - 自定义处理器

在请求完成后调用，用于添加自定义标签或日志：

	app.Use(tracing.New(
		tracing.WithTracer(myTracer),
		tracing.WithHandler(func(ctx httpx.Context, span tracing.Span) {
			// 添加自定义标签
			span.SetTag("user.id", ctx.Get("user_id"))
			span.SetTag("tenant.id", ctx.Get("tenant_id"))
		}),
	))

# 核心接口

## Tracer - 追踪器接口

实现此接口以集成不同的追踪系统：

	type Tracer interface {
		// Extract 从 HTTP 请求头提取追踪上下文
		Extract(ctx httpx.Context) context.Context

		// StartSpan 创建新的 span
		StartSpan(ctx context.Context, operationName string) Span

		// Inject 将追踪上下文注入到 context 中
		Inject(ctx context.Context, span Span) context.Context
	}

## Span - 追踪单元接口

表示追踪中的一个操作单元：

	type Span interface {
		// SetTag 设置标签
		SetTag(key string, value any)

		// SetBaggageItem 设置传播到子 span 的行李项
		SetBaggageItem(key, value string)

		// Finish 完成 span
		Finish()

		// Context 返回 span 的 context 用于传播
		Context() context.Context
	}

# 自动设置的标签

中间件自动为每个请求设置以下标签：

  - http.method: 请求方法（GET、POST 等）
  - http.url: 请求路径
  - http.client_ip: 客户端 IP
  - http.duration_ms: 请求耗时（毫秒）
  - error: 是否发生错误（true/false）
  - error.message: 错误信息（如果有）

# 获取当前 Span

使用 GetSpan 从上下文中获取当前 span：

	func handleGetUser(ctx httpx.Context) error {
		span := tracing.GetSpan(ctx)
		if span != nil {
			span.SetTag("user.id", ctx.Param("id"))
		}

		// 处理请求
		return ctx.JSON(200, user)
	}

# 实现示例

## OpenTelemetry 集成

	import (
		"go.opentelemetry.io/otel"
		"go.opentelemetry.io/otel/trace"
	)

	type otelTracer struct {
		tracer trace.Tracer
	}

	func (t *otelTracer) Extract(ctx httpx.Context) context.Context {
		// 从请求头提取追踪上下文
		return otel.GetTextMapPropagator().Extract(
			ctx.Request().Context(),
			propagation.HeaderCarrier(ctx.Request().Header),
		)
	}

	func (t *otelTracer) StartSpan(ctx context.Context, operationName string) tracing.Span {
		_, span := t.tracer.Start(ctx, operationName)
		return &otelSpan{span: span}
	}

	func (t *otelTracer) Inject(ctx context.Context, span tracing.Span) context.Context {
		return trace.ContextWithSpan(ctx, span.(*otelSpan).span)
	}

# 最佳实践

## 1. 始终提供 Tracer

如果不提供 tracer，中间件将不执行任何操作：

	// 推荐：提供 tracer
	app.Use(tracing.New(tracing.WithTracer(myTracer)))

	// 不推荐：未提供 tracer（中间件无效）
	app.Use(tracing.New())

## 2. 在业务逻辑中添加自定义标签

使用 GetSpan 获取当前 span 并添加业务相关标签：

	span := tracing.GetSpan(ctx)
	if span != nil {
		span.SetTag("order.id", orderID)
		span.SetTag("payment.method", "credit_card")
	}

## 3. 使用 WithHandler 统一添加标签

对于需要在所有请求中添加的标签，使用 WithHandler：

	app.Use(tracing.New(
		tracing.WithTracer(myTracer),
		tracing.WithHandler(func(ctx httpx.Context, span tracing.Span) {
			// 统一添加租户信息
			if tenantID, ok := ctx.Get("tenant_id"); ok {
				span.SetTag("tenant.id", tenantID)
			}
		}),
	))

## 4. 合理命名操作

操作名应清晰描述请求的业务含义：

	tracing.WithOperationName(func(ctx httpx.Context) string {
		// 包含路由参数的操作名
		return fmt.Sprintf("%s %s", ctx.Method(), ctx.Route())
	})

# 性能考虑

  - 追踪会增加少量延迟（通常 < 1ms）
  - 采样率应根据流量调整（高流量场景建议 1%-10%）
  - 避免在 span 中存储大量数据
  - 使用异步上报减少对请求的影响

# 线程安全

所有 Tracer 和 Span 实现都应是线程安全的，可在多个 goroutine 中并发使用。
*/
package tracing
