package facebook

import (
	"net/http"
	"time"
)

const (
	graphAPIVersion    = "v26.0"
	defaultAuthURL     = "https://www.facebook.com/" + graphAPIVersion + "/dialog/oauth"
	defaultTokenURL    = "https://graph.facebook.com/" + graphAPIVersion + "/oauth/access_token"
	defaultUserURL     = "https://graph.facebook.com/" + graphAPIVersion + "/me"
	defaultHTTPTimeout = 10 * time.Second
)

// Options 定义 Facebook 登录适配器配置。
type Options struct {
	// ClientID Facebook 应用编号。
	ClientID string
	// ClientSecret Facebook 应用密钥。
	ClientSecret string
	// RedirectURL 授权回调地址，必须与应用后台登记的地址一致。
	RedirectURL string
	// HTTPClient 用于请求 Facebook 接口的 HTTP 客户端。
	HTTPClient *http.Client
	// AuthURL 授权地址。
	AuthURL string
	// TokenURL 授权码换 token 地址。刷新地址为空时也使用该地址。
	TokenURL string
	// RefreshURL 短期令牌换长期令牌的地址，为空时复用 TokenURL。
	RefreshURL string
	// UserURL 用户信息地址。
	UserURL string
}

// Option 定义 Facebook 登录适配器配置函数。
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

// WithClientID 设置 Facebook 应用编号。
func WithClientID(clientID string) Option {
	return func(o *Options) {
		o.ClientID = clientID
	}
}

// WithClientSecret 设置 Facebook 应用密钥。
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

// WithEndpoints 设置 Facebook 接口地址，主要用于测试或私有代理。
// refreshURL 为空时不覆盖，换长期令牌时会复用 TokenURL。
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
