// Package graceful 提供应用程序的优雅关闭管理功能。
//
// 帮助管理应用程序关闭时的资源生命周期（服务器、数据库、连接等），
// 确保所有资源按正确的顺序优雅关闭，并支持超时控制。
//
// # 功能特性
//
//   - 基于优先级的资源注册和关闭顺序
//   - 信号处理（SIGTERM、SIGINT）
//   - 每个资源的超时控制
//   - 常用资源的内置适配器（HTTP/gRPC 服务器、数据库）
//   - 自定义关闭逻辑的钩子函数
//   - 并发安全的资源管理
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"log"
//		"time"
//
//		"github.com/f2xme/gox/graceful"
//	)
//
//	func main() {
//		// 创建管理器
//		mgr := graceful.New()
//
//		// 注册 HTTP 服务器（优先级 100，先关闭）
//		mgr.Register("http", graceful.HTTPServer(httpServer),
//			graceful.WithPriority(100),
//			graceful.WithResourceTimeout(5*time.Second),
//		)
//
//		// 注册数据库连接（优先级 50）
//		mgr.Register("database", graceful.DBCloser(db),
//			graceful.WithPriority(50),
//			graceful.WithResourceTimeout(3*time.Second),
//		)
//
//		// 等待关闭信号（SIGTERM 或 SIGINT）
//		if err := mgr.Wait(); err != nil {
//			log.Fatal(err)
//		}
//	}
//
// # 优先级控制
//
// 资源按优先级从高到低关闭：
//
//	// 优先级 100：先关闭 HTTP 服务器（停止接收新请求）
//	mgr.Register("http", closer, graceful.WithPriority(100))
//
//	// 优先级 50：再关闭数据库和缓存
//	mgr.Register("database", closer, graceful.WithPriority(50))
//	mgr.Register("redis", closer, graceful.WithPriority(50))
//
//	// 优先级 10：最后关闭日志
//	mgr.Register("logger", closer, graceful.WithPriority(10))
//
// # 自定义 Closer
//
// 使用 CloserFunc 适配器创建自定义关闭逻辑：
//
//	mgr.Register("custom", graceful.CloserFunc(func(ctx context.Context) error {
//		log.Println("cleaning up...")
//		// 执行清理逻辑
//		return nil
//	}))
//
// # 内置适配器
//
//	// HTTP 服务器
//	graceful.HTTPServer(httpServer)
//
//	// 数据库连接
//	graceful.DBCloser(db)
//
//	// 通用 io.Closer
//	graceful.IOCloser(closer)
//
//	// 通用函数
//	graceful.GenericCloser(func() error { return nil })
//
// # 手动触发关闭
//
// 不等待信号，立即关闭：
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	if err := mgr.Shutdown(ctx); err != nil {
//		log.Printf("shutdown error: %v", err)
//	}
//
// # 关闭钩子
//
// 在关闭流程的不同阶段执行自定义逻辑：
//
//	mgr := graceful.New(
//		graceful.OnBeforeShutdown(func() {
//			log.Println("开始关闭...")
//		}),
//		graceful.OnAfterShutdown(func() {
//			log.Println("关闭完成")
//		}),
//		graceful.OnTimeout(func(name string) {
//			log.Printf("资源 %s 关闭超时", name)
//		}),
//	)
//
// # 注意事项
//
//   - Manager 是并发安全的，可以在多个 goroutine 中注册资源
//   - 默认监听 SIGTERM 和 SIGINT 信号
//   - 资源关闭超时不会中断流程，会继续关闭其他资源
//   - 建议为不同类型的资源设置合理的优先级和超时时间
package graceful
