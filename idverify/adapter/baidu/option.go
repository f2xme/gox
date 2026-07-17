package baidu

import (
	"net/http"
	"time"
)

const (
	defaultTokenURL = "https://aip.baidubce.com/oauth/2.0/token"
	defaultMatchURL = "https://aip.baidubce.com/rest/2.0/face/v3/person/idmatch"
	defaultTimeout  = 10 * time.Second
	tokenMargin     = 5 * time.Minute
)

// Options 百度二要素配置。
type Options struct {
	// APIKey 百度 AI 应用 API Key。
	APIKey string
	// SecretKey 百度 AI 应用 Secret Key。
	SecretKey string
	// TokenURL OAuth token 地址，测试可覆盖。
	TokenURL string
	// MatchURL idmatch 地址，测试可覆盖。
	MatchURL string
	// HTTPClient 自定义 HTTP 客户端。
	HTTPClient *http.Client
	// Timeout 未设置 HTTPClient 时的默认超时。
	Timeout time.Duration
}

// Option 配置函数。
type Option func(*Options)

func defaultOptions() Options {
	return Options{
		TokenURL: defaultTokenURL,
		MatchURL: defaultMatchURL,
		Timeout:  defaultTimeout,
	}
}

// WithAPIKey 设置 API Key。
func WithAPIKey(key string) Option {
	return func(o *Options) { o.APIKey = key }
}

// WithSecretKey 设置 Secret Key。
func WithSecretKey(secret string) Option {
	return func(o *Options) { o.SecretKey = secret }
}

// WithTokenURL 覆盖 token URL。
func WithTokenURL(u string) Option {
	return func(o *Options) {
		if u != "" {
			o.TokenURL = u
		}
	}
}

// WithMatchURL 覆盖 idmatch URL。
func WithMatchURL(u string) Option {
	return func(o *Options) {
		if u != "" {
			o.MatchURL = u
		}
	}
}

// WithHTTPClient 设置 HTTP 客户端。
func WithHTTPClient(c *http.Client) Option {
	return func(o *Options) {
		if c != nil {
			o.HTTPClient = c
		}
	}
}

// WithTimeout 设置默认客户端超时。
func WithTimeout(d time.Duration) Option {
	return func(o *Options) {
		if d > 0 {
			o.Timeout = d
		}
	}
}

func (o *Options) validate() error {
	if o.APIKey == "" || o.SecretKey == "" {
		return idverifyNotConfigured()
	}
	if o.TokenURL == "" {
		o.TokenURL = defaultTokenURL
	}
	if o.MatchURL == "" {
		o.MatchURL = defaultMatchURL
	}
	if o.Timeout <= 0 {
		o.Timeout = defaultTimeout
	}
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: o.Timeout}
	}
	return nil
}
