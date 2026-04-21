/*
Package auth 提供基于 Bearer Token 的 HTTP 认证中间件。

# 功能特性

  - 从 Authorization 头提取并验证 Bearer token
  - 将认证声明（Claims）注入请求上下文
  - 实时用户状态检查（如封禁、禁用），每次请求立即生效
  - 完全跳过认证的路径（WithSkipPaths）
  - 可选认证路径：有 token 则验证，无 token 则放行（WithOptionalPaths）
  - 自定义 token 提取、错误处理和用户检查失败处理

# 快速开始

基本用法：

	app := httpx.New()

	app.Use(auth.New(
		auth.WithValidator(myTokenValidator),
		auth.WithSkipPaths("/login", "/register", "/public/*"),
	))

	func protectedHandler(ctx httpx.Context) error {
		uid := auth.GetUID(ctx)
		return ctx.JSON(200, map[string]any{"uid": uid})
	}

# 可选认证

适用于同一路由需要同时服务已登录和未登录用户的场景：

	app.Use(auth.New(
		auth.WithValidator(myValidator),
		auth.WithOptionalPaths("/feed", "/explore/*"),
	))

	func feedHandler(ctx httpx.Context) error {
		uid := auth.GetUID(ctx) // 已登录返回 uid，未登录返回 0
		return ctx.JSON(200, buildFeed(uid))
	}

# 实时用户状态检查

在 token 验证通过后，对每次请求额外检查用户状态：

	app.Use(auth.New(
		auth.WithValidator(myValidator),
		auth.WithUserChecker(myUserChecker),
		auth.WithCheckHandler(func(ctx httpx.Context, err error) {
			ctx.JSON(403, map[string]string{"error": err.Error()})
		}),
	))

# 获取认证信息

	uid := auth.GetUID(ctx)           // 未认证时返回 0
	claims := auth.GetClaims(ctx)     // 未认证时返回 nil
	role, _ := claims.Get("role")
*/
package auth
