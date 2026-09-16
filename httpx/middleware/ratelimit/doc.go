/*
Package ratelimit 提供 HTTP 请求限流中间件。

默认使用令牌桶策略，按客户端 IP 限流，每秒补充 100 个令牌，最多突发
100 个请求：

	app.Use(ratelimit.New())

可通过选项调整速率和限流键：

	app.Use(ratelimit.New(
		ratelimit.WithRate(10),
		ratelimit.WithBurst(5),
		ratelimit.WithKeyFunc(ratelimit.ByIPAndPath),
	))

包内支持令牌桶、漏桶、固定窗口和滑动窗口策略。固定窗口和滑动窗口的
Rate 表示每个 Window 允许的请求数：

	app.Use(ratelimit.New(
		ratelimit.WithStrategy(ratelimit.StrategySlidingWindow),
		ratelimit.WithRate(100),
		ratelimit.WithWindow(time.Minute),
	))

KeyFunc 可按 IP、路径、请求头或业务身份隔离额度。自定义函数应返回稳定、
非敏感的键：

	app.Use(ratelimit.New(
		ratelimit.WithKeyFunc(func(ctx httpx.Context) string {
			return "user:" + currentUserID(ctx)
		}),
	))

超限时默认返回 HTTP 429。WithHandler 可覆盖响应内容：

	app.Use(ratelimit.New(
		ratelimit.WithHandler(func(ctx httpx.Context) {
			ctx.JSON(429, map[string]any{
				"error": "rate_limit_exceeded",
			})
		}),
	))

所有策略支持并发调用。每个限流键的状态相互独立，长期未使用的状态会被
惰性清理。本包提供单进程限流；多实例共享额度需使用外部存储实现。
*/
package ratelimit
