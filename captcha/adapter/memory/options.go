package memory

import "time"

// Options 定义内存适配器的配置选项。
type Options struct {
	// TTL 默认过期时间，默认 5 分钟
	TTL time.Duration
	// CleanupInterval 清理间隔，默认 1 分钟
	CleanupInterval time.Duration
	// MaxSize 最大条目数，0 表示无限制
	MaxSize int
}

// Option 定义配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxSize:         0,
	}
}

// WithTTL 设置默认过期时间。
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}

// WithCleanupInterval 设置清理间隔。
func WithCleanupInterval(interval time.Duration) Option {
	return func(o *Options) {
		o.CleanupInterval = interval
	}
}

// WithMaxSize 设置最大条目数。
func WithMaxSize(size int) Option {
	return func(o *Options) {
		o.MaxSize = size
	}
}
