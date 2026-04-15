/*
Package timeout 提供 HTTP 请求超时中间件。

# 概述

timeout 中间件取消超过指定时长的请求，防止长时间运行的请求消耗服务器资源。

# 快速开始

基本用法：

	import (
		"time"
		"github.com/f2xme/gox/httpx/middleware/timeout"
	)

	// 创建 5 秒超时的中间件
	app.Use(timeout.New(
		timeout.WithTimeout(5 * time.Second),
	))

自定义超时处理函数：

	app.Use(timeout.New(
		timeout.WithTimeout(5 * time.Second),
		timeout.WithHandler(func(ctx httpx.Context) {
			ctx.JSON(http.StatusGatewayTimeout, map[string]string{
				"error": "Request timeout",
			})
		}),
	))

# 配置选项

## WithTimeout

设置请求超时时长。默认值：30 秒。

	timeout.New(
		timeout.WithTimeout(10 * time.Second),
	)

## WithHandler

设置自定义超时处理函数。当请求超时时调用此函数。
如果未设置，返回默认的 503 Service Unavailable 响应。

	timeout.New(
		timeout.WithHandler(func(ctx httpx.Context) {
			ctx.JSON(503, map[string]any{
				"error": "timeout",
				"message": "请求处理超时",
			})
		}),
	)

# 行为说明

当请求超时时：
  - 如果设置了自定义处理函数，调用该函数
  - 否则返回默认的 503 Service Unavailable 响应
  - 请求的 context 会被取消，后续操作应检查 context.Err()

# 最佳实践

## 1. 根据业务场景设置合理的超时时长

	// API 请求：短超时
	timeout.New(timeout.WithTimeout(5 * time.Second))

	// 文件上传：长超时
	timeout.New(timeout.WithTimeout(60 * time.Second))

## 2. 在处理函数中检查 context 取消

	func handler(ctx httpx.Context) error {
		select {
		case <-ctx.Request().Context().Done():
			return ctx.Request().Context().Err()
		case result := <-doWork():
			return ctx.JSON(200, result)
		}
	}

## 3. 使用自定义处理函数提供友好的错误信息

	timeout.New(
		timeout.WithHandler(func(ctx httpx.Context) {
			ctx.JSON(503, map[string]any{
				"error": "timeout",
				"message": "服务器处理超时，请稍后重试",
				"retry_after": 60,
			})
		}),
	)
*/
package timeout
