package cache

import (
	"github.com/f2xme/gox/cache"
	"github.com/f2xme/gox/captcha"
)

// New 创建一个新的 cache 适配器。
func New(c cache.Cache, opts ...Option) captcha.Store {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	return &cacheStore{
		cache: c,
		opts:  cfg,
	}
}
