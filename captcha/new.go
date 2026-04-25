package captcha

import "log"

// New 创建 Service 实例（通用构造函数）。
func New(store Store, opts ...Option) (Service, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	if store == nil {
		return nil, ErrNilStore
	}
	if options.Generator == nil {
		return nil, ErrNilGenerator
	}
	if err := options.validate(); err != nil {
		return nil, err
	}

	return &service{
		store:     store,
		generator: options.Generator,
		opts:      options,
	}, nil
}

// MustNew 创建 Service 实例，创建失败时退出程序。
func MustNew(store Store, opts ...Option) Service {
	c, err := New(store, opts...)
	if err != nil {
		log.Fatalf("captcha: create captcha failed: %v", err)
	}
	return c
}
