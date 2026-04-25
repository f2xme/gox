package cache

import (
	"github.com/f2xme/gox/cache"
	"github.com/f2xme/gox/captcha"
)

// NewCaptcha 创建使用 cache 存储的 Captcha 实例。
func NewCaptcha(c cache.Cache, opts ...captcha.Option) captcha.Captcha {
	store := New(c)
	return captcha.New(store, opts...)
}
