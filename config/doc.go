/*
Package config 提供统一的配置读取抽象层。

# 概述

config 包定义了配置读取的标准接口，支持多种配置源（文件、环境变量、远程配置中心等）。
通过这些接口，你可以轻松地在不同的配置实现之间切换，而无需修改业务代码。

# 核心接口

## Config - 配置读取接口

所有配置实现都必须实现此接口：

	type Config interface {
		Get(key string) any
		GetString(key string) string
		GetStringSlice(key string) []string
		GetStringMap(key string) map[string]any
		GetInt(key string) int
		GetInt64(key string) int64
		GetDuration(key string) time.Duration
		GetBool(key string) bool
	}

使用示例：

	// 读取字符串配置
	dbHost := cfg.GetString("database.host")

	// 读取整数配置
	port := cfg.GetInt("server.port")

	// 读取时间间隔
	timeout := cfg.GetDuration("server.timeout")

	// 读取布尔值
	debug := cfg.GetBool("app.debug")

## Watcher - 配置热重载接口

支持配置文件变更时自动重新加载：

	type Watcher interface {
		Watch(fn func()) error
	}

使用示例：

	if watcher, ok := cfg.(config.Watcher); ok {
		watcher.Watch(func() {
			log.Println("配置已更新")
			// 重新加载配置
		})
	}

# 配置键命名规范

推荐使用点分隔的层级结构：

	database.host
	database.port
	server.http.port
	server.grpc.port
	cache.redis.addr

# 最佳实践

## 1. 使用类型安全的 Get 方法

	// 推荐：使用类型安全的方法
	port := cfg.GetInt("server.port")

	// 不推荐：使用 Get 后手动类型断言
	port := cfg.Get("server.port").(int)

## 2. 提供默认值

	port := cfg.GetInt("server.port")
	if port == 0 {
		port = 8080 // 默认值
	}

## 3. 使用配置热重载

	if watcher, ok := cfg.(config.Watcher); ok {
		watcher.Watch(func() {
			// 重新初始化依赖配置的组件
			reloadDatabase()
			reloadCache()
		})
	}

# 线程安全

所有配置实现都应该是线程安全的，可以在多个 goroutine 中并发读取。
*/
package config
