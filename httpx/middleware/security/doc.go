/*
Package security 提供 HTTP 安全中间件，防护常见的 Web 攻击。

# 概述

security 中间件为 HTTP 应用提供多层安全防护，包括安全响应头、主机验证、
XSS 防护、SQL 注入检测和 CSRF 保护。通过简单的配置即可为应用添加企业级安全防护。

# 快速开始

基本用法（使用默认安全配置）：

	import (
		"github.com/f2xme/gox/httpx"
		"github.com/f2xme/gox/httpx/middleware/security"
	)

	func main() {
		app := httpx.New()

		// 应用默认安全中间件
		app.Use(security.New())

		app.GET("/", func(ctx httpx.Context) error {
			return ctx.String(200, "Hello, World!")
		})

		app.Run(":8080")
	}

# 配置选项

## 自定义安全响应头

覆盖默认的安全响应头：

	app.Use(security.New(
		security.WithSecurityHeaders(map[string]string{
			"X-Frame-Options":           "SAMEORIGIN",
			"X-Content-Type-Options":    "nosniff",
			"X-XSS-Protection":          "1; mode=block",
			"Strict-Transport-Security": "max-age=63072000; includeSubDomains; preload",
			"Content-Security-Policy":   "default-src 'self'; script-src 'self' 'unsafe-inline'",
		}),
	))

## 主机白名单验证

限制允许访问的主机名：

	app.Use(security.New(
		security.WithAllowedHosts("example.com", "api.example.com", "www.example.com"),
	))

不在白名单中的主机请求将被拒绝（返回 400 Bad Request）。

## XSS 防护

启用 XSS 攻击模式检测：

	app.Use(security.New(
		security.WithXSSProtection(true),
	))

中间件会检查请求参数中是否包含常见的 XSS 攻击载荷（如 <script>、javascript: 等）。

## SQL 注入防护

启用 SQL 注入模式检测：

	app.Use(security.New(
		security.WithSQLInjectionProtection(true),
	))

中间件会检查请求参数中是否包含常见的 SQL 注入模式（如 UNION SELECT、DROP TABLE 等）。

## CSRF 保护

启用跨站请求伪造保护：

	app.Use(security.New(
		security.WithCSRFProtection(security.CSRFConfig{
			TokenLength:    32,                    // CSRF token 长度
			TokenLookup:    "header:X-CSRF-Token", // token 查找位置
			CookieName:     "_csrf",               // cookie 名称
			CookiePath:     "/",                   // cookie 路径
			CookieMaxAge:   86400,                 // cookie 有效期（秒）
			CookieSecure:   true,                  // 仅 HTTPS
			CookieSameSite: "Strict",              // SameSite 策略
		}),
	))

CSRF 保护工作原理：
  - GET/HEAD/OPTIONS 请求：自动生成 CSRF token 并设置到 cookie
  - POST/PUT/DELETE 等修改请求：验证请求中的 token 是否与 cookie 中的匹配

前端需要从 cookie 中读取 token 并在请求头中发送：

	// JavaScript 示例
	fetch('/api/data', {
		method: 'POST',
		headers: {
			'X-CSRF-Token': getCookie('_csrf'),
			'Content-Type': 'application/json',
		},
		body: JSON.stringify(data),
	})

## 自定义错误处理

自定义安全检查失败时的响应：

	app.Use(security.New(
		security.WithErrorHandler(func(ctx httpx.Context, code int, message string) {
			ctx.JSON(code, map[string]any{
				"error":   message,
				"code":    code,
				"blocked": true,
			})
		}),
	))

## 自定义 CSP 策略

单独设置 Content-Security-Policy 头：

	app.Use(security.New(
		security.WithContentSecurityPolicy(
			"default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' https://cdn.example.com; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self' https://api.example.com",
		),
	))

# 完整示例

组合多个安全选项：

	app.Use(security.New(
		// 自定义安全响应头
		security.WithSecurityHeaders(map[string]string{
			"X-Frame-Options":           "DENY",
			"X-Content-Type-Options":    "nosniff",
			"X-XSS-Protection":          "1; mode=block",
			"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		}),

		// 主机白名单
		security.WithAllowedHosts("example.com", "api.example.com"),

		// 启用 XSS 和 SQL 注入防护
		security.WithXSSProtection(true),
		security.WithSQLInjectionProtection(true),

		// 启用 CSRF 保护
		security.WithCSRFProtection(security.CSRFConfig{
			TokenLength:    32,
			CookieName:     "_csrf",
			CookieSecure:   true,
			CookieSameSite: "Strict",
		}),

		// 自定义 CSP
		security.WithContentSecurityPolicy("default-src 'self'; script-src 'self' 'unsafe-inline'"),

		// 自定义错误处理
		security.WithErrorHandler(func(ctx httpx.Context, code int, message string) {
			ctx.JSON(code, map[string]string{"error": message})
		}),
	))

# 默认配置

如果不提供任何选项，中间件将使用以下默认配置：

安全响应头：
  - X-Content-Type-Options: nosniff
  - X-Frame-Options: DENY
  - X-XSS-Protection: 1; mode=block
  - Strict-Transport-Security: max-age=31536000; includeSubDomains
  - Content-Security-Policy: default-src 'self'

其他默认值：
  - XSS 防护：禁用
  - SQL 注入防护：禁用
  - CSRF 保护：禁用
  - 主机白名单：禁用（允许所有主机）

# 最佳实践

## 1. 生产环境必须启用 HTTPS

安全响应头（如 HSTS、Secure Cookie）只有在 HTTPS 下才能发挥作用：

	app.Use(security.New(
		security.WithCSRFProtection(security.CSRFConfig{
			CookieSecure: true, // 生产环境必须为 true
		}),
	))

## 2. 根据应用类型选择防护策略

API 服务：

	app.Use(security.New(
		security.WithSecurityHeaders(map[string]string{
			"X-Content-Type-Options": "nosniff",
		}),
		security.WithSQLInjectionProtection(true),
	))

Web 应用：

	app.Use(security.New(
		security.WithXSSProtection(true),
		security.WithSQLInjectionProtection(true),
		security.WithCSRFProtection(security.CSRFConfig{
			CookieSecure:   true,
			CookieSameSite: "Strict",
		}),
	))

## 3. CSP 策略应逐步收紧

开发阶段使用宽松策略：

	security.WithContentSecurityPolicy("default-src 'self' 'unsafe-inline' 'unsafe-eval'")

生产环境逐步收紧：

	security.WithContentSecurityPolicy("default-src 'self'; script-src 'self' https://cdn.example.com")

## 4. 主机白名单应包含所有合法域名

	security.WithAllowedHosts(
		"example.com",
		"www.example.com",
		"api.example.com",
		"admin.example.com",
	)

## 5. XSS/SQL 注入防护是辅助手段

这些模式检测不能替代参数验证和输出转义：

	// 仍需在业务层进行验证
	func createUser(ctx httpx.Context) error {
		name := ctx.FormValue("name")
		if containsHTML(name) {
			return ctx.String(400, "Invalid name")
		}
		// ...
	}

## 6. CSRF token 应定期轮换

	security.WithCSRFProtection(security.CSRFConfig{
		CookieMaxAge: 3600, // 1 小时后过期，强制重新生成
	})

# 性能考虑

  - 安全响应头：几乎无性能开销（仅设置 HTTP 头）
  - 主机验证：O(1) 查找，纳秒级开销
  - XSS/SQL 注入检测：需要遍历所有请求参数，对于大量参数的请求有一定开销
  - CSRF 保护：需要生成随机 token 和验证，微秒级开销

建议：
  - 对于高性能 API，可以禁用 XSS/SQL 注入检测，在业务层进行验证
  - CSRF 保护主要用于 Web 应用，纯 API 服务可以使用其他认证方式（如 JWT）

# 线程安全

中间件是无状态的，可以安全地在多个 goroutine 中并发使用。
*/
package security
