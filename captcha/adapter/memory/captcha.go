package memory

import "github.com/f2xme/gox/captcha"

// NewCaptcha 创建使用内存存储的 Captcha 实例。
func NewCaptcha(opts ...captcha.Option) captcha.Captcha {
	store := New()
	return captcha.New(store, opts...)
}
