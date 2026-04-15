package captcha

import "errors"

var (
	// ErrGenerateFailed 表示生成验证码失败
	ErrGenerateFailed = errors.New("captcha: failed to generate captcha")

	// ErrInvalidWidth 表示宽度无效
	ErrInvalidWidth = errors.New("captcha: width must be positive")

	// ErrInvalidHeight 表示高度无效
	ErrInvalidHeight = errors.New("captcha: height must be positive")

	// ErrInvalidLength 表示长度无效
	ErrInvalidLength = errors.New("captcha: length must be positive")

	// ErrInvalidNoiseCount 表示噪点数量无效
	ErrInvalidNoiseCount = errors.New("captcha: noiseCount cannot be negative")

	// ErrInvalidType 表示验证码类型无效
	ErrInvalidType = errors.New("captcha: invalid captcha type")
)
