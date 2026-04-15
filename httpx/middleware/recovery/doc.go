/*
Package recovery 提供 HTTP 中间件，用于捕获 panic 并将其转换为错误。

# 概述

recovery 中间件可以捕获处理器中的 panic，防止程序崩溃，并将 panic 转换为标准错误返回。
支持自定义 panic 处理回调，用于日志记录或错误上报。

# 快速开始

基本用法：

	import (
		"github.com/f2xme/gox/httpx"
		"github.com/f2xme/gox/httpx/middleware/recovery"
	)

	func main() {
		// 创建 recovery 中间件
		r := recovery.New()

		// 应用到处理器
		handler := r(func(ctx httpx.Context) error {
			// 可能会 panic 的代码
			panic("something went wrong")
		})
	}

# 配置选项

## WithHandler - 自定义 panic 处理回调

设置自定义回调函数，在捕获 panic 时执行：

	import "log"

	r := recovery.New(
		recovery.WithHandler(func(ctx httpx.Context, err error) {
			// 记录 panic 日志
			log.Printf("panic recovered: %v", err)

			// 可以访问请求上下文
			log.Printf("request path: %s", ctx.Request().URL.Path)
		}),
	)

回调函数接收两个参数：
  - ctx: HTTP 上下文，可用于访问请求信息
  - err: 从 panic 恢复的错误

# 使用场景

## 1. 基础 panic 恢复

防止单个请求的 panic 导致整个服务崩溃：

	r := recovery.New()

	handler := r(func(ctx httpx.Context) error {
		// 即使这里 panic，服务也不会崩溃
		panic("unexpected error")
	})

## 2. 日志记录

记录 panic 信息用于调试和监控：

	r := recovery.New(
		recovery.WithHandler(func(ctx httpx.Context, err error) {
			logger.Error("panic recovered",
				"error", err,
				"path", ctx.Request().URL.Path,
				"method", ctx.Request().Method,
			)
		}),
	)

## 3. 错误上报

将 panic 上报到错误追踪系统（如 Sentry）：

	r := recovery.New(
		recovery.WithHandler(func(ctx httpx.Context, err error) {
			sentry.CaptureException(err)
		}),
	)

## 4. 自定义错误响应

在 panic 时返回自定义错误响应：

	r := recovery.New(
		recovery.WithHandler(func(ctx httpx.Context, err error) {
			// 记录日志
			log.Printf("panic: %v", err)

			// 可以在这里设置自定义响应头或状态码
			ctx.Response().Header().Set("X-Error-Type", "panic")
		}),
	)

# 错误转换

中间件会自动将 panic 转换为错误：

  - 如果 panic 的值是 error 类型，直接使用该错误
  - 如果 panic 的值不是 error 类型，使用 fmt.Errorf 包装

示例：

	// panic(errors.New("db error")) -> 返回 error: "db error"
	// panic("string error")          -> 返回 error: "string error"
	// panic(404)                     -> 返回 error: "404"

# 最佳实践

## 1. 始终使用 recovery 中间件

在生产环境中，应始终使用 recovery 中间件作为第一个中间件：

	app := httpx.New()
	app.Use(recovery.New())  // 第一个中间件
	app.Use(logger.New())
	app.Use(cors.New())

## 2. 记录 panic 日志

使用 WithHandler 记录 panic 信息，便于排查问题：

	r := recovery.New(
		recovery.WithHandler(func(ctx httpx.Context, err error) {
			log.Printf("[PANIC] %v\nPath: %s\nMethod: %s",
				err,
				ctx.Request().URL.Path,
				ctx.Request().Method,
			)
		}),
	)

## 3. 避免在回调中 panic

WithHandler 回调函数应该是安全的，避免在其中再次 panic：

	// 推荐：安全的回调
	recovery.WithHandler(func(ctx httpx.Context, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in recovery handler: %v", r)
			}
		}()
		// 可能出错的代码
	})

## 4. 结合错误处理中间件

recovery 中间件只负责捕获 panic，错误响应应由专门的错误处理中间件负责：

	app.Use(recovery.New())
	app.Use(errorHandler.New())  // 统一处理错误响应

# 性能考虑

  - recovery 中间件使用 defer/recover 机制，性能开销极小
  - 只有在发生 panic 时才会执行回调函数
  - 建议在生产环境中始终启用

# 线程安全

recovery 中间件是线程安全的，可以在多个 goroutine 中并发使用。
*/
package recovery
