package oauth2

import (
	"net/http"
	"time"
)

const defaultHTTPTimeout = 10 * time.Second

// Options 定义标准 OAuth2 客户端配置。
type Options struct {
	// ClientID 客户端 ID。
	ClientID string
	// ClientSecret 客户端密钥。
	ClientSecret string
	// RedirectURL 授权回调地址。
	RedirectURL string
	// Endpoint OAuth2 端点地址。
	Endpoint Endpoint
	// HTTPClient 用于请求 OAuth2 接口的 HTTP 客户端。
	HTTPClient *http.Client
}

// Option 定义标准 OAuth2 客户端配置函数。
type Option func(*Options)

// WithClientID 设置客户端 ID。
func WithClientID(clientID string) Option {
	return func(o *Options) {
		o.ClientID = clientID
	}
}

// WithClientSecret 设置客户端密钥。
func WithClientSecret(clientSecret string) Option {
	return func(o *Options) {
		o.ClientSecret = clientSecret
	}
}

// WithRedirectURL 设置授权回调地址。
func WithRedirectURL(redirectURL string) Option {
	return func(o *Options) {
		o.RedirectURL = redirectURL
	}
}

// WithEndpoint 设置 OAuth2 端点地址。
func WithEndpoint(endpoint Endpoint) Option {
	return func(o *Options) {
		o.Endpoint = endpoint
	}
}

// WithHTTPClient 设置 HTTP 客户端。
func WithHTTPClient(client *http.Client) Option {
	return func(o *Options) {
		if client != nil {
			o.HTTPClient = client
		}
	}
}

func defaultOptions() Options {
	return Options{HTTPClient: &http.Client{Timeout: defaultHTTPTimeout}}
}
