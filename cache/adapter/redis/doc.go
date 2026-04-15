// Package redis 提供基于 Redis 的分布式缓存实现。
//
// # 功能特性
//
//   - 分布式缓存（跨进程共享）
//   - 批量操作（MultiCache 接口）
//   - 分布式锁（基于 Redis）
//   - 持久化支持
//   - 线程安全
//   - 支持 Redis Cluster 和 Sentinel
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
//		"github.com/f2xme/gox/cache/adapter/redis"
//	)
//
//	func main() {
//		// 创建 Redis 缓存
//		c, err := redis.New(
//			redis.WithAddr("localhost:6379"),
//		)
//		if err != nil {
//			panic(err)
//		}
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
// 基本连接配置：
//
//	c, _ := redis.New(
//		redis.WithAddr("localhost:6379"),
//		redis.WithPassword("secret"),
//		redis.WithDB(1),
//	)
//
// 使用自定义 Redis 客户端：
//
//	import goredis "github.com/redis/go-redis/v9"
//
//	// 单机模式
//	client := goredis.NewClient(&goredis.Options{
//		Addr:     "localhost:6379",
//		Password: "secret",
//		DB:       0,
//	})
//
//	// Cluster 模式
//	client := goredis.NewClusterClient(&goredis.ClusterOptions{
//		Addrs: []string{"localhost:7000", "localhost:7001"},
//	})
//
//	// Sentinel 模式
//	client := goredis.NewFailoverClient(&goredis.FailoverOptions{
//		MasterName:    "mymaster",
//		SentinelAddrs: []string{"localhost:26379"},
//	})
//
//	c, _ := redis.New(redis.WithClient(client))
//
// # 使用配置文件
//
// 从 config.Config 创建：
//
//	// 配置文件（YAML）
//	// cache:
//	//   redis:
//	//     addr: localhost:6379
//	//     password: secret
//	//     db: 1
//
//	c, err := redis.NewWithConfig(cfg)
//
// # 批量操作
//
// redis 包实现了 MultiCache 接口，支持批量操作以提高性能：
//
//	mc := c.(cache.MultiCache)
//
//	// 批量设置
//	items := map[string][]byte{
//		"key1": []byte("value1"),
//		"key2": []byte("value2"),
//		"key3": []byte("value3"),
//	}
//	mc.SetMulti(ctx, items, 5*time.Minute)
//
//	// 批量获取
//	keys := []string{"key1", "key2", "key3"}
//	results, err := mc.GetMulti(ctx, keys)
//
//	// 批量删除
//	mc.DeleteMulti(ctx, keys)
//
// # 分布式锁
//
// redis 包实现了基于 Redis 的分布式锁：
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
//	if err == cache.ErrLockFailed {
//		// 锁已被占用
//	}
//
// # 注意事项
//
//   - Redis 缓存适用于分布式应用，支持跨进程共享
//   - 需要确保 Redis 服务可用，否则操作会失败
//   - 批量操作可以显著减少网络往返次数
//   - 分布式锁的 TTL 应大于操作时间，避免锁提前过期
//   - 使用 WithClient 时，客户端的生命周期由调用方管理
package redis
