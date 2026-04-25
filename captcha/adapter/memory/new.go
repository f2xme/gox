package memory

import "github.com/f2xme/gox/captcha"

// New 创建一个新的内存适配器。
func New(opts ...Option) captcha.Store {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	s := &memoryStore{
		items:  make(map[string]*item),
		opts:   cfg,
		stopCh: make(chan struct{}),
	}

	go s.cleanupLoop()
	return s
}
