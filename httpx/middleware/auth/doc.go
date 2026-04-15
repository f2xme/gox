/*
Package auth 提供 JWT 认证中间件，用于保护 HTTP 端点。

# 概述

auth 包实现了基于 Bearer Token 的认证中间件，支持：
  - 从 Authorization 头提取和验证 JWT token
  - 将认证信息注入请求上下文
  - 实时用户状态检查（如封禁、禁用）
  - 灵活的路径跳过规则
  - 自定义错误处理

# 快速开始

基本用法：

	import "github.com/f2xme/gox/httpx/middleware/auth"

	func main() {
		app := httpx.New()

		// 创建认证中间件
		app.Use(auth.New(
			auth.WithValidator(myTokenValidator),
		))

		app.GET("/protected", protectedHandler)
		app.Run(":8080")
	}

	func protectedHandler(ctx httpx.Context) error {
		claims := auth.GetClaims(ctx)
		userID := claims.GetSubject()
		return ctx.JSON(200, map[string]string{"user_id": userID})
	}

# 核心接口

## TokenValidator - Token 验证器

实现此接口以提供 token 验证逻辑：

	type TokenValidator interface {
		Validate(token string) (Claims, error)
	}

示例实现：

	type JWTValidator struct {
		secret []byte
	}

	func (v *JWTValidator) Validate(token string) (auth.Claims, error) {
		// 解析和验证 JWT token
		// 返回 Claims 实现
	}

## Claims - 认证声明

表示已认证的 token 声明：

	type Claims interface {
		GetSubject() string
		Get(key string) (any, bool)
	}

## UserStatusChecker - 用户状态检查器

可选接口，用于实时检查用户状态（每次请求都会调用）：

	type UserStatusChecker interface {
		IsBanned(userID string) (bool, error)
	}

# 配置选项

## WithValidator

设置 Token 验证器（必需）：

	app.Use(auth.New(
		auth.WithValidator(myValidator),
	))

## WithSkipPaths

跳过认证的路径

支持精确匹配和通配符模式：

	app.Use(auth.New(
		auth.WithValidator(myValidator),
		auth.WithSkipPaths(
			"/login",           // 精确匹配
			"/register",
			"/public/*",        // 通配符匹配
			"/api/v1/health",
		),
	))

## WithUserStatusChecker

实时用户状态检查，用于立即生效的封禁或禁用操作：

	app.Use(auth.New(
		auth.WithValidator(myValidator),
		auth.WithUserStatusChecker(myStatusChecker),
	))

注意：如果状态检查器不可用，中间件会 fail-open（允许请求通过），以防止级联故障。

## WithErrorHandler

自定义未授权错误处理：

	app.Use(auth.New(
		auth.WithValidator(myValidator),
		auth.WithErrorHandler(func(ctx httpx.Context) {
			ctx.JSON(401, map[string]string{
				"error": "unauthorized",
				"message": "Invalid or missing token",
			})
		}),
	))

## WithBanHandler

自定义封禁用户错误处理：

	app.Use(auth.New(
		auth.WithValidator(myValidator),
		auth.WithUserStatusChecker(myStatusChecker),
		auth.WithBanHandler(func(ctx httpx.Context) {
			ctx.JSON(403, map[string]string{
				"error": "forbidden",
				"message": "Your account has been suspended",
			})
		}),
	))

## WithTokenExtractor

自定义 Token 提取逻辑，默认从 Authorization 头提取 Bearer token：

	app.Use(auth.New(
		auth.WithValidator(myValidator),
		auth.WithTokenExtractor(func(ctx httpx.Context) string {
			// 从 cookie 提取
			return ctx.Cookie("auth_token")
		}),
	))

# 获取认证信息

使用 GetClaims 从上下文中获取认证声明：

	func handler(ctx httpx.Context) error {
		claims := auth.GetClaims(ctx)
		if claims == nil {
			// 未认证（不应该发生在受保护的路由中）
			return ctx.JSON(401, "unauthorized")
		}

		userID := claims.GetSubject()
		role, _ := claims.Get("role")

		return ctx.JSON(200, map[string]any{
			"user_id": userID,
			"role": role,
		})
	}

# 完整示例

	import (
		"github.com/f2xme/gox/httpx"
		"github.com/f2xme/gox/httpx/middleware/auth"
	)

	func main() {
		app := httpx.New()

		// 配置认证中间件
		app.Use(auth.New(
			auth.WithValidator(&MyJWTValidator{secret: []byte("secret")}),
			auth.WithUserStatusChecker(&MyUserStatusChecker{}),
			auth.WithSkipPaths("/login", "/register", "/public/*"),
			auth.WithErrorHandler(func(ctx httpx.Context) {
				ctx.JSON(401, map[string]string{"error": "unauthorized"})
			}),
		))

		// 公开路由
		app.POST("/login", loginHandler)
		app.POST("/register", registerHandler)

		// 受保护路由
		app.GET("/profile", profileHandler)
		app.POST("/posts", createPostHandler)

		app.Run(":8080")
	}

	func profileHandler(ctx httpx.Context) error {
		claims := auth.GetClaims(ctx)
		userID := claims.GetSubject()
		// 查询用户信息
		return ctx.JSON(200, map[string]string{"user_id": userID})
	}

# 最佳实践

## 1. 始终设置 Token 验证器

验证器是必需的，否则所有请求都会被拒绝：

	// 错误：缺少验证器
	app.Use(auth.New())

	// 正确
	app.Use(auth.New(
		auth.WithValidator(myValidator),
	))

## 2. 合理配置跳过路径

公开端点应明确跳过认证：

	auth.WithSkipPaths(
		"/login",
		"/register",
		"/health",
		"/metrics",
		"/public/*",
	)

## 3. 实时状态检查的性能考虑

UserStatusChecker 在每次请求时都会调用，应确保：
  - 使用缓存减少数据库查询
  - 设置合理的超时时间
  - 实现 fail-open 策略（检查器不可用时允许请求）

示例：

	type CachedStatusChecker struct {
		cache cache.Cache
		db    *sql.DB
	}

	func (c *CachedStatusChecker) IsBanned(userID string) (bool, error) {
		// 先查缓存
		if cached, err := c.cache.Get(ctx, "ban:"+userID); err == nil {
			return string(cached) == "true", nil
		}

		// 缓存未命中，查数据库
		var banned bool
		err := c.db.QueryRow("SELECT banned FROM users WHERE id = ?", userID).Scan(&banned)
		if err != nil {
			return false, err
		}

		// 写入缓存（TTL 1 分钟）
		c.cache.Set(ctx, "ban:"+userID, []byte(fmt.Sprint(banned)), time.Minute)
		return banned, nil
	}

## 4. 自定义错误响应

根据 API 规范自定义错误格式：

	auth.WithErrorHandler(func(ctx httpx.Context) {
		ctx.JSON(401, map[string]any{
			"code": "UNAUTHORIZED",
			"message": "Authentication required",
			"timestamp": time.Now().Unix(),
		})
	})

## 5. 使用通配符优化路径匹配

	// 不推荐：逐个列出
	auth.WithSkipPaths(
		"/public/images/logo.png",
		"/public/images/banner.png",
		"/public/css/style.css",
		// ...
	)

	// 推荐：使用通配符
	auth.WithSkipPaths(
		"/public/*",
	)

# 安全考虑

## 1. Token 验证

确保验证器检查：
  - Token 签名有效性
  - Token 过期时间
  - Token 颁发者（issuer）
  - Token 受众（audience）

## 2. 实时状态检查

对于安全敏感操作，使用 UserStatusChecker 确保封禁立即生效：

	auth.WithUserStatusChecker(myStatusChecker)

## 3. HTTPS

始终在生产环境使用 HTTPS，防止 token 被窃取。

## 4. Token 存储

客户端应安全存储 token：
  - Web: HttpOnly cookie 或 localStorage（注意 XSS 风险）
  - Mobile: 安全存储（Keychain/Keystore）

# 常见问题

## Q: 如何处理 token 刷新？

A: 在业务层实现刷新逻辑，中间件只负责验证当前 token。

## Q: 如何支持多种认证方式（JWT + API Key）？

A: 实现自定义 TokenExtractor 和 Validator：

	auth.WithTokenExtractor(func(ctx httpx.Context) string {
		// 优先尝试 Bearer token
		if token := extractBearerToken(ctx); token != "" {
			return token
		}
		// 回退到 API Key
		return ctx.Header("X-API-Key")
	})

## Q: 如何获取更多用户信息？

A: 在 Claims 实现中存储额外字段：

	type MyClaims struct {
		UserID string
		Role   string
		Email  string
	}

	func (c *MyClaims) GetSubject() string { return c.UserID }
	func (c *MyClaims) Get(key string) (any, bool) {
		switch key {
		case "role": return c.Role, true
		case "email": return c.Email, true
		default: return nil, false
		}
	}

# 性能

  - 路径跳过使用 O(1) 哈希表查找（精确匹配）
  - 通配符模式使用前缀匹配，性能为 O(n)，n 为模式数量
  - 建议优先使用精确匹配，减少通配符模式数量

# 线程安全

中间件本身是无状态的，可以安全地在多个 goroutine 中使用。
*/
package auth
