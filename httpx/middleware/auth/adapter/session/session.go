package session

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/middleware/auth"
	goxsession "github.com/f2xme/gox/session"
)

const (
	// DefaultCookieName 是默认会话 Cookie 名称。
	DefaultCookieName = "sid"
	// DefaultUIDKey 是 session.Values 中默认用户 ID 键。
	DefaultUIDKey = goxsession.DefaultUIDKey
)

// ErrInvalidSession 表示 session 认证失败。
var ErrInvalidSession = goxsession.ErrInvalidSession

// Claims 表示从 session 中解析出的认证声明。
type Claims = goxsession.Claims

// Validator 使用 session.Manager 验证 session ID 并生成 Claims。
type Validator struct {
	// Manager 是 session 管理器。
	Manager goxsession.Manager
	// UIDKey 是 session.Values 中保存用户 ID 的键。
	// 为空时使用 DefaultUIDKey。
	UIDKey string
	// RefreshThreshold 是滑动过期刷新窗口。
	// 值大于 0 时，如果会话剩余有效期小于等于该值，Validate 会刷新会话。
	RefreshThreshold time.Duration
}

var _ auth.Validator = (*Validator)(nil)

// NewValidator 创建基于 session 的 auth.Validator。
func NewValidator(manager goxsession.Manager, opts ...goxsession.ValidatorOption) Validator {
	v := goxsession.NewValidator(manager, opts...)
	return Validator{
		Manager:          v.Manager,
		UIDKey:           v.UIDKey,
		RefreshThreshold: v.RefreshThreshold,
	}
}

// Validate 验证 session ID 并返回认证声明。
func (v Validator) Validate(ctx context.Context, sid string) (auth.Claims, error) {
	return goxsession.Validator{
		Manager:          v.Manager,
		UIDKey:           v.UIDKey,
		RefreshThreshold: v.RefreshThreshold,
	}.Validate(ctx, sid)
}

// NewExtractor 创建从 Cookie 中提取 session ID 的 token 提取函数。
func NewExtractor(name string) func(httpx.Context) string {
	if name == "" {
		name = DefaultCookieName
	}
	return func(ctx httpx.Context) string {
		cookie, err := ctx.Cookie(name)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				return ""
			}
			return ""
		}
		return cookie.Value
	}
}
