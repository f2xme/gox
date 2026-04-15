// Package viper 提供基于 viper 的配置适配器实现。
//
// 该包将 spf13/viper 封装为 gox/config 接口的实现，支持多种配置文件格式、
// 环境变量绑定和配置热重载功能。
//
// # 功能特性
//
//   - 支持多种配置文件格式（JSON、YAML、TOML、HCL 等）
//   - 自动环境变量绑定（支持前缀和键名转换）
//   - 配置文件热重载（基于 fsnotify）
//   - 默认值设置
//   - 类型安全的配置读取
//   - 线程安全的配置变更通知
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"log"
//		"github.com/f2xme/gox/config/adapter/viper"
//	)
//
//	func main() {
//		// 创建配置实例
//		cfg, err := viper.New("config.yaml")
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// 读取配置
//		port := cfg.GetInt("server.port")
//		host := cfg.GetString("server.host")
//	}
//
// # 环境变量绑定
//
// 启用环境变量自动绑定：
//
//	cfg, err := viper.New("config.yaml",
//		viper.WithEnvPrefix("APP"),
//	)
//
//	// 配置键 "server.port" 会自动绑定到环境变量 APP_SERVER_PORT
//	port := cfg.GetInt("server.port")
//
// # 配置热重载
//
// 监听配置文件变更：
//
//	cfg, err := viper.New("config.yaml",
//		viper.WithWatch(),
//	)
//
//	// 注册变更回调
//	cfg.(config.Watcher).Watch(func() {
//		log.Println("配置已更新")
//		// 重新读取配置
//	})
//
// # 设置默认值
//
// 为配置键设置默认值：
//
//	cfg, err := viper.New("config.yaml",
//		viper.WithDefaults(map[string]any{
//			"server.port": 8080,
//			"server.host": "localhost",
//			"debug":       false,
//		}),
//	)
//
// # 完整示例
//
// 结合多个选项使用：
//
//	cfg, err := viper.New("config.yaml",
//		viper.WithEnvPrefix("MYAPP"),
//		viper.WithDefaults(map[string]any{
//			"server.port":    8080,
//			"server.timeout": "30s",
//		}),
//		viper.WithWatch(),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 注册配置变更监听
//	if watcher, ok := cfg.(config.Watcher); ok {
//		watcher.Watch(func() {
//			log.Println("配置文件已更新，重新加载...")
//		})
//	}
//
//	// 读取配置
//	port := cfg.GetInt("server.port")
//	timeout := cfg.GetDuration("server.timeout")
//
// # 注意事项
//
//   - 配置文件路径必须包含文件扩展名（如 .yaml、.json）
//   - 环境变量键名会自动将 "." 替换为 "_"（如 server.port → SERVER_PORT）
//   - 配置热重载需要文件系统支持 fsnotify
//   - Watch 回调函数应避免阻塞操作，建议使用 goroutine 处理耗时任务
//   - Set 方法设置的是默认值，不会覆盖配置文件中已存在的值
package viper
