package memory

import (
	"time"

	"github.com/f2xme/gox/captcha"
	"github.com/f2xme/gox/captcha/generator/base64"
)

// captchaConfig 用于收集生成器配置
type captchaConfig struct {
	generatorOpts []base64.Option
	captchaOpts   []captcha.Option
}

// WithCaptchaType 设置验证码类型（便捷选项）。
func WithCaptchaType(t base64.CaptchaType) captcha.Option {
	return func(opts *captcha.Options) {
		// 延迟创建生成器，收集所有配置
		if opts.Generator == nil {
			opts.Generator = base64.New(base64.WithType(t))
		}
	}
}

// WithLength 设置验证码长度（便捷选项）。
func WithLength(length int) captcha.Option {
	return func(opts *captcha.Options) {
		// 这个需要在生成器层面设置，暂时通过重新创建生成器
		// TODO: 改进设计以支持增量配置
		if opts.Generator == nil {
			opts.Generator = base64.New(base64.WithLength(length))
		}
	}
}

// WithSize 设置验证码尺寸（便捷选项）。
func WithSize(width, height int) captcha.Option {
	return func(opts *captcha.Options) {
		if opts.Generator == nil {
			opts.Generator = base64.New(base64.WithSize(width, height))
		}
	}
}

// WithNoiseCount 设置噪点数量（便捷选项）。
func WithNoiseCount(count int) captcha.Option {
	return func(opts *captcha.Options) {
		if opts.Generator == nil {
			opts.Generator = base64.New(base64.WithNoiseCount(count))
		}
	}
}

// WithLanguage 设置音频验证码语言（便捷选项）。
func WithLanguage(lang string) captcha.Option {
	return func(opts *captcha.Options) {
		if opts.Generator == nil {
			opts.Generator = base64.New(base64.WithLanguage(lang))
		}
	}
}

// WithCaptchaTTL 设置验证码过期时间（便捷选项）。
func WithCaptchaTTL(ttl time.Duration) captcha.Option {
	return captcha.WithTTL(ttl)
}

// WithCaptchaIDLength 设置验证码 ID 长度（便捷选项）。
func WithCaptchaIDLength(length int) captcha.Option {
	return captcha.WithIDLength(length)
}

// WithGenerator 设置自定义生成器（便捷选项）。
func WithGenerator(opts ...base64.Option) captcha.Option {
	return captcha.WithGenerator(base64.New(opts...))
}
