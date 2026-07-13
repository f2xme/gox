package httpadapter

import (
	"net/http"
	"time"
)

const (
	defaultTimeout      = 5 * time.Second
	defaultMaxBodyBytes = int64(1 << 20)
)

// Format 定义远端配置内容格式。
type Format string

const (
	// Auto 根据 URL 扩展名或 Content-Type 自动判断配置格式。
	Auto Format = ""
	// JSON 表示 JSON 配置内容。
	JSON Format = "json"
	// YAML 表示 YAML 配置内容。
	YAML Format = "yaml"
)

// Options 定义 HTTP 配置适配器的配置选项。
type Options struct {
	client          *http.Client
	headers         map[string]string
	defaults        map[string]any
	format          Format
	timeout         time.Duration
	maxBodyBytes    int64
	watch           bool
	watchInterval   time.Duration
	failOnLoadError bool
}

// Option 定义配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置选项。
func defaultOptions() Options {
	return Options{
		headers:         make(map[string]string),
		timeout:         defaultTimeout,
		maxBodyBytes:    defaultMaxBodyBytes,
		failOnLoadError: true,
	}
}

// WithClient 设置自定义 HTTP 客户端。
func WithClient(client *http.Client) Option {
	return func(o *Options) {
		o.client = client
	}
}

// WithHeader 设置拉取远端配置时附加的请求头。
func WithHeader(key, value string) Option {
	return func(o *Options) {
		if o.headers == nil {
			o.headers = make(map[string]string)
		}
		o.headers[key] = value
	}
}

// WithHeaders 批量设置拉取远端配置时附加的请求头。
func WithHeaders(headers map[string]string) Option {
	return func(o *Options) {
		if o.headers == nil {
			o.headers = make(map[string]string)
		}
		for k, v := range headers {
			o.headers[k] = v
		}
	}
}

// WithDefaults 设置配置键的默认值。
//
// 默认值仅在远端配置中不存在对应键时生效。
func WithDefaults(defaults map[string]any) Option {
	return func(o *Options) {
		o.defaults = defaults
	}
}

// WithFormat 设置远端配置内容格式。
func WithFormat(format Format) Option {
	return func(o *Options) {
		o.format = format
	}
}

// WithTimeout 设置拉取远端配置的请求超时时间。
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		if timeout > 0 {
			o.timeout = timeout
		}
	}
}

// WithMaxBodyBytes 设置远端配置响应体最大字节数。
func WithMaxBodyBytes(size int64) Option {
	return func(o *Options) {
		if size > 0 {
			o.maxBodyBytes = size
		}
	}
}

// WithWatch 启用远端配置轮询监听。
//
// interval <= 0 时使用默认请求超时时间作为轮询间隔。
func WithWatch(interval time.Duration) Option {
	return func(o *Options) {
		o.watch = true
		o.watchInterval = interval
	}
}

// WithFailOnLoadError 设置初始化拉取失败时是否返回错误。
//
// 默认值为 true。设置为 false 后，如果远端配置加载失败，New 会使用默认值
// 创建配置实例。
func WithFailOnLoadError(fail bool) Option {
	return func(o *Options) {
		o.failOnLoadError = fail
	}
}
