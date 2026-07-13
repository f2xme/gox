package wechat

import (
	"net/http"
	"time"
)

const (
	defaultAuthURL     = "https://open.weixin.qq.com/connect/qrconnect"
	defaultTokenURL    = "https://api.weixin.qq.com/sns/oauth2/access_token"
	defaultRefreshURL  = "https://api.weixin.qq.com/sns/oauth2/refresh_token"
	defaultUserURL     = "https://api.weixin.qq.com/sns/userinfo"
	defaultHTTPTimeout = 10 * time.Second
)

// Options 定义微信登录适配器配置。
type Options struct {
	// ClientID 微信开放平台 AppID。
	ClientID string
	// ClientSecret 微信开放平台 AppSecret。
	ClientSecret string
	// RedirectURL 授权回调地址。
	RedirectURL string
	// HTTPClient 用于请求微信接口的 HTTP 客户端。
	HTTPClient *http.Client
	// AuthURL 授权地址。
	AuthURL string
	// TokenURL 授权码换 token 地址。
	TokenURL string
	// RefreshURL 刷新 token 地址。
	RefreshURL string
	// UserURL 用户信息地址。
	UserURL string
}

// Option 定义微信登录适配器配置函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		HTTPClient: &http.Client{Timeout: defaultHTTPTimeout},
		AuthURL:    defaultAuthURL,
		TokenURL:   defaultTokenURL,
		RefreshURL: defaultRefreshURL,
		UserURL:    defaultUserURL,
	}
}

// WithClientID 设置微信开放平台 AppID。
func WithClientID(clientID string) Option {
	return func(o *Options) {
		o.ClientID = clientID
	}
}

// WithClientSecret 设置微信开放平台 AppSecret。
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

// WithEndpoints 设置微信接口地址，主要用于测试或私有代理。
func WithEndpoints(authURL, tokenURL, refreshURL, userURL string) Option {
	return func(o *Options) {
		if authURL != "" {
			o.AuthURL = authURL
		}
		if tokenURL != "" {
			o.TokenURL = tokenURL
		}
		if refreshURL != "" {
			o.RefreshURL = refreshURL
		}
		if userURL != "" {
			o.UserURL = userURL
		}
	}
}
