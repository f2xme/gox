package session

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/middleware/auth"
	goxsession "github.com/f2xme/gox/session"
)

const (
	// DefaultCookieName 是默认会话 Cookie 名称。
	DefaultCookieName = "sid"
	// DefaultUIDKey 是 session.Values 中默认用户 ID 键。
	DefaultUIDKey = "user_id"
	// ClaimsSessionIDKey 是 Claims.Get 读取 session ID 使用的键。
	ClaimsSessionIDKey = "session_id"
)

// ErrInvalidSession 表示 session 认证失败。
var ErrInvalidSession = errors.New("session: invalid session")

// Claims 表示从 session 中解析出的认证声明。
type Claims struct {
	// SessionID 是当前会话 ID。
	SessionID string
	// UID 是当前用户 ID。
	UID int64
	// Values 是 session 中的业务数据。
	Values map[string]any
}

// GetUID 返回当前用户 ID。
func (c Claims) GetUID() int64 {
	return c.UID
}

// Get 从会话数据中读取指定键。
func (c Claims) Get(key string) (any, bool) {
	if key == ClaimsSessionIDKey {
		return c.SessionID, c.SessionID != ""
	}
	if key == DefaultUIDKey {
		return c.UID, c.UID > 0
	}
	v, ok := c.Values[key]
	return v, ok
}

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
func NewValidator(manager goxsession.Manager, opts ...ValidatorOption) Validator {
	v := Validator{
		Manager: manager,
	}
	for _, opt := range opts {
		opt(&v)
	}
	return v
}

// Validate 验证 session ID 并返回认证声明。
func (v Validator) Validate(ctx context.Context, sid string) (auth.Claims, error) {
	if sid == "" || v.Manager == nil {
		return nil, ErrInvalidSession
	}

	sess, err := v.Manager.Get(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidSession, err)
	}

	uidKey := v.UIDKey
	if uidKey == "" {
		uidKey = DefaultUIDKey
	}

	uid, err := parseUID(sess.Values[uidKey])
	if err != nil || uid <= 0 {
		return nil, ErrInvalidSession
	}

	if v.RefreshThreshold > 0 && time.Until(sess.ExpiresAt) <= v.RefreshThreshold {
		refreshed, refreshErr := v.Manager.Refresh(ctx, sid)
		if refreshErr != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidSession, refreshErr)
		}
		sess = refreshed
	}

	return Claims{
		SessionID: sess.ID,
		UID:       uid,
		Values:    sess.Values,
	}, nil
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

func parseUID(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		if uint64(v) > math.MaxInt64 {
			return 0, ErrInvalidSession
		}
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		if v > math.MaxInt64 {
			return 0, ErrInvalidSession
		}
		return int64(v), nil
	case float64:
		if v != float64(int64(v)) {
			return 0, ErrInvalidSession
		}
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, ErrInvalidSession
	}
}
