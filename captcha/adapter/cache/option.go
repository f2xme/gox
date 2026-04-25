package cache

import "time"

// Options 定义 cache 适配器的配置选项。
type Options struct {
	// TTL 默认过期时间，默认 5 分钟
	TTL time.Duration
	// Prefix key 前缀，默认 "captcha:"
	Prefix string
}

// Option 定义配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		TTL:    5 * time.Minute,
		Prefix: "captcha:",
	}
}

// WithTTL 设置默认过期时间。
//
// 示例：
//
//	cacheadapter.New(c, cacheadapter.WithTTL(10*time.Minute))
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}

// WithPrefix 设置 key 前缀。
//
// 示例：
//
//	cacheadapter.New(c, cacheadapter.WithPrefix("captcha:login:"))
func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
	}
}
