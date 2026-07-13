package oauth2

import (
	"context"
)

// Provider 定义第三方登录服务提供商接口。
type Provider interface {
	// Name 返回服务提供商名称。
	Name() string

	// AuthCodeURL 生成授权码登录地址。
	AuthCodeURL(state string, opts ...AuthCodeOption) string

	// Exchange 使用授权码换取访问令牌。
	Exchange(ctx context.Context, code string) (*Token, error)

	// RefreshToken 使用刷新令牌续期访问令牌。
	RefreshToken(ctx context.Context, refreshToken string) (*Token, error)

	// UserInfo 使用访问令牌获取用户信息。
	UserInfo(ctx context.Context, token *Token) (*User, error)
}

// User 表示第三方平台用户基础信息。
type User struct {
	// Provider 服务提供商名称。
	Provider string
	// ID 用户主标识，优先使用平台 openid。
	ID string
	// OpenID 用户在当前应用下的唯一标识。
	OpenID string
	// UnionID 用户在平台开放体系下的统一标识。
	UnionID string
	// Nickname 用户昵称。
	Nickname string
	// AvatarURL 用户头像地址。
	AvatarURL string
	// Gender 用户性别原始描述。
	Gender string
	// Country 用户所在国家或地区。
	Country string
	// Province 用户所在省份。
	Province string
	// City 用户所在城市。
	City string
	// Raw 保存平台原始响应，便于排查字段差异。
	Raw []byte
}
