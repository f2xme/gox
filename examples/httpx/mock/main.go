package main

import (
	"fmt"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/mock"
)

// User 示例用户结构
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name string `json:"name" binding:"required"`
	Age  int    `json:"age" binding:"required,min=1,max=150"`
}

func main() {
	fmt.Println("=== httpx/mock 示例 ===")
	fmt.Println()

	// 示例 1: 测试 GET 请求
	testGetUser()

	// 示例 2: 测试 POST 请求
	testCreateUser()

	// 示例 3: 测试错误响应
	testErrorResponse()

	// 示例 4: 测试统一响应格式
	testUnifiedResponse()
}

// testGetUser 测试 GET 请求处理
func testGetUser() {
	fmt.Println("1. 测试 GET 请求")

	// 创建 mock context
	ctx := mock.NewMockContext("GET", "/api/users/123")
	ctx.PathParams["id"] = "123"
	ctx.QueryParams["detail"] = []string{"true"}
	ctx.Headers["Authorization"] = "Bearer token123"

	// 定义 handler
	handler := func(ctx httpx.Context) error {
		id := ctx.Param("id")
		detail := ctx.Query("detail")
		token := ctx.Header("Authorization")

		user := User{
			ID:   123,
			Name: "张三",
			Age:  25,
		}

		return ctx.JSON(200, map[string]any{
			"id":     id,
			"detail": detail,
			"token":  token,
			"user":   user,
		})
	}

	// 执行 handler
	if err := handler(ctx); err != nil {
		fmt.Printf("   错误: %v\n", err)
		return
	}

	// 验证结果
	fmt.Printf("   状态码: %d\n", ctx.RespCode)
	fmt.Printf("   响应体: %+v\n\n", ctx.RespBody)
}

// testCreateUser 测试 POST 请求处理
func testCreateUser() {
	fmt.Println("2. 测试 POST 请求")

	// 创建 mock context
	ctx := mock.NewMockContext("POST", "/api/users")

	// 定义 handler（mock 中 BindJSON 总是返回 nil）
	handler := func(ctx httpx.Context) error {
		var req CreateUserRequest
		if err := ctx.BindJSON(&req); err != nil {
			return ctx.JSON(400, map[string]string{"error": "请求参数无效"})
		}

		// 模拟创建用户
		user := User{
			ID:   456,
			Name: "李四",
			Age:  30,
		}

		return ctx.JSON(201, user)
	}

	// 执行 handler
	if err := handler(ctx); err != nil {
		fmt.Printf("   错误: %v\n", err)
		return
	}

	// 验证结果
	fmt.Printf("   状态码: %d\n", ctx.RespCode)
	fmt.Printf("   响应体: %+v\n\n", ctx.RespBody)
}

// testErrorResponse 测试错误响应
func testErrorResponse() {
	fmt.Println("3. 测试错误响应")

	// 创建 mock context
	ctx := mock.NewMockContext("GET", "/api/error")

	// 定义 handler（模拟错误）
	handler := func(ctx httpx.Context) error {
		return httpx.ErrNotFound("用户不存在")
	}

	// 执行 handler，错误通过 DefaultErrorHandler 转为响应
	if err := handler(ctx); err != nil {
		httpx.DefaultErrorHandler(ctx, err)
	}

	// 验证结果
	fmt.Printf("   状态码: %d\n", ctx.RespCode)
	fmt.Printf("   响应体: %+v\n\n", ctx.RespBody)
}

// testUnifiedResponse 测试统一响应格式
func testUnifiedResponse() {
	fmt.Println("4. 测试统一响应格式")

	// 测试成功响应
	ctx1 := mock.NewMockContext("GET", "/api/success")
	handler1 := func(ctx httpx.Context) error {
		return httpx.Success(ctx, map[string]string{
			"message": "操作成功",
		})
	}
	handler1(ctx1)

	fmt.Printf("   成功响应:\n")
	fmt.Printf("   状态码: %d\n", ctx1.RespCode)
	if resp, ok := ctx1.RespBody.(*httpx.Response); ok {
		fmt.Printf("   Success: %v\n", resp.Success)
		fmt.Printf("   Message: %s\n", resp.Message)
		fmt.Printf("   Data: %+v\n", resp.Data)
	}

	// 测试失败响应
	ctx2 := mock.NewMockContext("POST", "/api/fail")
	handler2 := func(ctx httpx.Context) error {
		return httpx.Fail(ctx, "操作失败")
	}
	handler2(ctx2)

	fmt.Printf("\n   失败响应:\n")
	fmt.Printf("   状态码: %d\n", ctx2.RespCode)
	if resp, ok := ctx2.RespBody.(*httpx.Response); ok {
		fmt.Printf("   Success: %v\n", resp.Success)
		fmt.Printf("   Message: %s\n", resp.Message)
		fmt.Printf("   Data: %+v\n", resp.Data)
	}
}
