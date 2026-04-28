package session

import "time"

// MinIDLength 是安全会话 ID 的最小十六进制字符串长度。
const MinIDLength = 32

// Options 定义会话管理器配置。
type Options struct {
	// TTL 是会话默认有效期。
	// 默认 2 小时。
	TTL time.Duration
	// IDLength 是随机会话 ID 的十六进制字符串长度。
	// 默认 32。
	IDLength int
}

// Option 定义会话管理器配置函数。
type Option func(*Options)

func defaultOptions() Options {
	return Options{
		TTL:      2 * time.Hour,
		IDLength: MinIDLength,
	}
}

func (o Options) validate() error {
	if o.TTL <= 0 {
		return ErrInvalidTTL
	}
	if o.IDLength < MinIDLength {
		return ErrInvalidID
	}
	return nil
}

// WithTTL 设置会话默认有效期。
//
// 示例：
//
//	session.New(store, session.WithTTL(30*time.Minute))
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}

// WithIDLength 设置随机会话 ID 长度。
// 长度不能小于 MinIDLength。
//
// 示例：
//
//	session.New(store, session.WithIDLength(48))
func WithIDLength(length int) Option {
	return func(o *Options) {
		o.IDLength = length
	}
}
