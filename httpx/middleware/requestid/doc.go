/*
Package requestid 提供 HTTP 请求 ID 中间件，用于追踪和关联请求。

# 概述

requestid 中间件为每个 HTTP 请求生成或提取唯一标识符，并将其注入到响应头和上下文中。
这对于分布式追踪、日志关联和问题排查非常有用。

# 快速开始

基本用法：

	import (
		"github.com/f2xme/gox/httpx"
		"github.com/f2xme/gox/httpx/middleware/requestid"
	)

	func main() {
		app := httpx.New()

		// 使用默认配置
		app.Use(requestid.New())

		app.GET("/hello", func(ctx httpx.Context) error {
			// 获取请求 ID
			id := requestid.Get(ctx)
			return ctx.String(200, "Request ID: "+id)
		})

		app.Run(":8080")
	}

# 工作原理

中间件按以下顺序处理请求 ID：

1. 检查请求头中是否存在 X-Request-ID（可配置）
2. 如果存在，使用该值；否则生成新的 UUID
3. 将请求 ID 设置到响应头
4. 将请求 ID 存储到上下文的 "request_id" 键中
5. 如果配置了回调函数，调用该函数

# 配置选项

## WithHeaderKey - 自定义请求头名称

默认使用 X-Request-ID，可以自定义：

	app.Use(requestid.New(
		requestid.WithHeaderKey("X-Trace-ID"),
	))

## WithGenerator - 自定义 ID 生成器

默认生成 UUID 格式的 ID，可以使用自定义生成器：

	import "github.com/google/uuid"

	app.Use(requestid.New(
		requestid.WithGenerator(func() string {
			return uuid.New().String()
		}),
	))

使用递增 ID：

	var counter atomic.Int64

	app.Use(requestid.New(
		requestid.WithGenerator(func() string {
			return fmt.Sprintf("REQ-%d", counter.Add(1))
		}),
	))

## WithHandler - 请求 ID 回调

在生成或提取请求 ID 后执行自定义逻辑：

	app.Use(requestid.New(
		requestid.WithHandler(func(ctx httpx.Context, id string) {
			// 记录到日志
			logger.Info("request started", "request_id", id)

			// 或者设置到其他上下文
			ctx.Set("trace_id", id)
		}),
	))

# 获取请求 ID

## 使用 Get 函数

最简单的方式：

	func handler(ctx httpx.Context) error {
		id := requestid.Get(ctx)
		log.Printf("Processing request: %s", id)
		return ctx.JSON(200, map[string]string{"request_id": id})
	}

## 从上下文获取

请求 ID 也存储在上下文的 "request_id" 键中：

	func handler(ctx httpx.Context) error {
		id := ctx.Get("request_id").(string)
		return ctx.String(200, id)
	}

## 从响应头获取

客户端可以从响应头中读取请求 ID：

	curl -i http://localhost:8080/api
	# HTTP/1.1 200 OK
	# X-Request-ID: a1b2c3d4-e5f6-7890-abcd-ef1234567890

# 使用场景

## 1. 分布式追踪

将请求 ID 传递给下游服务：

	func callDownstream(ctx httpx.Context) error {
		id := requestid.Get(ctx)

		req, _ := http.NewRequest("GET", "http://downstream/api", nil)
		req.Header.Set("X-Request-ID", id)

		resp, err := http.DefaultClient.Do(req)
		// ...
	}

## 2. 日志关联

在日志中包含请求 ID，便于追踪整个请求链路：

	app.Use(requestid.New(
		requestid.WithHandler(func(ctx httpx.Context, id string) {
			// 将请求 ID 注入到结构化日志中
			logger := log.With("request_id", id)
			ctx.Set("logger", logger)
		}),
	))

	func handler(ctx httpx.Context) error {
		logger := ctx.Get("logger").(*slog.Logger)
		logger.Info("processing request")
		// ...
	}

## 3. 错误追踪

在错误响应中返回请求 ID，方便用户报告问题：

	func errorHandler(ctx httpx.Context, err error) error {
		id := requestid.Get(ctx)
		return ctx.JSON(500, map[string]string{
			"error":      "Internal Server Error",
			"request_id": id,
			"message":    "Please contact support with this request ID",
		})
	}

## 4. 性能监控

使用请求 ID 关联性能指标：

	app.Use(requestid.New(
		requestid.WithHandler(func(ctx httpx.Context, id string) {
			start := time.Now()
			ctx.Set("start_time", start)

			// 在响应后记录耗时
			ctx.OnAfterResponse(func() {
				duration := time.Since(start)
				metrics.RecordLatency(id, duration)
			})
		}),
	))

# 与其他中间件集成

## 与 logger 中间件配合

	import (
		"github.com/f2xme/gox/httpx/middleware/logger"
		"github.com/f2xme/gox/httpx/middleware/requestid"
	)

	app.Use(requestid.New())
	app.Use(logger.New(
		logger.WithFormatter(func(ctx httpx.Context, info logger.Info) string {
			id := requestid.Get(ctx)
			return fmt.Sprintf("[%s] %s %s %d %s",
				id, info.Method, info.Path, info.Status, info.Latency)
		}),
	))

## 与 recovery 中间件配合

	import (
		"github.com/f2xme/gox/httpx/middleware/recovery"
		"github.com/f2xme/gox/httpx/middleware/requestid"
	)

	app.Use(requestid.New())
	app.Use(recovery.New(
		recovery.WithHandler(func(ctx httpx.Context, err any) {
			id := requestid.Get(ctx)
			log.Printf("[%s] panic recovered: %v", id, err)
		}),
	))

# 最佳实践

## 1. 优先使用客户端提供的请求 ID

如果客户端已经提供了请求 ID（如从网关传递），中间件会自动使用它：

	curl -H "X-Request-ID: client-provided-id" http://localhost:8080/api

## 2. 在中间件链的最前面使用

确保请求 ID 在所有其他中间件之前生成：

	app.Use(requestid.New())        // 第一个
	app.Use(logger.New())           // 可以使用请求 ID
	app.Use(recovery.New())         // 可以使用请求 ID

## 3. 使用结构化日志

将请求 ID 作为结构化字段，而不是拼接到消息中：

	// 推荐
	logger.Info("request processed", "request_id", id, "user_id", userID)

	// 不推荐
	logger.Info(fmt.Sprintf("request %s processed for user %d", id, userID))

## 4. 传递给下游服务

在调用其他服务时，始终传递请求 ID：

	func callAPI(ctx httpx.Context, url string) error {
		id := requestid.Get(ctx)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("X-Request-ID", id)
		// ...
	}

## 5. 在错误响应中包含请求 ID

帮助用户和支持团队快速定位问题：

	{
		"error": "Resource not found",
		"request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"timestamp": "2026-04-15T10:30:00Z"
	}

# 性能考虑

- 默认生成器使用 crypto/rand，性能足够好（约 1-2 微秒/次）
- 如果需要更高性能，可以使用自定义生成器（如 UUID v4 或递增 ID）
- 中间件本身开销极小（< 1 微秒），可以放心使用

# 线程安全

所有配置选项和中间件本身都是线程安全的，可以在多个 goroutine 中并发使用。
*/
package requestid
