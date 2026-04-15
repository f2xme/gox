package timeout

import (
	"time"

	"github.com/f2xme/gox/httpx"
)

// Options 配置超时中间件的选项。
type Options struct {
	// Timeout 设置请求超时时长。
	// 默认值：30 秒。
	Timeout time.Duration

	// Handler 设置自定义超时处理函数。
	// 当请求超时时调用此函数。
	// 如果未设置，返回默认的 503 Service Unavailable 响应。
	Handler func(httpx.Context)
}

// Option 配置超时中间件。
type Option func(*Options)

// WithTimeout 设置请求超时时长。
// 默认值：30 秒。
func WithTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.Timeout = d
	}
}

// WithHandler 设置自定义超时处理函数。
// 当请求超时时调用此函数。
// 如果未设置，返回默认的 503 Service Unavailable 响应。
func WithHandler(fn func(httpx.Context)) Option {
	return func(o *Options) {
		o.Handler = fn
	}
}
