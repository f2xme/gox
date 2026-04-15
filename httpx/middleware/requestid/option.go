package requestid

import (
	"crypto/rand"
	"fmt"

	"github.com/f2xme/gox/httpx"
)

// Options 配置请求 ID 中间件
type Options struct {
	headerKey string
	generator func() string
	handler   func(httpx.Context, string)
}

// Option 配置请求 ID 中间件
type Option func(*Options)

const defaultHeaderKey = "X-Request-ID"

func defaultOptions() *Options {
	return &Options{
		headerKey: defaultHeaderKey,
		generator: defaultGenerator,
		handler:   nil,
	}
}

func defaultGenerator() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// WithHeaderKey 设置请求 ID 的 HTTP 头键名
func WithHeaderKey(key string) Option {
	return func(o *Options) {
		o.headerKey = key
	}
}

// WithGenerator 设置自定义请求 ID 生成函数
func WithGenerator(fn func() string) Option {
	return func(o *Options) {
		o.generator = fn
	}
}

// WithHandler 设置可选的回调函数，在生成请求 ID 后调用
func WithHandler(fn func(httpx.Context, string)) Option {
	return func(o *Options) {
		o.handler = fn
	}
}
