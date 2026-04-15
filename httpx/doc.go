/*
Package httpx 提供统一的 HTTP 框架抽象层。

# 概述

httpx 包定义了 HTTP 服务器的标准接口，支持多种 HTTP 框架（Gin、Echo、Fiber 等）。
通过这些接口，你可以轻松地在不同的 HTTP 框架之间切换，而无需修改业务代码。

# 核心接口

## Engine - HTTP 服务器引擎

所有 HTTP 框架实现都必须实现此接口：

	type Engine interface {
		Router
		Start(addr string) error
		Shutdown(ctx context.Context) error
		SetErrorHandler(h ErrorHandler)
		SetRenderer(r Renderer)
		Raw() any
	}

## Router - 路由接口

定义路由注册方法：

	type Router interface {
		GET(path string, handler HandlerFunc)
		POST(path string, handler HandlerFunc)
		PUT(path string, handler HandlerFunc)
		DELETE(path string, handler HandlerFunc)
		PATCH(path string, handler HandlerFunc)
		Group(prefix string) Router
		Use(middleware ...MiddlewareFunc)
	}

## Context - 请求上下文

提供请求和响应的操作方法：

	type Context interface {
		Request() *http.Request
		Response() http.ResponseWriter
		Param(key string) string
		Query(key string) string
		Bind(v any) error
		JSON(code int, v any) error
		String(code int, s string) error
		Set(key string, value any)
		Get(key string) any
	}

# 使用示例

## 创建服务器

	import "github.com/f2xme/gox/httpx/adapter/gin"

	engine := gin.New()

	engine.GET("/users/:id", func(c httpx.Context) error {
		id := c.Param("id")
		return c.JSON(200, map[string]string{"id": id})
	})

	engine.Start(":8080")

## 路由分组

	api := engine.Group("/api/v1")
	api.GET("/users", listUsers)
	api.POST("/users", createUser)

	admin := api.Group("/admin")
	admin.Use(authMiddleware)
	admin.DELETE("/users/:id", deleteUser)

## 中间件

	// 全局中间件
	engine.Use(loggerMiddleware, recoveryMiddleware)

	// 路由组中间件
	api := engine.Group("/api")
	api.Use(authMiddleware)

## 错误处理

	engine.SetErrorHandler(func(c httpx.Context, err error) {
		if errorx.IsKind(err, errorx.KindValidation) {
			c.JSON(400, map[string]string{"error": err.Error()})
			return
		}
		c.JSON(500, map[string]string{"error": "internal error"})
	})

## 自定义渲染器

	engine.SetRenderer(&customRenderer{})

# 可用适配器

## Gin 适配器

	import "github.com/f2xme/gox/httpx/adapter/gin"

	engine := gin.New()

## Echo 适配器

	import "github.com/f2xme/gox/httpx/adapter/echoadapter"

	engine := echoadapter.New()

# 最佳实践

## 1. 使用统一的错误处理

	engine.SetErrorHandler(func(c httpx.Context, err error) {
		statusCode := errorx.HTTPStatus(err)
		c.JSON(statusCode, map[string]string{
			"error": err.Error(),
			"code":  errorx.Code(err),
		})
	})

## 2. 使用中间件进行横切关注点

	// 日志中间件
	engine.Use(func(c httpx.Context) error {
		start := time.Now()
		err := c.Next()
		log.Printf("%s %s %v", c.Request().Method, c.Request().URL.Path, time.Since(start))
		return err
	})

## 3. 使用路由分组组织 API

	v1 := engine.Group("/api/v1")
	{
		users := v1.Group("/users")
		users.GET("", listUsers)
		users.POST("", createUser)
		users.GET("/:id", getUser)
	}

## 4. 优雅关闭

	go func() {
		if err := engine.Start(":8080"); err != nil {
			log.Fatal(err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	engine.Shutdown(ctx)

# 线程安全

所有 HTTP 框架实现都应该是线程安全的，可以在多个 goroutine 中并发处理请求。
*/
package httpx
