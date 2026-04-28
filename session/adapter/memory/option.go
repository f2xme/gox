package memory

import "time"

// Options 定义内存会话存储配置。
type Options struct {
	// CleanupInterval 是过期会话清理间隔。
	// 默认 1 分钟。
	CleanupInterval time.Duration
}

// Option 定义内存会话存储配置函数。
type Option func(*Options)

func defaultOptions() Options {
	return Options{
		CleanupInterval: time.Minute,
	}
}

func (o Options) validate() error {
	if o.CleanupInterval <= 0 {
		return nil
	}
	return nil
}

// WithCleanupInterval 设置过期会话清理间隔。
//
// 示例：
//
//	memory.New(memory.WithCleanupInterval(5*time.Minute))
func WithCleanupInterval(interval time.Duration) Option {
	return func(o *Options) {
		if interval <= 0 {
			interval = time.Minute
		}
		o.CleanupInterval = interval
	}
}
