package captcha

import (
	"time"

	"github.com/f2xme/gox/captcha/generator"
)

// Options 定义验证码配置选项。
type Options struct {
	// TTL 默认过期时间，默认 5 分钟
	TTL time.Duration
	// IDLength ID 长度，默认 20
	IDLength int
	// Generator 自定义生成器
	Generator generator.Generator
}

// Option 定义配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		TTL:      5 * time.Minute,
		IDLength: 20,
		// Generator 将在 New() 中设置默认值
	}
}

// WithTTL 设置默认过期时间。
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}

// WithIDLength 设置 ID 长度。
func WithIDLength(length int) Option {
	return func(o *Options) {
		o.IDLength = length
	}
}

// WithGenerator 设置自定义生成器。
func WithGenerator(gen generator.Generator) Option {
	return func(o *Options) {
		o.Generator = gen
	}
}
