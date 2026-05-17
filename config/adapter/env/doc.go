// Package env 提供基于环境变量的配置适配器实现。
//
// 该包将当前进程的环境变量封装为 gox/config 接口的实现，适用于容器、
// Serverless 和十二要素应用等以环境变量作为主要配置来源的场景。
//
// # 功能特性
//
//   - 零第三方依赖
//   - 支持环境变量前缀
//   - 支持默认值
//   - 支持环境变量变更监听
//   - 自动将配置键转换为环境变量名
//   - 类型安全的配置读取
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import "github.com/f2xme/gox/config/adapter/env"
//
//	func main() {
//		cfg := env.New(
//			env.WithPrefix("APP"),
//			env.WithDefaults(map[string]any{
//				"server.port": 8080,
//			}),
//		)
//
//		// 配置键 "server.port" 会读取环境变量 APP_SERVER_PORT
//		port := cfg.GetInt("server.port")
//		_ = port
//	}
//
// # 键名转换
//
// 默认会将配置键转为大写，并把 "." 和 "-" 替换为 "_"：
//
//   - server.port -> SERVER_PORT
//   - database.max-connections -> DATABASE_MAX_CONNECTIONS
//   - app.debug -> APP_DEBUG
//
// 设置前缀后，前缀会追加在转换后的键名前：
//
//	cfg := env.New(env.WithPrefix("MYAPP"))
//	host := cfg.GetString("database.host") // 读取 MYAPP_DATABASE_HOST
//	_ = host
//
// # 变更监听
//
// 启用环境变量变更监听：
//
//	cfg := env.New(env.WithPrefix("APP"), env.WithWatch())
//	cfg.(config.Watcher).Watch(func() {
//		// 当前进程环境变量发生变化
//	})
//
// # 注意事项
//
//   - 该适配器只读取当前进程环境变量，不加载 .env 文件
//   - 环境变量值始终以字符串读取，再按 Get 方法转换为目标类型
//   - GetStringSlice 默认按英文逗号分隔环境变量值
//   - Watch 通过轮询当前进程环境变量实现，无法感知进程外环境变量变化
package env
