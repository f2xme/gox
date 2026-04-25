package memory

import (
	"time"

	"github.com/f2xme/gox/captcha"
	"github.com/f2xme/gox/captcha/generator/base64"
)

// WithCaptchaType 设置验证码类型（便捷选项）。
func WithCaptchaType(t base64.CaptchaType) captcha.Option {
	return captcha.WithGenerator(base64.New(base64.WithType(t)))
}

// WithLength 设置验证码长度（便捷选项）。
func WithLength(length int) captcha.Option {
	return captcha.WithGenerator(base64.New(base64.WithLength(length)))
}

// WithSize 设置验证码尺寸（便捷选项）。
func WithSize(width, height int) captcha.Option {
	return captcha.WithGenerator(base64.New(base64.WithSize(width, height)))
}

// WithNoiseCount 设置噪点数量（便捷选项）。
func WithNoiseCount(count int) captcha.Option {
	return captcha.WithGenerator(base64.New(base64.WithNoiseCount(count)))
}

// WithLanguage 设置音频验证码语言（便捷选项）。
func WithLanguage(lang string) captcha.Option {
	return captcha.WithGenerator(base64.New(base64.WithLanguage(lang)))
}

// WithCaptchaTTL 设置验证码过期时间（便捷选项）。
func WithCaptchaTTL(ttl time.Duration) captcha.Option {
	return captcha.WithTTL(ttl)
}

// WithCaptchaIDLength 设置验证码 ID 长度（便捷选项）。
func WithCaptchaIDLength(length int) captcha.Option {
	return captcha.WithIDLength(length)
}
