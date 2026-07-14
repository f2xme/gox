package http

import (
	"net/http"
	"time"

	"github.com/f2xme/gox/geo"
)

// ResponseParser 将 HTTP 响应体解析为地区信息。
//
// body 是原始响应内容，statusCode 是 HTTP 状态码，ip 是规范化后的查询 IP。
// 返回的 Location 可以不填充 IP，适配器会自动补全。
type ResponseParser func(body []byte, statusCode int, ip string) (*geo.Location, error)

// Options 定义 HTTP IP 地区适配器配置选项。
type Options struct {
	// Endpoint 查询端点。
	//
	// 支持两种形式：
	//  1. 包含 "%s" 的 URL 模板，例如 "https://example.com/ip/%s"
	//  2. 前缀 URL，会自动拼接 IP，例如 "http://ip-api.com/json/"
	Endpoint string
	// Timeout 单次请求超时，默认 5 秒；<=0 时回落为默认 5 秒。
	// 自定义 Client 的非零 Timeout 优先；Client.Timeout==0 时会用本字段写入 clone，
	// 不修改调用方传入的原 Client。需要无限等待时请使用 Timeout>0 的自定义 Client。
	Timeout time.Duration
	// Client 自定义 HTTP 客户端。为空时使用内置客户端。
	Client *http.Client
	// Headers 额外请求头。
	Headers map[string]string
	// Parser 自定义响应解析函数。为空时使用默认 JSON 解析器。
	Parser ResponseParser
}

// Option 定义配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		Timeout: 5 * time.Second,
		Headers: make(map[string]string),
	}
}

// WithEndpoint 设置查询端点。
//
// 示例：
//
//	New(WithEndpoint("http://ip-api.com/json/"))
//	New(WithEndpoint("https://example.com/geo?ip=%s"))
func WithEndpoint(endpoint string) Option {
	return func(o *Options) {
		o.Endpoint = endpoint
	}
}

// WithTimeout 设置请求超时时间。
//
// 示例：
//
//	New(WithEndpoint("http://ip-api.com/json/"), WithTimeout(3*time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithHTTPClient 设置自定义 HTTP 客户端。
//
// 示例：
//
//	New(WithEndpoint("http://ip-api.com/json/"), WithHTTPClient(customClient))
func WithHTTPClient(client *http.Client) Option {
	return func(o *Options) {
		o.Client = client
	}
}

// WithHeader 设置单个请求头。
//
// 示例：
//
//	New(WithEndpoint("https://api.example.com/"), WithHeader("Authorization", "Bearer xxx"))
func WithHeader(key, value string) Option {
	return func(o *Options) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		o.Headers[key] = value
	}
}

// WithHeaders 批量设置请求头。
//
// 示例：
//
//	New(WithEndpoint("https://api.example.com/"), WithHeaders(map[string]string{
//		"Authorization": "Bearer xxx",
//	}))
func WithHeaders(headers map[string]string) Option {
	return func(o *Options) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		for k, v := range headers {
			o.Headers[k] = v
		}
	}
}

// WithResponseParser 设置自定义响应解析函数。
//
// 示例：
//
//	New(WithEndpoint("https://api.example.com/"), WithResponseParser(myParser))
func WithResponseParser(parser ResponseParser) Option {
	return func(o *Options) {
		o.Parser = parser
	}
}
