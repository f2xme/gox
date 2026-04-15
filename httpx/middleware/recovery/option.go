package recovery

import "github.com/f2xme/gox/httpx"

// Options 配置 recovery 中间件的选项。
type Options struct {
	// Handler 自定义 panic 处理回调函数。
	// 接收上下文和恢复的错误。
	Handler func(httpx.Context, error)
}

// Option 是配置 Options 的函数类型。
type Option func(*Options)

// WithHandler 设置自定义 panic 处理回调函数。
// 回调函数接收上下文和恢复的错误。
func WithHandler(fn func(httpx.Context, error)) Option {
	return func(o *Options) {
		o.Handler = fn
	}
}

// defaultOptions 返回默认配置。
func defaultOptions() *Options {
	return &Options{
		Handler: nil,
	}
}
