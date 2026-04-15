package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/f2xme/gox/graceful"
)

func main() {
	fmt.Println("=== graceful 包使用示例 ===")
	fmt.Println("提示：按 Ctrl+C 触发优雅关闭")

	// 创建优雅关闭管理器
	mgr := graceful.New()

	// 示例 1: 注册 HTTP 服务器
	fmt.Println("\n启动 HTTP 服务器...")
	httpServer := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, World!")
		}),
	}

	// 在后台启动 HTTP 服务器
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP 服务器错误: %v", err)
		}
	}()

	// 注册 HTTP 服务器（优先级 100，超时 5 秒）
	mgr.Register("http-server", graceful.HTTPServer(httpServer),
		graceful.WithPriority(100),
		graceful.WithResourceTimeout(5*time.Second),
	)
	fmt.Println("✓ HTTP 服务器已注册（优先级 100）")

	// 示例 2: 注册自定义资源
	fmt.Println("\n注册自定义资源...")

	mgr.Register("database", graceful.CloserFunc(func(ctx context.Context) error {
		fmt.Println("  → 关闭数据库连接...")
		time.Sleep(1 * time.Second)
		fmt.Println("  ✓ 数据库连接已关闭")
		return nil
	}),
		graceful.WithPriority(50),
		graceful.WithResourceTimeout(3*time.Second),
	)
	fmt.Println("✓ 数据库已注册（优先级 50）")

	mgr.Register("redis", graceful.CloserFunc(func(ctx context.Context) error {
		fmt.Println("  → 关闭 Redis 连接...")
		time.Sleep(500 * time.Millisecond)
		fmt.Println("  ✓ Redis 连接已关闭")
		return nil
	}),
		graceful.WithPriority(50),
		graceful.WithResourceTimeout(3*time.Second),
	)
	fmt.Println("✓ Redis 已注册（优先级 50）")

	mgr.Register("worker", graceful.CloserFunc(func(ctx context.Context) error {
		fmt.Println("  → 停止后台任务...")
		time.Sleep(2 * time.Second)
		fmt.Println("  ✓ 后台任务已停止")
		return nil
	}),
		graceful.WithPriority(80),
		graceful.WithResourceTimeout(5*time.Second),
	)
	fmt.Println("✓ 后台任务已注册（优先级 80）")

	// 模拟日志系统
	mgr.Register("logger", graceful.CloserFunc(func(ctx context.Context) error {
		fmt.Println("  → 刷新日志缓冲区...")
		time.Sleep(300 * time.Millisecond)
		fmt.Println("  ✓ 日志系统已关闭")
		return nil
	}),
		graceful.WithPriority(10),
		graceful.WithResourceTimeout(2*time.Second),
	)
	fmt.Println("✓ 日志系统已注册（优先级 10）")

	// 示例 3: 使用 GenericCloser 适配器
	fmt.Println("\n注册通用资源...")
	mgr.Register("cache", graceful.GenericCloser(func() error {
		fmt.Println("  → 清理缓存...")
		time.Sleep(500 * time.Millisecond)
		fmt.Println("  ✓ 缓存已清理")
		return nil
	}),
		graceful.WithPriority(40),
		graceful.WithResourceTimeout(2*time.Second),
	)
	fmt.Println("✓ 缓存已注册（优先级 40）")

	fmt.Println("\n所有资源已注册，应用正在运行...")
	fmt.Println("访问 http://localhost:8080 测试 HTTP 服务器")
	fmt.Println("\n关闭顺序（按优先级从高到低）：")
	fmt.Println("  1. HTTP 服务器（优先级 100）- 停止接收新请求")
	fmt.Println("  2. 后台任务（优先级 80）- 停止处理任务")
	fmt.Println("  3. 数据库和 Redis（优先级 50）- 关闭数据层连接")
	fmt.Println("  4. 缓存（优先级 40）- 清理缓存")
	fmt.Println("  5. 日志系统（优先级 10）- 最后关闭，确保所有日志被记录")

	// 等待关闭信号
	fmt.Println("\n等待关闭信号（SIGTERM 或 SIGINT）...")
	if err := mgr.Wait(); err != nil {
		log.Printf("优雅关闭出错: %v", err)
	}

	fmt.Println("\n=== 应用已完全关闭 ===")
}
