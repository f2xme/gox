package baidu

import (
	"net/http"
	"time"
)

const defaultEndpoint = "https://sp0.baidu.com/8aQDcjqpAAV3otqbppnN2DJv/api.php"

// Options 定义百度 IP 查询适配器配置选项。
type Options struct {
	// Endpoint API 地址，默认百度 IP 查询接口。
	Endpoint string
	// Timeout 单次请求超时，默认 5 秒；<=0 时回落为默认 5 秒。
	// 自定义 Client 的非零 Timeout 优先；Client.Timeout==0 时用本字段写入 clone。
	Timeout time.Duration
	// Client 自定义 HTTP 客户端。为空时使用内置客户端。
	Client *http.Client
}

// Option 定义配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		Endpoint: defaultEndpoint,
		Timeout:  5 * time.Second,
	}
}

// WithEndpoint 设置 API 端点，主要用于测试。
//
// 示例：
//
//	New(WithEndpoint(server.URL))
func WithEndpoint(endpoint string) Option {
	return func(o *Options) {
		o.Endpoint = endpoint
	}
}

// WithTimeout 设置请求超时时间。
//
// 示例：
//
//	New(WithTimeout(3*time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithHTTPClient 设置自定义 HTTP 客户端。
//
// 示例：
//
//	New(WithHTTPClient(customClient))
func WithHTTPClient(client *http.Client) Option {
	return func(o *Options) {
		o.Client = client
	}
}
