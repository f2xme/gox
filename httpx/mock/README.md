# httpx/mock

`httpx/mock` 包提供用于测试 `httpx` 应用的 mock 对象。

## 安装

```bash
go get github.com/f2xme/gox/httpx/mock
```

## 快速开始

### 基本用法

```go
package myapp

import (
    "testing"
    "github.com/f2xme/gox/httpx"
    "github.com/f2xme/gox/httpx/mock"
)

func TestMyHandler(t *testing.T) {
    // 创建 mock context
    ctx := mock.NewMockContext("GET", "/api/users")
    
    // 定义 handler
    handler := func(ctx httpx.Context) error {
        return ctx.JSON(200, map[string]string{"status": "ok"})
    }
    
    // 执行 handler
    if err := handler(ctx); err != nil {
        t.Fatal(err)
    }
    
    // 验证结果
    if ctx.RespCode != 200 {
        t.Errorf("expected 200, got %d", ctx.RespCode)
    }
}
```

### 设置请求参数

```go
func TestWithParams(t *testing.T) {
    ctx := mock.NewMockContext("GET", "/api/users/:id")
    
    // 设置路径参数
    ctx.PathParams["id"] = "123"
    
    // 设置查询参数
    ctx.QueryParams["page"] = "1"
    ctx.QueryParams["limit"] = "10"
    
    // 设置请求头
    ctx.Headers["Authorization"] = "Bearer token"
    ctx.Headers["Content-Type"] = "application/json"
    
    // 设置客户端 IP
    ctx.ClientIPValue = "192.168.1.100"
    
    handler := func(ctx httpx.Context) error {
        id := ctx.Param("id").String()       // 或 ctx.Param("id").Int64()
        page := ctx.Query("page").IntOr(1)
        token := ctx.Header("Authorization").String()

        return ctx.JSON(200, map[string]any{
            "id":    id,
            "page":  page,
            "token": token,
        })
    }
    
    handler(ctx)
    
    if ctx.RespCode != 200 {
        t.Errorf("expected 200, got %d", ctx.RespCode)
    }
}
```

### 测试 Middleware

```go
func TestMiddleware(t *testing.T) {
    // 定义 middleware
    middleware := func(next httpx.Handler) httpx.Handler {
        return func(ctx httpx.Context) error {
            ctx.Set("request_id", "abc123")
            return next(ctx)
        }
    }
    
    // 定义 handler
    handler := func(ctx httpx.Context) error {
        requestID := ctx.MustGet("request_id")
        return ctx.JSON(200, map[string]any{"request_id": requestID})
    }
    
    // 应用 middleware
    wrappedHandler := middleware(handler)
    
    // 测试
    ctx := mock.NewMockContext("GET", "/test")
    if err := wrappedHandler(ctx); err != nil {
        t.Fatal(err)
    }
    
    if ctx.RespCode != 200 {
        t.Errorf("expected 200, got %d", ctx.RespCode)
    }
}
```

### 验证响应

```go
func TestResponse(t *testing.T) {
    ctx := mock.NewMockContext("POST", "/api/users")
    
    handler := func(ctx httpx.Context) error {
        return ctx.Success(map[string]any{
            "id":   123,
            "name": "张三",
        })
    }
    
    handler(ctx)
    
    // 验证状态码
    if ctx.RespCode != 200 {
        t.Errorf("expected 200, got %d", ctx.RespCode)
    }
    
    // 验证响应体
    resp, ok := ctx.RespBody.(*httpx.Response)
    if !ok {
        t.Fatal("expected *httpx.Response")
    }
    
    if !resp.Success {
        t.Error("expected success=true")
    }
    
    if resp.Message != "ok" {
        t.Errorf("expected message=ok, got %s", resp.Message)
    }
    
    // 验证响应头
    if ctx.RespHeaders["X-Custom"] != "value" {
        t.Error("expected X-Custom header")
    }
}
```

### 表驱动测试

```go
func TestHandlerCases(t *testing.T) {
    tests := []struct {
        name     string
        method   string
        path     string
        setup    func(*mock.MockContext)
        wantCode int
    }{
        {
            name:     "success",
            method:   "GET",
            path:     "/api/users",
            setup:    func(ctx *mock.MockContext) {},
            wantCode: 200,
        },
        {
            name:   "with auth",
            method: "GET",
            path:   "/api/protected",
            setup: func(ctx *mock.MockContext) {
                ctx.Headers["Authorization"] = "Bearer token"
            },
            wantCode: 200,
        },
        {
            name:     "not found",
            method:   "GET",
            path:     "/api/missing",
            setup:    func(ctx *mock.MockContext) {},
            wantCode: 404,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := mock.NewMockContext(tt.method, tt.path)
            tt.setup(ctx)
            
            handler(ctx)
            
            if ctx.RespCode != tt.wantCode {
                t.Errorf("code = %d, want %d", ctx.RespCode, tt.wantCode)
            }
        })
    }
}
```

## MockContext 字段

`MockContext` 的所有字段都是可导出的，方便在测试中直接访问和修改：

### 请求信息
- `MethodValue` - HTTP 方法
- `PathValue` - 请求路径
- `ClientIPValue` - 客户端 IP
- `Headers` - 请求头
- `QueryParams` - 查询参数
- `PathParams` - 路径参数

### 响应信息
- `RespCode` - 响应状态码
- `RespBody` - 响应体
- `RespHeaders` - 响应头

### 上下文存储
- `Store` - 上下文键值存储

### 底层对象
- `Req` - 底层 `*http.Request`（可选）
- `Writer` - 底层 `http.ResponseWriter`（可选）

## 最佳实践

1. **使用表驱动测试** - 覆盖多种场景，代码简洁
2. **测试隔离** - 每个测试创建独立的 MockContext
3. **验证关键路径** - 测试正常和异常情况
4. **Mock 最小化** - 只 mock 必要的部分

## 参考

- [httpx 文档](../README.md)
- [示例代码](example_test.go)
