// Package memory 提供基于内存的 session.Store 实现。
//
// # 功能特性
//
//   - 在进程内保存 session.Session 数据
//   - 按会话过期时间自动清理过期数据
//   - 支持配置清理间隔
//   - 适合测试、单实例应用和本地开发
//
// # 快速开始
//
//	store, err := memory.New(
//		memory.WithCleanupInterval(time.Minute),
//	)
//	_ = store
//	_ = err
package memory
