package qq

import (
	"net/http"
	"time"
)

const (
	defaultAuthURL     = "https://graph.qq.com/oauth2.0/authorize"
	defaultTokenURL    = "https://graph.qq.com/oauth2.0/token"
	defaultOpenIDURL   = "https://graph.qq.com/oauth2.0/me"
	defaultUserURL     = "https://graph.qq.com/user/get_user_info"
	defaultHTTPTimeout = 10 * time.Second
)

// Options 定义 QQ 登录适配器配置。
type Options struct {
	// ClientID QQ 互联 App ID。
	ClientID string
	// ClientSecret QQ 互联 App Key。
	ClientSecret string
	// RedirectURL 授权回调地址。
	RedirectURL string
	// HTTPClient 用于请求 QQ 接口的 HTTP 客户端。
	HTTPClient *http.Client
	// AuthURL 授权地址。
	AuthURL string
	// TokenURL 授权码换 token 地址。
	TokenURL string
	// OpenIDURL 获取 openid 地址。
	OpenIDURL string
	// UserURL 用户信息地址。
	UserURL string
}

// Option 定义 QQ 登录适配器配置函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		HTTPClient: &http.Client{Timeout: defaultHTTPTimeout},
		AuthURL:    defaultAuthURL,
		TokenURL:   defaultTokenURL,
		OpenIDURL:  defaultOpenIDURL,
		UserURL:    defaultUserURL,
	}
}

// WithClientID 设置 QQ 互联 App ID。
func WithClientID(clientID string) Option {
	return func(o *Options) {
		o.ClientID = clientID
	}
}

// WithClientSecret 设置 QQ 互联 App Key。
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

// WithHTTPClient 设置 HTTP 客户端。
func WithHTTPClient(client *http.Client) Option {
	return func(o *Options) {
		if client != nil {
			o.HTTPClient = client
		}
	}
}

// WithEndpoints 设置 QQ 接口地址，主要用于测试或私有代理。
func WithEndpoints(authURL, tokenURL, openIDURL, userURL string) Option {
	return func(o *Options) {
		if authURL != "" {
			o.AuthURL = authURL
		}
		if tokenURL != "" {
			o.TokenURL = tokenURL
		}
		if openIDURL != "" {
			o.OpenIDURL = openIDURL
		}
		if userURL != "" {
			o.UserURL = userURL
		}
	}
}
