package mock_test

import (
	"fmt"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/mock"
)

// 示例：测试一个简单的 handler
func ExampleMockContext_basic() {
	// 创建 mock context
	ctx := mock.NewMockContext("GET", "/api/users")

	// 定义 handler
	handler := func(ctx httpx.Context) error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	}

	// 执行 handler
	handler(ctx)

	// 验证结果
	fmt.Println(ctx.RespCode)
	// Output: 200
}

// 示例：测试带查询参数的请求
func ExampleMockContext_queryParams() {
	ctx := mock.NewMockContext("GET", "/api/search")
	ctx.QueryParams["q"] = "golang"
	ctx.QueryParams["page"] = "2"

	handler := func(ctx httpx.Context) error {
		query := ctx.Query("q")
		page := ctx.QueryDefault("page", "1")
		return ctx.JSON(200, map[string]string{
			"query": query,
			"page":  page,
		})
	}

	handler(ctx)
	fmt.Println(ctx.RespCode)
	// Output: 200
}

// 示例：测试带路径参数的请求
func ExampleMockContext_pathParams() {
	ctx := mock.NewMockContext("GET", "/api/users/:id")
	ctx.PathParams["id"] = "123"

	handler := func(ctx httpx.Context) error {
		id := ctx.Param("id")
		return ctx.JSON(200, map[string]string{"user_id": id})
	}

	handler(ctx)
	fmt.Println(ctx.RespCode)
	// Output: 200
}

// 示例：测试带请求头的请求
func ExampleMockContext_headers() {
	ctx := mock.NewMockContext("GET", "/api/protected")
	ctx.Headers["Authorization"] = "Bearer token123"

	handler := func(ctx httpx.Context) error {
		token := ctx.Header("Authorization")
		if token == "" {
			return ctx.JSON(401, map[string]string{"error": "unauthorized"})
		}
		return ctx.Success(map[string]string{"message": "authorized"})
	}

	handler(ctx)
	fmt.Println(ctx.RespCode)
	// Output: 200
}

// 示例：测试 middleware
func ExampleMockContext_middleware() {
	// 定义一个简单的 middleware
	middleware := func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			// 在请求前设置一个值
			ctx.Set("request_id", "abc123")
			// 调用下一个 handler
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

	// 执行测试
	ctx := mock.NewMockContext("GET", "/api/test")
	wrappedHandler(ctx)

	fmt.Println(ctx.RespCode)
	// Output: 200
}

// 示例：测试错误响应
func ExampleMockContext_errorResponses() {
	ctx := mock.NewMockContext("GET", "/api/users/999")

	handler := func(ctx httpx.Context) error {
		// 模拟用户不存在
		return ctx.JSON(404, httpx.NewFailResponse("用户不存在"))
	}

	handler(ctx)
	fmt.Println(ctx.RespCode)
	// Output: 404
}

// 示例：测试统一响应格式
func ExampleMockContext_unifiedResponse() {
	ctx := mock.NewMockContext("POST", "/api/users")

	handler := func(ctx httpx.Context) error {
		// 模拟创建成功
		return ctx.Success(map[string]any{
			"id":   123,
			"name": "张三",
		})
	}

	handler(ctx)

	resp := ctx.RespBody.(*httpx.Response)
	fmt.Println(resp.Success)
	fmt.Println(resp.Message)
	// Output:
	// true
	// ok
}
