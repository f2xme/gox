package mem

import (
	"errors"
	"time"
)

// Options 存储内存缓存的配置。
type Options struct {
	// MaxSize 是缓存中的最大条目数。
	// 值为 0 表示无限制。
	MaxSize int
	// CleanupInterval 设置清理 goroutine 运行的频率以删除过期条目。
	// 默认为 1 分钟。
	CleanupInterval time.Duration
	// EvictionPolicy 设置达到 MaxSize 时的淘汰策略。
	// 支持的值："lru"（最近最少使用）、"lfu"（最不经常使用）。
	// 默认为 "lru"。
	EvictionPolicy string
}

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		MaxSize:         0,               // 无限制
		CleanupInterval: 1 * time.Minute, // 每分钟清理一次
		EvictionPolicy:  "lru",           // 最近最少使用
	}
}

// Validate 验证配置。
func (o *Options) Validate() error {
	// 将负数 MaxSize 规范化为 0
	if o.MaxSize < 0 {
		o.MaxSize = 0
	}

	// 验证 CleanupInterval
	if o.CleanupInterval <= 0 {
		return errors.New("mem: cleanup interval must be positive")
	}

	// 验证 EvictionPolicy
	if o.EvictionPolicy != "lru" && o.EvictionPolicy != "lfu" {
		return errors.New("mem: eviction policy must be 'lru' or 'lfu'")
	}

	return nil
}

// Option 是修改 Options 的函数。
type Option func(*Options)

// WithMaxSize 设置缓存中的最大条目数。
// 值为 0 表示无限制。负值将被视为 0。
//
// 示例：
//
//	mem.New(mem.WithMaxSize(1000))
func WithMaxSize(size int) Option {
	return func(o *Options) {
		if size < 0 {
			size = 0
		}
		o.MaxSize = size
	}
}

// WithCleanupInterval 设置清理 goroutine 运行的频率
// 以删除过期条目。零或负值默认为 1 分钟。
//
// 示例：
//
//	mem.New(mem.WithCleanupInterval(5 * time.Minute))
func WithCleanupInterval(interval time.Duration) Option {
	return func(o *Options) {
		if interval <= 0 {
			interval = time.Minute
		}
		o.CleanupInterval = interval
	}
}

// WithEvictionPolicy 设置达到 maxSize 时的淘汰策略。
// 支持的值："lru"（最近最少使用）、"lfu"（最不经常使用）。
// 无效值默认为 "lru"。
//
// 示例：
//
//	mem.New(mem.WithEvictionPolicy("lfu"))
func WithEvictionPolicy(policy string) Option {
	return func(o *Options) {
		if policy != "lru" && policy != "lfu" {
			policy = "lru"
		}
		o.EvictionPolicy = policy
	}
}
