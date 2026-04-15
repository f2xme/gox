// Package mem 提供内存队列实现。
//
// 本包使用 Go channel 作为后端实现了 queue.Queue 接口。
// 支持多个主题和并发订阅者，可配置缓冲区大小。
//
// # 功能特性
//
//   - 进程内消息传递
//   - 零外部依赖
//   - 高性能（纳秒级延迟）
//   - 支持多主题和多订阅者
//   - 线程安全
//   - 可配置缓冲区大小
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"context"
//		"fmt"
//		"log"
//
//		"github.com/f2xme/gox/queue"
//		"github.com/f2xme/gox/queue/adapter/mem"
//	)
//
//	func main() {
//		ctx := context.Background()
//
//		// 创建内存队列
//		q := mem.New()
//		defer q.(queue.Closer).Close()
//
//		// 订阅主题
//		sub, err := q.Subscribe(ctx, "events", func(ctx context.Context, msg *queue.Message) error {
//			fmt.Printf("收到消息: %s\n", msg.Body)
//			return nil
//		})
//		if err != nil {
//			log.Fatal(err)
//		}
//		defer sub.Unsubscribe()
//
//		// 发布消息
//		err = q.Publish(ctx, "events", []byte("hello"))
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//
// # 配置选项
//
// 使用 Option 函数配置：
//
//	q := mem.New(
//		mem.WithBufferSize(128), // 设置通道缓冲区大小
//	)
//
// # 线程安全
//
// 所有操作都是线程安全的，可以从多个 goroutine 并发调用。
// 队列使用读写锁保护内部状态。
//
// # 资源管理
//
// 使用完毕后务必调用 Close() 清理所有订阅并释放资源。
// 订阅不再需要时也应调用 Unsubscribe()。
package mem
