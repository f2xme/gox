package traceid

import (
	"crypto/rand"
	"encoding/hex"
)

const (
	defaultHeaderKey = "X-Trace-ID"
	contextKey       = "trace_id"
)

// Options 配置 Trace ID 中间件。
type Options struct {
	headerKey string
	generator func() string
}

// Option 配置 Trace ID 中间件。
type Option func(*Options)

func defaultOptions() *Options {
	return &Options{
		headerKey: defaultHeaderKey,
		generator: defaultGenerator,
	}
}

func defaultGenerator() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// WithHeaderKey 设置 Trace ID 请求头名称。
func WithHeaderKey(key string) Option {
	return func(o *Options) {
		o.headerKey = key
	}
}

// WithGenerator 设置 Trace ID 生成函数。
func WithGenerator(fn func() string) Option {
	return func(o *Options) {
		o.generator = fn
	}
}
