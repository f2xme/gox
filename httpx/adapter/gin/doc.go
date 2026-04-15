// Package gin 提供基于 Gin 框架的 httpx.Engine 实现。
//
// 本包将 Gin 框架适配为 httpx 统一接口，提供类型安全的路由、中间件和错误处理。
//
// # 功能特性
//
//   - 统一接口：实现 httpx.Engine 接口，提供框架无关的 HTTP 服务
//   - 中间件支持：支持全局和路由级别的中间件
//   - 错误处理：统一的错误处理机制
//   - 路由分组：支持路由分组和嵌套分组
//   - 静态文件：支持静态文件和目录服务
//   - 配置集成：支持从 config.Config 加载配置
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"github.com/f2xme/gox/httpx"
//		ginadapter "github.com/f2xme/gox/httpx/adapter/gin"
//	)
//
//	func main() {
//		engine := ginadapter.New()
//		engine.GET("/hello", func(c httpx.Context) error {
//			return c.JSON(200, map[string]string{"message": "Hello"})
//		})
//		engine.Start(":8080")
//	}
//
// # 配置选项
//
// 使用 Options 模式配置：
//
//	engine := ginadapter.New(
//		ginadapter.WithMode("debug"),
//	)
//
// 从配置文件加载：
//
//	engine := ginadapter.NewWithConfig(cfg)
//
// 配置键：
//   - httpx.gin.mode: Gin 运行模式（debug/release/test，默认 release）
//
// # 中间件和错误处理
//
// 使用中间件：
//
//	engine.Use(loggingMiddleware, authMiddleware)
//	engine.GET("/api/users", getUsersHandler)
//
// 自定义错误处理：
//
//	engine.SetErrorHandler(func(c httpx.Context, err error) {
//		c.JSON(500, map[string]string{"error": err.Error()})
//	})
//
// # 路由分组
//
// 创建路由组：
//
//	api := engine.Group("/api")
//	api.GET("/users", getUsersHandler)
//	api.POST("/users", createUserHandler)
//
//	v1 := api.Group("/v1")
//	v1.GET("/products", getProductsHandler)
package gin
