// Package memory 提供基于内存的缓存实现。
//
// # 功能特性
//
//   - 高性能本地缓存（纳秒级延迟）
//   - 支持 LRU/LFU 淘汰策略
//   - 自动过期清理机制
//   - 进程内分布式锁
//   - 零外部依赖
//   - 线程安全
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"context"
//		"time"
//
//		"github.com/f2xme/gox/cache"
//		"github.com/f2xme/gox/cache/adapter/memory"
//	)
//
//	func main() {
//		// 创建内存缓存
//		c, err := memory.New()
//		if err != nil {
//			panic(err)
//		}
//		defer c.(cache.Closer).Close()
//
//		ctx := context.Background()
//
//		// 设置值
//		c.Set(ctx, "key", []byte("value"), 5*time.Minute)
//
//		// 获取值
//		value, err := c.Get(ctx, "key")
//		if err != nil {
//			panic(err)
//		}
//		println(string(value))
//	}
//
// # 配置选项
//
// 限制缓存大小：
//
//	c, _ := memory.New(
//		memory.WithMaxSize(1000), // 最多存储 1000 个条目
//	)
//
// 设置淘汰策略：
//
//	// LRU（最近最少使用，默认）
//	c, _ := memory.New(
//		memory.WithMaxSize(1000),
//		memory.WithEvictionPolicy("lru"),
//	)
//
//	// LFU（最不经常使用）
//	c, _ := memory.New(
//		memory.WithMaxSize(1000),
//		memory.WithEvictionPolicy("lfu"),
//	)
//
// 自定义清理间隔：
//
//	c, _ := memory.New(
//		memory.WithCleanupInterval(5 * time.Minute), // 每 5 分钟清理一次过期条目
//	)
//
// # 使用配置文件
//
// 从 config.Config 创建：
//
//	// 配置文件（YAML）
//	// cache:
//	//   memory:
//	//     maxSize: 1000
//	//     cleanupInterval: 5m
//	//     evictionPolicy: lru
//
//	c, err := memory.NewWithConfig(cfg)
//
// # 分布式锁
//
// memory 包实现了进程内锁（适用于单机应用）：
//
//	locker := c.(cache.Locker)
//
//	// 阻塞式获取锁
//	unlock, err := locker.Lock(ctx, "resource:1", 30*time.Second)
//	if err != nil {
//		panic(err)
//	}
//	defer unlock()
//
//	// 执行需要保护的操作
//	// ...
//
//	// 非阻塞式尝试获取锁
//	unlock, err := locker.TryLock(ctx, "resource:2", 30*time.Second)
//	if err == cache.ErrLocked {
//		// 锁已被占用
//	}
//
// # 资源清理
//
// 内存缓存会启动后台 goroutine 进行清理，使用完毕后必须调用 Close()：
//
//	c, _ := memory.New()
//	defer c.(cache.Closer).Close()
//
// # 注意事项
//
//   - 内存缓存仅适用于单机应用，不支持跨进程共享
//   - 进程重启后缓存数据会丢失
//   - MaxSize 为 0 表示无限制，可能导致内存溢出
//   - 锁仅在当前进程内有效，不适用于分布式场景
//   - 必须调用 Close() 以防止 goroutine 泄漏
package memory
