// Package mock 提供用于测试 httpx 应用的 mock 对象。
//
// 基于 httpx.Context 接口实现的 MockContext，提供完整的请求/响应模拟能力，
// 适用于单元测试和集成测试场景。
//
// # 功能特性
//
//   - MockContext: 完整实现 httpx.Context 接口的 mock 对象
//   - 可导出字段: 所有字段均可直接访问和修改，方便测试
//   - 请求模拟: 支持设置 HTTP 方法、路径、查询参数、请求头等
//   - 响应记录: 自动记录响应状态码、响应体和响应头
//   - 上下文存储: 支持 Set/Get 操作，模拟中间件数据传递
//   - 零依赖: 不需要启动真实的 HTTP 服务器
//
// # 快速开始
//
// 基本使用：
//
//	package myapp
//
//	import (
//	    "testing"
//	    "github.com/f2xme/gox/httpx"
//	    "github.com/f2xme/gox/httpx/mock"
//	)
//
//	func TestMyHandler(t *testing.T) {
//	    // 创建 mock context
//	    ctx := mock.NewMockContext("GET", "/api/users")
//
//	    // 定义 handler
//	    handler := func(ctx httpx.Context) error {
//	        return ctx.JSON(200, map[string]string{"status": "ok"})
//	    }
//
//	    // 执行 handler
//	    if err := handler(ctx); err != nil {
//	        t.Fatal(err)
//	    }
//
//	    // 验证结果
//	    if ctx.RespCode != 200 {
//	        t.Errorf("expected 200, got %d", ctx.RespCode)
//	    }
//	}
//
// # 设置请求参数
//
// 模拟完整的 HTTP 请求：
//
//	ctx := mock.NewMockContext("POST", "/api/users/:id")
//
//	// 设置路径参数
//	ctx.PathParams["id"] = "123"
//
//	// 设置查询参数
//	ctx.QueryParams["page"] = "1"
//	ctx.QueryParams["limit"] = "10"
//
//	// 设置请求头
//	ctx.Headers["Authorization"] = "Bearer token"
//	ctx.Headers["Content-Type"] = "application/json"
//
//	// 设置客户端 IP
//	ctx.ClientIPValue = "192.168.1.100"
//
// # 测试 Middleware
//
// 验证中间件行为：
//
//	func TestAuthMiddleware(t *testing.T) {
//	    middleware := func(next httpx.Handler) httpx.Handler {
//	        return func(ctx httpx.Context) error {
//	            token := ctx.Header("Authorization")
//	            if token == "" {
//	                return ctx.JSON(401, map[string]string{"error": "unauthorized"})
//	            }
//	            ctx.Set("user_id", "123")
//	            return next(ctx)
//	        }
//	    }
//
//	    handler := func(ctx httpx.Context) error {
//	        userID := ctx.MustGet("user_id")
//	        return ctx.JSON(200, map[string]any{"user_id": userID})
//	    }
//
//	    wrappedHandler := middleware(handler)
//
//	    // 测试无 token 的情况
//	    ctx := mock.NewMockContext("GET", "/api/protected")
//	    wrappedHandler(ctx)
//	    if ctx.RespCode != 401 {
//	        t.Error("expected 401")
//	    }
//
//	    // 测试有 token 的情况
//	    ctx2 := mock.NewMockContext("GET", "/api/protected")
//	    ctx2.Headers["Authorization"] = "Bearer token"
//	    wrappedHandler(ctx2)
//	    if ctx2.RespCode != 200 {
//	        t.Error("expected 200")
//	    }
//	}
//
// # 验证响应
//
// 检查响应内容：
//
//	ctx := mock.NewMockContext("POST", "/api/users")
//	handler(ctx)
//
//	// 验证状态码
//	if ctx.RespCode != 200 {
//	    t.Errorf("expected 200, got %d", ctx.RespCode)
//	}
//
//	// 验证响应体
//	resp, ok := ctx.RespBody.(*httpx.Response)
//	if !ok {
//	    t.Fatal("expected *httpx.Response")
//	}
//	if !resp.Success {
//	    t.Error("expected success=true")
//	}
//
//	// 验证响应头
//	if ctx.RespHeaders["X-Request-ID"] == "" {
//	    t.Error("expected X-Request-ID header")
//	}
//
// # 注意事项
//
//   - MockContext 不是线程安全的，每个测试应创建独立的实例
//   - Bind 系列方法总是返回 nil，如需测试绑定逻辑请使用真实的 HTTP 请求
//   - Cookie 方法总是返回 http.ErrNoCookie
//   - Raw 方法总是返回 nil
package mock
