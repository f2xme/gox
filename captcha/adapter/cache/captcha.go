package cache

import (
	"log"

	"github.com/f2xme/gox/captcha"
	"github.com/f2xme/gox/captcha/generator/base64"
)

// NewCaptcha 创建使用 cache 存储的 Service 实例。
func NewCaptcha(c Backend, opts ...CaptchaOption) (captcha.Service, error) {
	cfg := captchaConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	gen, err := base64.New(cfg.generatorOpts...)
	if err != nil {
		return nil, err
	}

	store := New(c)
	cfg.captchaOpts = append(cfg.captchaOpts, captcha.WithGenerator(gen))
	return captcha.New(store, cfg.captchaOpts...)
}

// MustNewCaptcha 创建使用 cache 存储的 Service 实例，创建失败时退出程序。
func MustNewCaptcha(c Backend, opts ...CaptchaOption) captcha.Service {
	svc, err := NewCaptcha(c, opts...)
	if err != nil {
		log.Fatalf("captcha/cache: create captcha failed: %v", err)
	}
	return svc
}
