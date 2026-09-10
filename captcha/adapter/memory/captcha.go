package memory

import (
	"github.com/f2xme/gox/captcha"
	"github.com/f2xme/gox/captcha/generator/base64"
)

// Captcha 是拥有内存存储的验证码服务，使用完毕后应调用 Close。
type Captcha struct {
	captcha.Service
	store *memoryStore
}

// Close 停止服务所拥有的存储清理协程，可重复调用。
func (c *Captcha) Close() error { return c.store.Close() }

// NewCaptcha 创建使用内存存储的服务；调用方负责调用 Close。
func NewCaptcha(opts ...CaptchaOption) (*Captcha, error) {
	cfg := captchaConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	gen, err := base64.New(cfg.generatorOpts...)
	if err != nil {
		return nil, err
	}

	store := New().(*memoryStore)
	cfg.captchaOpts = append(cfg.captchaOpts, captcha.WithGenerator(gen))
	svc, err := captcha.New(store, cfg.captchaOpts...)
	if err != nil {
		_ = store.Close()
		return nil, err
	}
	return &Captcha{Service: svc, store: store}, nil
}
