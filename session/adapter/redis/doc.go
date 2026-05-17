// Package redis 提供基于 Redis 的 session.Store 实现。
//
// # 功能特性
//
//   - 使用 Redis 保存 session.Session 数据
//   - 根据会话过期时间设置 Redis key TTL
//   - 支持自定义 go-redis 客户端和 key 前缀
//   - 适合多实例服务共享会话状态
//
// # 快速开始
//
//	store, err := redis.New(
//		redis.WithAddr("127.0.0.1:6379"),
//		redis.WithKeyPrefix("app:session:"),
//	)
//	_ = store
//	_ = err
package redis
