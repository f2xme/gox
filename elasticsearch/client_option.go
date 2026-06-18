package elasticsearch

import (
	"net/http"
	"time"
)

const (
	defaultMaxRetries            = 3
	defaultMaxIdleConnsPerHost   = 100
	defaultResponseHeaderTimeout = 10 * time.Second
	defaultDialTimeout           = 5 * time.Second
	defaultIdleConnTimeout       = 90 * time.Second
)

// Options 定义 Elasticsearch 客户端配置选项。
type Options struct {
	// Addresses Elasticsearch 节点地址列表。
	Addresses []string
	// APIKey API Key 认证凭证。
	APIKey string
	// Username Basic Auth 用户名。
	Username string
	// Password Basic Auth 密码。
	Password string
	// CloudID Elastic Cloud ID。
	CloudID string
	// ServiceToken Service Token 认证凭证。
	ServiceToken string
	// MaxRetries 最大重试次数。
	MaxRetries int
	// MaxIdleConnsPerHost 每个 Host 最大空闲连接数。
	MaxIdleConnsPerHost int
	// ResponseHeaderTimeout 响应头超时时间。
	ResponseHeaderTimeout time.Duration
	// DialTimeout 建连超时时间。
	DialTimeout time.Duration
	// IdleConnTimeout 空闲连接超时时间。
	IdleConnTimeout time.Duration
	// Transport 自定义 HTTP RoundTripper。
	Transport http.RoundTripper
	// SkipPing 是否跳过创建客户端后的连通性检查。
	SkipPing bool
}

func defaultOptions() Options {
	return Options{
		MaxRetries:            defaultMaxRetries,
		MaxIdleConnsPerHost:   defaultMaxIdleConnsPerHost,
		ResponseHeaderTimeout: defaultResponseHeaderTimeout,
		DialTimeout:           defaultDialTimeout,
		IdleConnTimeout:       defaultIdleConnTimeout,
	}
}

// Option 定义 Elasticsearch 客户端配置选项函数。
type Option func(*Options)

// WithAddresses 设置 Elasticsearch 节点地址列表。
//
// 示例：
//
//	elastic.New(elastic.WithAddresses("http://localhost:9200"))
func WithAddresses(addresses ...string) Option {
	return func(o *Options) {
		o.Addresses = addresses
	}
}

// WithAPIKey 设置 API Key 认证凭证。
func WithAPIKey(apiKey string) Option {
	return func(o *Options) {
		o.APIKey = apiKey
	}
}

// WithBasicAuth 设置 Basic Auth 用户名和密码。
func WithBasicAuth(username, password string) Option {
	return func(o *Options) {
		o.Username = username
		o.Password = password
	}
}

// WithCloudID 设置 Elastic Cloud ID。
func WithCloudID(cloudID string) Option {
	return func(o *Options) {
		o.CloudID = cloudID
	}
}

// WithServiceToken 设置 Service Token 认证凭证。
func WithServiceToken(serviceToken string) Option {
	return func(o *Options) {
		o.ServiceToken = serviceToken
	}
}

// WithMaxRetries 设置最大重试次数。
func WithMaxRetries(maxRetries int) Option {
	return func(o *Options) {
		if maxRetries < 0 {
			maxRetries = 0
		}
		o.MaxRetries = maxRetries
	}
}

// WithMaxIdleConnsPerHost 设置每个 Host 最大空闲连接数。
func WithMaxIdleConnsPerHost(n int) Option {
	return func(o *Options) {
		if n > 0 {
			o.MaxIdleConnsPerHost = n
		}
	}
}

// WithResponseHeaderTimeout 设置响应头超时时间。
func WithResponseHeaderTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		if timeout > 0 {
			o.ResponseHeaderTimeout = timeout
		}
	}
}

// WithDialTimeout 设置建连超时时间。
func WithDialTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		if timeout > 0 {
			o.DialTimeout = timeout
		}
	}
}

// WithIdleConnTimeout 设置空闲连接超时时间。
func WithIdleConnTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		if timeout > 0 {
			o.IdleConnTimeout = timeout
		}
	}
}

// WithTransport 设置自定义 HTTP RoundTripper。
func WithTransport(transport http.RoundTripper) Option {
	return func(o *Options) {
		o.Transport = transport
	}
}

// WithSkipPing 设置是否跳过创建客户端后的连通性检查。
func WithSkipPing(skip bool) Option {
	return func(o *Options) {
		o.SkipPing = skip
	}
}

// WithOptions 组合多个配置选项。
func WithOptions(opts ...Option) Option {
	return func(o *Options) {
		for _, opt := range opts {
			opt(o)
		}
	}
}
