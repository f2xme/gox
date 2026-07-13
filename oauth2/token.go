package oauth2

import "time"

// Token 表示 OAuth2 访问令牌。
type Token struct {
	// AccessToken 接口调用凭证。
	AccessToken string
	// TokenType 令牌类型，平台未返回时通常为空。
	TokenType string
	// RefreshToken 用于刷新访问令牌的凭证。
	RefreshToken string
	// Expiry 访问令牌过期时间。
	Expiry time.Time
	// ExpiresIn 访问令牌有效秒数。
	ExpiresIn int64
	// OpenID 授权用户在当前应用下的唯一标识。
	OpenID string
	// UnionID 授权用户在平台开放体系下的统一标识。
	UnionID string
	// Scope 用户授权作用域。
	Scope string
	// Raw 保存平台原始响应，便于排查字段差异。
	Raw []byte
}

// TokenInfo 定义创建 Token 所需的字段。
type TokenInfo struct {
	// AccessToken 接口调用凭证。
	AccessToken string
	// TokenType 令牌类型。
	TokenType string
	// RefreshToken 用于刷新访问令牌的凭证。
	RefreshToken string
	// ExpiresIn 访问令牌有效秒数。
	ExpiresIn int64
	// OpenID 授权用户在当前应用下的唯一标识。
	OpenID string
	// UnionID 授权用户在平台开放体系下的统一标识。
	UnionID string
	// Scope 用户授权作用域。
	Scope string
	// Raw 保存平台原始响应。
	Raw []byte
}

// NewToken 使用标准字段创建 Token。
func NewToken(info TokenInfo) *Token {
	token := &Token{
		AccessToken:  info.AccessToken,
		TokenType:    info.TokenType,
		RefreshToken: info.RefreshToken,
		ExpiresIn:    info.ExpiresIn,
		OpenID:       info.OpenID,
		UnionID:      info.UnionID,
		Scope:        info.Scope,
		Raw:          info.Raw,
	}
	if info.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(info.ExpiresIn) * time.Second)
	}
	return token
}

// Expired 判断访问令牌是否已经过期。
func (t *Token) Expired() bool {
	if t == nil || t.Expiry.IsZero() {
		return false
	}
	return time.Now().After(t.Expiry)
}
