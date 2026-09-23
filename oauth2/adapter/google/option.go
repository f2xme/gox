package google

import (
	"net/http"
	"time"
)

const (
	defaultAuthURL     = "https://accounts.google.com/o/oauth2/v2/auth"
	defaultTokenURL    = "https://oauth2.googleapis.com/token"
	defaultUserURL     = "https://openidconnect.googleapis.com/v1/userinfo"
	defaultHTTPTimeout = 10 * time.Second
)

// Options 定义 Google 登录适配器配置。
type Options struct {
	// ClientID Google OAuth 客户端 ID。
	ClientID string
	// ClientSecret Google OAuth 客户端密钥。
	ClientSecret string
	// RedirectURL 授权回调地址，必须与 Google Cloud 控制台登记的地址一致。
	RedirectURL string
	// HTTPClient 用于请求 Google 接口的 HTTP 客户端。
	HTTPClient *http.Client
	// AuthURL 授权地址。
	AuthURL string
	// TokenURL 授权码换 token 地址。刷新地址为空时也使用该地址。
	TokenURL string
	// RefreshURL 刷新 token 地址，为空时复用 TokenURL。
	RefreshURL string
	// UserURL OpenID Connect 用户信息地址。
	UserURL string
}

// Option 定义 Google 登录适配器配置函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		HTTPClient: &http.Client{Timeout: defaultHTTPTimeout},
		AuthURL:    defaultAuthURL,
		TokenURL:   defaultTokenURL,
		UserURL:    defaultUserURL,
	}
}

// WithClientID 设置 Google OAuth 客户端 ID。
func WithClientID(clientID string) Option {
	return func(o *Options) {
		o.ClientID = clientID
	}
}

// WithClientSecret 设置 Google OAuth 客户端密钥。
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

// WithEndpoints 设置 Google 接口地址，主要用于测试或私有代理。
// refreshURL 为空时不覆盖，调用刷新接口时会复用 TokenURL。
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
