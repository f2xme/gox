// Package memory 提供内存存储适配器。
//
// 内存适配器适合单机场景，提供轻量级的验证码存储方案。
// 支持自动过期清理和容量限制。
//
// 示例：
//
//	store := memory.New(
//		memory.WithTTL(10*time.Minute),
//		memory.WithCleanupInterval(2*time.Minute),
//		memory.WithMaxSize(10000),
//	)
package memory
