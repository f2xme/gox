/*
Package cors 提供 CORS（跨域资源共享）中间件。

# 概述

cors 包实现了 W3C CORS 规范，允许你控制哪些来源可以访问你的 API。
它处理预检请求（OPTIONS）和实际请求的 CORS 头设置。

# 快速开始

基本用法：

	import (
		"github.com/f2xme/gox/httpx/middleware/cors"
		"github.com/f2xme/gox/httpx/adapter/gin"
	)

	func main() {
		engine := gin.New()

		// 使用默认配置（允许所有来源）
		engine.Use(cors.New())

		// 自定义配置
		engine.Use(cors.New(
			cors.WithOrigins([]string{"https://example.com", "https://app.example.com"}),
			cors.WithMethods([]string{"GET", "POST", "PUT", "DELETE"}),
			cors.WithHeaders([]string{"Content-Type", "Authorization"}),
			cors.WithCredentials(true),
			cors.WithMaxAge(3600),
		))

		engine.Run(":8080")
	}

# 配置选项

## WithOrigins

设置允许的来源列表：

	cors.New(
		cors.WithOrigins([]string{"https://example.com"}),
	)

使用 "*" 允许所有来源（默认）：

	cors.New(
		cors.WithOrigins([]string{"*"}),
	)

## WithMethods

设置允许的 HTTP 方法：

	cors.New(
		cors.WithMethods([]string{"GET", "POST", "PUT", "DELETE"}),
	)

默认值：GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS

## WithHeaders

设置允许的请求头：

	cors.New(
		cors.WithHeaders([]string{"Content-Type", "Authorization", "X-Request-ID"}),
	)

默认值：Origin, Content-Length, Content-Type, Authorization

## WithExposeHeaders

设置暴露给浏览器的响应头：

	cors.New(
		cors.WithExposeHeaders([]string{"X-Total-Count", "X-Page-Number"}),
	)

默认不暴露任何自定义响应头。

## WithCredentials

启用凭证支持（cookies、HTTP 认证）：

	cors.New(
		cors.WithCredentials(true),
	)

注意：启用凭证时，不能使用 "*" 作为来源。

## WithMaxAge

设置预检请求的缓存时间（秒）：

	cors.New(
		cors.WithMaxAge(3600), // 缓存 1 小时
	)

默认不设置缓存。

# 工作原理

## 简单请求

对于简单请求（GET、POST 等），中间件会：
1. 检查 Origin 头是否在允许列表中
2. 设置 Access-Control-Allow-Origin 响应头
3. 如果启用了凭证，设置 Access-Control-Allow-Credentials
4. 如果配置了暴露头，设置 Access-Control-Expose-Headers

## 预检请求

对于预检请求（OPTIONS），中间件会：
1. 执行简单请求的所有检查
2. 设置 Access-Control-Allow-Methods
3. 设置 Access-Control-Allow-Headers
4. 如果配置了 MaxAge，设置 Access-Control-Max-Age
5. 返回 204 No Content

# 使用场景

## 场景 1：开发环境（允许所有来源）

	engine.Use(cors.New()) // 默认允许所有来源

## 场景 2：生产环境（限制特定来源）

	engine.Use(cors.New(
		cors.WithOrigins([]string{
			"https://example.com",
			"https://app.example.com",
		}),
		cors.WithCredentials(true),
	))

## 场景 3：公共 API（允许所有来源，但限制方法）

	engine.Use(cors.New(
		cors.WithOrigins([]string{"*"}),
		cors.WithMethods([]string{"GET", "POST"}),
	))

## 场景 4：需要自定义头的 API

	engine.Use(cors.New(
		cors.WithOrigins([]string{"https://example.com"}),
		cors.WithHeaders([]string{
			"Content-Type",
			"Authorization",
			"X-API-Key",
			"X-Request-ID",
		}),
		cors.WithExposeHeaders([]string{
			"X-Total-Count",
			"X-Page-Number",
		}),
	))

# 最佳实践

## 1. 生产环境不要使用 "*"

	// 不推荐
	cors.New(cors.WithOrigins([]string{"*"}))

	// 推荐
	cors.New(cors.WithOrigins([]string{
		"https://example.com",
		"https://app.example.com",
	}))

## 2. 启用凭证时必须指定具体来源

	// 错误：启用凭证时不能使用 "*"
	cors.New(
		cors.WithOrigins([]string{"*"}),
		cors.WithCredentials(true), // 浏览器会拒绝
	)

	// 正确
	cors.New(
		cors.WithOrigins([]string{"https://example.com"}),
		cors.WithCredentials(true),
	)

## 3. 合理设置 MaxAge 减少预检请求

	cors.New(
		cors.WithMaxAge(3600), // 缓存 1 小时
	)

## 4. 只暴露必要的响应头

	// 不推荐：暴露所有头
	cors.New(cors.WithExposeHeaders([]string{"*"}))

	// 推荐：只暴露需要的头
	cors.New(cors.WithExposeHeaders([]string{"X-Total-Count"}))

## 5. 根据环境使用不同配置

	var corsMiddleware httpx.Middleware
	if os.Getenv("ENV") == "production" {
		corsMiddleware = cors.New(
			cors.WithOrigins([]string{"https://example.com"}),
			cors.WithCredentials(true),
		)
	} else {
		corsMiddleware = cors.New() // 开发环境允许所有来源
	}
	engine.Use(corsMiddleware)

# 安全考虑

1. 生产环境必须限制来源列表
2. 启用凭证时要特别小心，确保来源可信
3. 不要暴露敏感的响应头
4. 定期审查允许的来源列表

# 性能考虑

1. 使用 WithMaxAge 缓存预检请求结果
2. 中间件会在来源不匹配时快速返回，不影响性能
3. 字符串比较是 O(n)，但来源列表通常很小

# 线程安全

中间件是线程安全的，可以在多个 goroutine 中并发使用。
*/
package cors
