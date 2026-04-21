package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/f2xme/gox/httpx"
	ginadapter "github.com/f2xme/gox/httpx/adapter/gin"
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
	// 创建 HTTP 引擎（使用 Gin 适配器）
	engine := ginadapter.New()

	// 注册路由
	engine.GET("/", handleHome)
	engine.GET("/users/:id", handleGetUser)
	engine.POST("/users", handleCreateUser)
	engine.GET("/error", handleError)

	// 创建路由组
	api := engine.Group("/api/v1")
	api.GET("/health", handleHealth)

	// 启动服务器（使用 goroutine 以便优雅关闭）
	addr := ":8080"
	go func() {
		fmt.Printf("HTTP 服务器启动在 %s\n", addr)
		fmt.Println("访问 http://localhost:8080 查看示例")
		if err := engine.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n正在关闭服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := engine.Shutdown(ctx); err != nil {
		log.Fatalf("服务器强制关闭: %v", err)
	}

	fmt.Println("服务器已关闭")
}

// handleHome 处理首页请求
func handleHome(ctx httpx.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "欢迎使用 gox/httpx 示例",
		"version": "1.0.0",
	})
}

// handleGetUser 获取用户信息（演示路径参数）
func handleGetUser(ctx httpx.Context) error {
	// 获取路径参数
	id := ctx.Param("id")

	// 模拟数据库查询
	user := User{
		ID:   1,
		Name: "张三",
		Age:  25,
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"id":   id,
		"user": user,
	})
}

// handleCreateUser 创建用户（演示请求体绑定和验证）
func handleCreateUser(ctx httpx.Context) error {
	var req CreateUserRequest

	// 绑定并验证 JSON 请求体
	if err := ctx.BindJSON(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "请求参数无效: " + err.Error(),
		})
	}

	// 模拟创建用户
	user := User{
		ID:   100,
		Name: req.Name,
		Age:  req.Age,
	}

	// 使用统一响应格式
	return httpx.Success(ctx, user)
}

// handleError 演示错误处理
func handleError(ctx httpx.Context) error {
	// 使用统一失败响应
	return httpx.Fail(ctx, "这是一个示例错误")
}

// handleHealth 健康检查
func handleHealth(ctx httpx.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
