package cache

import (
	"github.com/f2xme/gox/cache"
	"github.com/f2xme/gox/captcha"
)

// New 创建 cache 适配器；后端实现 cache.AtomicStore 时，同时提供 captcha.AtomicStore。
func New(c Backend, opts ...Option) captcha.Store {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	store := &cacheStore{
		cache: c,
		opts:  cfg,
	}
	if atomic, ok := c.(cache.AtomicStore); ok {
		return &atomicCacheStore{cacheStore: store, atomic: atomic}
	}
	return store
}
