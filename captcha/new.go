package captcha

import "log"

// New 创建 Captcha 实例（通用构造函数）。
// 注意：必须通过 WithGenerator 选项提供生成器，否则会退出程序。
func New(store Store, opts ...Option) Captcha {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	if options.Generator == nil {
		log.Fatalf("captcha: generator is required, use WithGenerator option or use adapter's NewCaptcha function")
	}

	return &captcha{
		store:     store,
		generator: options.Generator,
		opts:      options,
	}
}
