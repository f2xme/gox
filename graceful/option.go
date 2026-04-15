package graceful

import (
	"os"
	"time"
)

// Option 配置 Manager
type Option func(*manager)

// RegisterOption 配置资源注册
type RegisterOption func(*registerConfig)

// registerConfig 资源注册配置
type registerConfig struct {
	priority int
	timeout  time.Duration
}

// WithPriority 设置资源的优先级
//
// 优先级高的资源先关闭。默认优先级为 0。
//
// 示例：
//
//	graceful.WithPriority(100)
func WithPriority(priority int) RegisterOption {
	return func(cfg *registerConfig) {
		cfg.priority = priority
	}
}

// WithResourceTimeout 设置关闭特定资源的超时时间
//
// 如果不设置，使用 Manager 的默认超时时间。
//
// 示例：
//
//	graceful.WithResourceTimeout(5 * time.Second)
func WithResourceTimeout(timeout time.Duration) RegisterOption {
	return func(cfg *registerConfig) {
		cfg.timeout = timeout
	}
}

// WithTimeout 设置所有资源的默认超时时间
//
// 默认为 30 秒。
//
// 示例：
//
//	graceful.New(graceful.WithTimeout(10 * time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(m *manager) {
		m.timeout = timeout
	}
}

// WithSignals 设置要监听的信号
//
// 默认监听 SIGTERM 和 SIGINT。
//
// 示例：
//
//	graceful.New(graceful.WithSignals(syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT))
func WithSignals(signals ...os.Signal) Option {
	return func(m *manager) {
		m.signals = signals
	}
}

// WithLogger 设置自定义日志记录器
//
// 示例：
//
//	graceful.New(graceful.WithLogger(myLogger))
func WithLogger(logger Logger) Option {
	return func(m *manager) {
		m.logger = logger
	}
}

// OnBeforeShutdown 注册在关闭开始前调用的钩子
//
// 示例：
//
//	graceful.New(graceful.OnBeforeShutdown(func() {
//		log.Println("开始关闭...")
//	}))
func OnBeforeShutdown(fn func()) Option {
	return func(m *manager) {
		m.hooks.beforeShutdown = fn
	}
}

// OnAfterShutdown 注册在关闭完成后调用的钩子
//
// 示例：
//
//	graceful.New(graceful.OnAfterShutdown(func() {
//		log.Println("关闭完成")
//	}))
func OnAfterShutdown(fn func()) Option {
	return func(m *manager) {
		m.hooks.afterShutdown = fn
	}
}

// OnTimeout 注册在资源超时时调用的钩子
//
// 参数为超时的资源名称。
//
// 示例：
//
//	graceful.New(graceful.OnTimeout(func(name string) {
//		log.Printf("资源 %s 关闭超时", name)
//	}))
func OnTimeout(fn func(string)) Option {
	return func(m *manager) {
		m.hooks.onTimeout = fn
	}
}
