/*
Package validator 提供 HTTP 请求验证中间件。

# 概述

validator 中间件用于在请求到达业务逻辑之前进行预验证，支持多种验证规则：
  - 请求体大小限制
  - Content-Type 白名单
  - 必需请求头检查
  - 自定义验证逻辑

通过提前拦截不合法的请求，可以减轻后端负担并提供统一的错误响应。

# 快速开始

基本用法：

	import (
		"github.com/f2xme/gox/httpx"
		"github.com/f2xme/gox/httpx/middleware/validator"
	)

	func main() {
		app := httpx.New()

		// 限制请求体大小为 10MB
		app.Use(validator.New(
			validator.WithMaxBodySize(10 * 1024 * 1024),
		))

		app.Start(":8080")
	}

# 配置选项

## WithMaxBodySize - 请求体大小限制

限制请求体的最大字节数，防止过大的请求消耗服务器资源：

	app.Use(validator.New(
		validator.WithMaxBodySize(5 * 1024 * 1024), // 5MB
	))

  - 基于 Content-Length 请求头进行检查
  - 超过限制返回 413 Payload Too Large
  - 设置为 0 表示不限制（默认）

## WithAllowedContentTypes - Content-Type 白名单

仅允许特定的 Content-Type，拒绝其他类型的请求：

	app.Use(validator.New(
		validator.WithAllowedContentTypes(
			"application/json",
			"application/xml",
		),
	))

  - 缺少 Content-Type 请求头返回 415 Unsupported Media Type
  - 不在白名单内返回 415 Unsupported Media Type
  - 适用于严格的 API 接口

## WithRequiredHeaders - 必需请求头

要求请求必须包含指定的请求头：

	app.Use(validator.New(
		validator.WithRequiredHeaders(
			"X-API-Key",
			"X-Request-ID",
		),
	))

  - 缺少任何一个必需请求头返回 400 Bad Request
  - 常用于 API 密钥、追踪 ID 等场景

## WithCustomValidator - 自定义验证逻辑

添加自定义验证函数，实现业务特定的验证规则：

	app.Use(validator.New(
		validator.WithCustomValidator(func(ctx httpx.Context) error {
			// 验证 API 密钥格式
			apiKey := ctx.Header("X-API-Key")
			if len(apiKey) != 32 {
				return fmt.Errorf("invalid API key format")
			}
			return nil
		}),
		validator.WithCustomValidator(func(ctx httpx.Context) error {
			// 验证请求来源
			origin := ctx.Header("Origin")
			if !isAllowedOrigin(origin) {
				return fmt.Errorf("origin not allowed")
			}
			return nil
		}),
	))

  - 可以添加多个自定义验证器，按顺序执行
  - 返回 error 时请求被拒绝，返回 400 Bad Request
  - 错误信息会作为响应消息返回

## WithErrorHandler - 自定义错误处理

自定义验证失败时的错误响应格式：

	app.Use(validator.New(
		validator.WithMaxBodySize(1024 * 1024),
		validator.WithErrorHandler(func(ctx httpx.Context, code int, message string) {
			ctx.JSON(code, map[string]any{
				"success": false,
				"error": map[string]any{
					"code":    code,
					"message": message,
				},
			})
		}),
	))

  - 默认使用 JSON 格式返回错误
  - 可以自定义为 XML、纯文本等格式
  - 适配不同的 API 规范

# 组合使用

多个验证规则可以组合使用：

	app.Use(validator.New(
		validator.WithMaxBodySize(10 * 1024 * 1024),
		validator.WithAllowedContentTypes("application/json"),
		validator.WithRequiredHeaders("X-API-Key", "X-Request-ID"),
		validator.WithCustomValidator(validateAPIKey),
	))

验证顺序：
  1. 请求体大小检查
  2. Content-Type 检查
  3. 必需请求头检查
  4. 自定义验证器（按添加顺序）

任何一步失败都会立即返回错误，不再执行后续验证。

# 使用场景

## API 网关

在网关层统一验证所有请求：

	gateway := httpx.New()
	gateway.Use(validator.New(
		validator.WithMaxBodySize(5 * 1024 * 1024),
		validator.WithRequiredHeaders("X-API-Key"),
		validator.WithCustomValidator(validateAPIKey),
	))

## 文件上传接口

限制上传文件的大小和类型：

	app.POST("/upload", func(ctx httpx.Context) error {
		// 处理文件上传
		return ctx.String(200, "OK")
	}, validator.New(
		validator.WithMaxBodySize(50 * 1024 * 1024), // 50MB
		validator.WithAllowedContentTypes("multipart/form-data"),
	))

## Webhook 接收

验证 webhook 请求的签名和格式：

	app.POST("/webhook", func(ctx httpx.Context) error {
		// 处理 webhook
		return ctx.String(200, "OK")
	}, validator.New(
		validator.WithRequiredHeaders("X-Webhook-Signature"),
		validator.WithCustomValidator(verifyWebhookSignature),
	))

# 错误响应

默认错误响应格式（JSON）：

	{
		"code": 413,
		"message": "Request body too large"
	}

常见错误码：
  - 400 Bad Request: 缺少必需请求头、自定义验证失败
  - 413 Payload Too Large: 请求体超过大小限制
  - 415 Unsupported Media Type: Content-Type 不被允许

# 最佳实践

## 1. 在路由组级别应用

对一组路由应用相同的验证规则：

	api := app.Group("/api", validator.New(
		validator.WithRequiredHeaders("X-API-Key"),
	))

	api.GET("/users", getUsers)
	api.POST("/users", createUser)

## 2. 合理设置请求体大小限制

根据接口类型设置不同的限制：

	// JSON API: 1MB
	app.Use(validator.New(validator.WithMaxBodySize(1024 * 1024)))

	// 文件上传: 50MB
	app.POST("/upload", uploadHandler, validator.New(
		validator.WithMaxBodySize(50 * 1024 * 1024),
	))

## 3. 自定义验证器保持简单

自定义验证器应该只做验证，不要包含复杂的业务逻辑：

	// 好的做法：简单验证
	validator.WithCustomValidator(func(ctx httpx.Context) error {
		if ctx.Header("X-API-Key") == "" {
			return fmt.Errorf("API key is required")
		}
		return nil
	})

	// 不好的做法：包含数据库查询等复杂逻辑
	validator.WithCustomValidator(func(ctx httpx.Context) error {
		user, err := db.FindUserByAPIKey(ctx.Header("X-API-Key"))
		// ...
	})

## 4. 提供清晰的错误信息

错误信息应该明确指出问题所在：

	validator.WithCustomValidator(func(ctx httpx.Context) error {
		apiKey := ctx.Header("X-API-Key")
		if len(apiKey) != 32 {
			return fmt.Errorf("X-API-Key must be exactly 32 characters")
		}
		return nil
	})

# 性能考虑

  - 验证在请求处理的最早阶段进行，可以快速拒绝不合法请求
  - Content-Length 检查不需要读取请求体，开销极小
  - 自定义验证器应避免阻塞操作（如数据库查询）
  - 验证失败后立即返回，不会执行后续中间件和处理器

# 线程安全

validator 中间件是无状态的，可以安全地在多个 goroutine 中并发使用。
*/
package validator
