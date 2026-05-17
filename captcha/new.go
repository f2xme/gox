package captcha

import "fmt"

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

// MustNew 创建 Service 实例，创建失败时 panic。
func MustNew(store Store, opts ...Option) Service {
	c, err := New(store, opts...)
	if err != nil {
		panic(fmt.Errorf("captcha: create captcha failed: %w", err))
	}
	return c
}
