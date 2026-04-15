/*
Package ratelimit 提供 HTTP 请求限流中间件。

# 概述

ratelimit 包实现了多种限流算法，用于保护 API 免受滥用并确保公平使用：
  - Token Bucket（令牌桶）：允许突发流量，同时维持平均速率
  - Leaky Bucket（漏桶）：以恒定速率平滑流量
  - Fixed Window（固定窗口）：在固定时间窗口内限制请求数
  - Sliding Window（滑动窗口）：提供更精确的限流

# 快速开始

基本用法：

	import (
		"github.com/f2xme/gox/httpx/adapter/gin"
		"github.com/f2xme/gox/httpx/middleware/ratelimit"
	)

	func main() {
		app := gin.New()

		// 使用默认配置（令牌桶，100 req/s）
		app.Use(ratelimit.New())

		app.GET("/api/data", handler)
		app.Run(":8080")
	}

# 配置选项

## 设置速率和突发量

	app.Use(ratelimit.New(
		ratelimit.WithRate(100),              // 每秒 100 个请求
		ratelimit.WithBurst(20),              // 允许突发 20 个请求
		ratelimit.WithStrategy(ratelimit.StrategyTokenBucket),
	))

## 选择限流策略

	// 令牌桶（默认）：适合允许短时突发的场景
	app.Use(ratelimit.New(
		ratelimit.WithStrategy(ratelimit.StrategyTokenBucket),
		ratelimit.WithRate(100),
		ratelimit.WithBurst(20),
	))

	// 漏桶：适合需要平滑流量的场景
	app.Use(ratelimit.New(
		ratelimit.WithStrategy(ratelimit.StrategyLeakyBucket),
		ratelimit.WithRate(100),
	))

	// 固定窗口：简单但可能有边界问题
	app.Use(ratelimit.New(
		ratelimit.WithStrategy(ratelimit.StrategyFixedWindow),
		ratelimit.WithRate(1000),
		ratelimit.WithWindow(time.Minute), // 每分钟 1000 个请求
	))

	// 滑动窗口：更精确但占用更多内存
	app.Use(ratelimit.New(
		ratelimit.WithStrategy(ratelimit.StrategySlidingWindow),
		ratelimit.WithRate(1000),
		ratelimit.WithWindow(time.Minute),
	))

## 自定义限流键

默认按客户端 IP 限流，可以自定义：

	// 按 IP 限流（默认）
	app.Use(ratelimit.New(
		ratelimit.WithKeyFunc(ratelimit.ByIP),
	))

	// 按请求路径限流
	app.Use(ratelimit.New(
		ratelimit.WithKeyFunc(ratelimit.ByPath),
	))

	// 按 IP + 路径组合限流
	app.Use(ratelimit.New(
		ratelimit.WithKeyFunc(ratelimit.ByIPAndPath),
	))

	// 按请求头限流（如 API Key）
	app.Use(ratelimit.New(
		ratelimit.WithKeyFunc(ratelimit.ByHeader("X-API-Key")),
	))

	// 自定义限流键
	app.Use(ratelimit.New(
		ratelimit.WithKeyFunc(func(ctx httpx.Context) string {
			// 从 JWT 中提取用户 ID
			userID := extractUserID(ctx)
			return "user:" + userID
		}),
	))

## 自定义限流响应

	app.Use(ratelimit.New(
		ratelimit.WithRate(10),
		ratelimit.WithHandler(func(ctx httpx.Context) {
			ctx.JSON(429, map[string]any{
				"error":   "rate_limit_exceeded",
				"message": "请求过于频繁，请稍后再试",
				"retry_after": 60,
			})
		}),
	))

# 限流策略对比

## Token Bucket（令牌桶）

特点：
  - 允许突发流量（最多 burst 个请求）
  - 以恒定速率补充令牌
  - 适合大多数场景

使用场景：
  - API 网关
  - 需要允许短时突发的服务

配置：

	ratelimit.New(
		ratelimit.WithStrategy(ratelimit.StrategyTokenBucket),
		ratelimit.WithRate(100),   // 每秒补充 100 个令牌
		ratelimit.WithBurst(20),   // 桶容量 20 个令牌
	)

## Leaky Bucket（漏桶）

特点：
  - 以恒定速率处理请求
  - 平滑突发流量
  - 不允许突发

使用场景：
  - 需要严格控制流量速率
  - 后端服务处理能力有限

配置：

	ratelimit.New(
		ratelimit.WithStrategy(ratelimit.StrategyLeakyBucket),
		ratelimit.WithRate(100), // 每秒处理 100 个请求
	)

## Fixed Window（固定窗口）

特点：
  - 在固定时间窗口内限制请求数
  - 实现简单，内存占用小
  - 窗口边界可能导致突发（边界问题）

使用场景：
  - 简单的限流需求
  - 对精度要求不高的场景

配置：

	ratelimit.New(
		ratelimit.WithStrategy(ratelimit.StrategyFixedWindow),
		ratelimit.WithRate(1000),
		ratelimit.WithWindow(time.Minute), // 每分钟 1000 个请求
	)

## Sliding Window（滑动窗口）

特点：
  - 更精确的限流
  - 避免固定窗口的边界问题
  - 内存占用较高（需要记录每个请求时间）

使用场景：
  - 需要精确限流
  - 对公平性要求高的场景

配置：

	ratelimit.New(
		ratelimit.WithStrategy(ratelimit.StrategySlidingWindow),
		ratelimit.WithRate(1000),
		ratelimit.WithWindow(time.Minute),
	)

# 使用场景

## 全局限流

保护整个应用：

	app := gin.New()
	app.Use(ratelimit.New(
		ratelimit.WithRate(1000),
		ratelimit.WithBurst(100),
	))

## 路由级限流

为特定路由设置不同的限流策略：

	app := gin.New()

	// 公开 API：严格限流
	publicAPI := app.Group("/api/public")
	publicAPI.Use(ratelimit.New(
		ratelimit.WithRate(10),
		ratelimit.WithBurst(5),
	))

	// 认证 API：宽松限流
	authAPI := app.Group("/api/auth")
	authAPI.Use(ratelimit.New(
		ratelimit.WithRate(100),
		ratelimit.WithBurst(20),
	))

## 按用户限流

为不同用户设置不同的限流策略：

	app.Use(ratelimit.New(
		ratelimit.WithKeyFunc(func(ctx httpx.Context) string {
			user := getCurrentUser(ctx)
			if user.IsPremium {
				return "premium:" + user.ID
			}
			return "free:" + user.ID
		}),
		ratelimit.WithRate(100),
	))

## 防止暴力破解

保护登录接口：

	app.POST("/login", ratelimit.New(
		ratelimit.WithRate(5),              // 每秒 5 次尝试
		ratelimit.WithWindow(time.Minute),  // 每分钟
		ratelimit.WithKeyFunc(ratelimit.ByIP),
	), loginHandler)

# 最佳实践

## 1. 选择合适的限流策略

  - 大多数场景：使用 Token Bucket（默认）
  - 需要平滑流量：使用 Leaky Bucket
  - 简单限流：使用 Fixed Window
  - 精确限流：使用 Sliding Window

## 2. 合理设置速率和突发量

	// 高流量 API
	ratelimit.New(
		ratelimit.WithRate(1000),
		ratelimit.WithBurst(200),
	)

	// 敏感操作（登录、支付）
	ratelimit.New(
		ratelimit.WithRate(5),
		ratelimit.WithBurst(2),
	)

## 3. 提供友好的错误信息

	ratelimit.New(
		ratelimit.WithHandler(func(ctx httpx.Context) {
			ctx.JSON(429, map[string]any{
				"error":       "rate_limit_exceeded",
				"message":     "请求过于频繁",
				"retry_after": 60, // 建议重试时间（秒）
			})
		}),
	)

## 4. 监控限流指标

记录被限流的请求，用于调整限流策略：

	ratelimit.New(
		ratelimit.WithHandler(func(ctx httpx.Context) {
			// 记录限流事件
			logRateLimitEvent(ctx.ClientIP(), ctx.Path())

			ctx.JSON(429, map[string]string{
				"error": "Too Many Requests",
			})
		}),
	)

## 5. 分层限流

结合全局限流和路由级限流：

	app := gin.New()

	// 全局限流：防止整体过载
	app.Use(ratelimit.New(
		ratelimit.WithRate(10000),
		ratelimit.WithBurst(1000),
	))

	// 路由级限流：保护特定端点
	app.POST("/api/expensive", ratelimit.New(
		ratelimit.WithRate(10),
		ratelimit.WithBurst(2),
	), expensiveHandler)

# 性能考虑

  - Token Bucket 和 Leaky Bucket：内存占用小，性能好
  - Fixed Window：最轻量，适合高并发场景
  - Sliding Window：内存占用较高，适合中等并发场景
  - 所有实现都是线程安全的，可以在多个 goroutine 中并发使用
  - 自动清理过期的限流状态，避免内存泄漏

# 线程安全

所有限流实现都是线程安全的，可以在多个 goroutine 中并发使用。
*/
package ratelimit
