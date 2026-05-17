package session

import "time"

// ValidatorOption 定义会话认证验证器配置函数。
type ValidatorOption func(*Validator)

// WithValidatorUIDKey 设置 session.Values 中保存用户 ID 的键。
func WithValidatorUIDKey(key string) ValidatorOption {
	return func(v *Validator) {
		v.UIDKey = key
	}
}

// WithRefreshThreshold 设置滑动过期刷新窗口。
//
// 值大于 0 时，如果会话剩余有效期小于等于该值，Validate 会刷新会话。
func WithRefreshThreshold(threshold time.Duration) ValidatorOption {
	return func(v *Validator) {
		v.RefreshThreshold = threshold
	}
}
