// Package idgen 提供多种 ID 生成算法
//
// 支持多种 ID 生成策略：
//   - Snowflake: 基于 Twitter Snowflake 算法的分布式唯一 ID
//   - UUID/ULID: 通用唯一标识符
//   - ShortID: 短小的 URL 安全 ID
//   - AutoIncrement: 线程安全的自增 ID
//   - GeneratorFunc: 自定义生成逻辑的适配器
//
// 所有函数都是并发安全的
package idgen

// Generator 定义 ID 生成器接口
type Generator interface {
	Generate() string
}

// GeneratorFunc 是 Generator 接口的函数适配器
type GeneratorFunc func() string

// Generate 实现 Generator 接口
func (f GeneratorFunc) Generate() string {
	return f()
}
