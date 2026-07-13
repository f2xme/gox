package oauth2

import (
	"net/url"
	"strings"
)

// AuthCodeURLConfig 定义授权码登录地址构造配置。
type AuthCodeURLConfig struct {
	// AuthURL 授权地址。
	AuthURL string
	// ClientID 客户端 ID。
	ClientID string
	// ClientIDParam 客户端 ID 查询参数名，默认 client_id。
	ClientIDParam string
	// RedirectURL 授权回调地址。
	RedirectURL string
	// RedirectURLParam 授权回调地址查询参数名，默认 redirect_uri。
	RedirectURLParam string
	// ResponseType 授权响应类型，默认 code。
	ResponseType string
	// ResponseTypeParam 授权响应类型查询参数名，默认 response_type。
	ResponseTypeParam string
	// ScopeParam 授权作用域查询参数名，默认 scope。
	ScopeParam string
	// ScopeSeparator 多个授权作用域的连接符，默认空格。
	ScopeSeparator string
	// DefaultScopes 默认授权作用域。
	DefaultScopes []string
	// Fragment 授权地址片段。
	Fragment string
}

// AuthCodeOptions 定义授权地址生成选项。
type AuthCodeOptions struct {
	// Scopes 授权作用域列表。
	Scopes []string
	// Extra 附加查询参数。
	Extra url.Values
}

// AuthCodeOption 定义授权地址配置函数。
type AuthCodeOption func(*AuthCodeOptions)

// ApplyAuthCodeOptions 应用授权地址配置函数。
func ApplyAuthCodeOptions(opts ...AuthCodeOption) AuthCodeOptions {
	options := AuthCodeOptions{Extra: make(url.Values)}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// BuildAuthCodeURL 构造授权码登录地址。
func BuildAuthCodeURL(config AuthCodeURLConfig, state string, opts ...AuthCodeOption) string {
	options := ApplyAuthCodeOptions(opts...)
	u, err := url.Parse(config.AuthURL)
	if err != nil {
		return ""
	}
	scopes := options.Scopes
	if len(scopes) == 0 {
		scopes = config.DefaultScopes
	}

	values := u.Query()
	values.Set(defaultString(config.ClientIDParam, "client_id"), config.ClientID)
	values.Set(defaultString(config.RedirectURLParam, "redirect_uri"), config.RedirectURL)
	values.Set(defaultString(config.ResponseTypeParam, "response_type"), defaultString(config.ResponseType, "code"))
	if len(scopes) > 0 {
		values.Set(defaultString(config.ScopeParam, "scope"), strings.Join(scopes, config.scopeSeparator()))
	}
	if state != "" {
		values.Set("state", state)
	}
	for key, vals := range options.Extra {
		for _, val := range vals {
			values.Add(key, val)
		}
	}

	u.RawQuery = values.Encode()
	if config.Fragment != "" {
		u.Fragment = config.Fragment
	}
	return u.String()
}

// WithScopes 设置授权作用域。
//
// 示例：
//
//	provider.AuthCodeURL("state", oauth2.WithScopes("snsapi_login"))
func WithScopes(scopes ...string) AuthCodeOption {
	return func(o *AuthCodeOptions) {
		o.Scopes = append([]string(nil), scopes...)
	}
}

// WithAuthParam 设置授权地址的附加查询参数。
//
// 示例：
//
//	provider.AuthCodeURL("state", oauth2.WithAuthParam("display", "mobile"))
func WithAuthParam(key, value string) AuthCodeOption {
	return func(o *AuthCodeOptions) {
		if o.Extra == nil {
			o.Extra = make(url.Values)
		}
		o.Extra.Set(key, value)
	}
}

func (c AuthCodeURLConfig) scopeSeparator() string {
	if c.ScopeSeparator == "" {
		return " "
	}
	return c.ScopeSeparator
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
