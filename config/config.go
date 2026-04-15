package config

import "time"

// Config 定义配置读取接口。
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

// Watcher 支持配置热重载。
// 实现此接口的类型会在配置文件变更时调用 fn。
type Watcher interface {
	Watch(fn func()) error
}
