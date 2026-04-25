package cache

import (
	"time"

	"github.com/f2xme/gox/captcha"
	"github.com/f2xme/gox/captcha/generator/base64"
)

// captchaConfig 用于收集生成器配置。
type captchaConfig struct {
	generatorOpts []base64.Option
	captchaOpts   []captcha.Option
}

// CaptchaOption 定义 cache 验证码便捷配置选项。
type CaptchaOption func(*captchaConfig)

// WithCaptchaType 设置验证码类型（便捷选项）。
//
// 示例：
//
//	cacheadapter.NewCaptcha(c, cacheadapter.WithCaptchaType(base64.TypeDigit))
func WithCaptchaType(t base64.CaptchaType) CaptchaOption {
	return func(cfg *captchaConfig) {
		cfg.generatorOpts = append(cfg.generatorOpts, base64.WithType(t))
	}
}

// WithLength 设置验证码长度（便捷选项）。
//
// 示例：
//
//	cacheadapter.NewCaptcha(c, cacheadapter.WithLength(6))
func WithLength(length int) CaptchaOption {
	return func(cfg *captchaConfig) {
		cfg.generatorOpts = append(cfg.generatorOpts, base64.WithLength(length))
	}
}

// WithSize 设置验证码尺寸（便捷选项）。
//
// 示例：
//
//	cacheadapter.NewCaptcha(c, cacheadapter.WithSize(300, 100))
func WithSize(width, height int) CaptchaOption {
	return func(cfg *captchaConfig) {
		cfg.generatorOpts = append(cfg.generatorOpts, base64.WithSize(width, height))
	}
}

// WithNoiseCount 设置噪点数量（便捷选项）。
//
// 示例：
//
//	cacheadapter.NewCaptcha(c, cacheadapter.WithNoiseCount(5))
func WithNoiseCount(count int) CaptchaOption {
	return func(cfg *captchaConfig) {
		cfg.generatorOpts = append(cfg.generatorOpts, base64.WithNoiseCount(count))
	}
}

// WithLanguage 设置音频验证码语言（便捷选项）。
//
// 示例：
//
//	cacheadapter.NewCaptcha(c, cacheadapter.WithLanguage("en"))
func WithLanguage(lang string) CaptchaOption {
	return func(cfg *captchaConfig) {
		cfg.generatorOpts = append(cfg.generatorOpts, base64.WithLanguage(lang))
	}
}

// WithCaptchaTTL 设置验证码过期时间（便捷选项）。
//
// 示例：
//
//	cacheadapter.NewCaptcha(c, cacheadapter.WithCaptchaTTL(10*time.Minute))
func WithCaptchaTTL(ttl time.Duration) CaptchaOption {
	return func(cfg *captchaConfig) {
		cfg.captchaOpts = append(cfg.captchaOpts, captcha.WithTTL(ttl))
	}
}

// WithCaptchaIDLength 设置验证码 ID 长度（便捷选项）。
//
// 示例：
//
//	cacheadapter.NewCaptcha(c, cacheadapter.WithCaptchaIDLength(32))
func WithCaptchaIDLength(length int) CaptchaOption {
	return func(cfg *captchaConfig) {
		cfg.captchaOpts = append(cfg.captchaOpts, captcha.WithIDLength(length))
	}
}
