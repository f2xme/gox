package captcha

import "errors"

var (
	// ErrNotFound 表示验证码不存在或已过期
	ErrNotFound = errors.New("captcha: not found")

	// ErrNilStore 表示验证码存储为空
	ErrNilStore = errors.New("captcha: store is nil")

	// ErrNilGenerator 表示验证码生成器为空
	ErrNilGenerator = errors.New("captcha: generator is nil")

	// ErrInvalidID 表示无效的验证码 ID
	ErrInvalidID = errors.New("captcha: invalid id")

	// ErrInvalidTTL 表示过期时间无效
	ErrInvalidTTL = errors.New("captcha: ttl must be positive")

	// ErrInvalidConsumeMode 表示验证码消费策略无效
	ErrInvalidConsumeMode = errors.New("captcha: invalid consume mode")

	// ErrGenerateFailed 表示生成验证码失败
	ErrGenerateFailed = errors.New("captcha: failed to generate captcha")
)
