package captcha

import "time"

// Options 定义验证码配置选项。
type Options struct {
	// TTL 默认过期时间，默认 5 分钟
	TTL time.Duration
	// IDLength ID 长度，默认 20
	IDLength int
	// Generator 自定义生成器
	Generator Generator
	// ConsumeMode 验证码消费策略，默认 ConsumeOnSuccess
	ConsumeMode ConsumeMode
}

// ConsumeMode 定义验证码验证后的消费策略。
type ConsumeMode int

const (
	// ConsumeOnSuccess 表示验证成功后删除验证码
	ConsumeOnSuccess ConsumeMode = iota
	// ConsumeAlways 表示只要验证码存在，验证后总是删除验证码
	ConsumeAlways
	// ConsumeNever 表示验证后不自动删除验证码
	ConsumeNever
)

// Option 定义配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		TTL:         5 * time.Minute,
		IDLength:    20,
		ConsumeMode: ConsumeOnSuccess,
		// Generator 将在 New() 中设置默认值
	}
}

// validate 校验配置选项。
func (o Options) validate() error {
	if o.TTL <= 0 {
		return ErrInvalidTTL
	}
	if o.IDLength <= 0 {
		return ErrInvalidID
	}
	if o.ConsumeMode < ConsumeOnSuccess || o.ConsumeMode > ConsumeNever {
		return ErrInvalidConsumeMode
	}
	return nil
}

// WithTTL 设置默认过期时间。
//
// 示例：
//
//	captcha.New(store, captcha.WithTTL(10*time.Minute), captcha.WithGenerator(gen))
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}

// WithIDLength 设置 ID 长度。
//
// 示例：
//
//	captcha.New(store, captcha.WithIDLength(32), captcha.WithGenerator(gen))
func WithIDLength(length int) Option {
	return func(o *Options) {
		o.IDLength = length
	}
}

// WithGenerator 设置自定义生成器。
//
// 示例：
//
//	captcha.New(store, captcha.WithGenerator(gen))
func WithGenerator(gen Generator) Option {
	return func(o *Options) {
		o.Generator = gen
	}
}

// WithConsumeMode 设置验证码消费策略。
//
// 示例：
//
//	captcha.New(store, captcha.WithGenerator(gen), captcha.WithConsumeMode(captcha.ConsumeAlways))
func WithConsumeMode(mode ConsumeMode) Option {
	return func(o *Options) {
		o.ConsumeMode = mode
	}
}
