package base64

import (
	"context"
	"log"

	"github.com/mojocn/base64Captcha"

	"github.com/f2xme/gox/captcha"
)

// base64Generator 实现 Generator 接口，基于 base64Captcha 库。
type base64Generator struct {
	driver base64Captcha.Driver
	opts   Options
}

// New 创建一个新的 base64 生成器。
func New(opts ...Option) (captcha.Generator, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	if err := options.validate(); err != nil {
		return nil, err
	}

	return &base64Generator{
		driver: createDriver(options),
		opts:   options,
	}, nil
}

// MustNew 创建一个新的 base64 生成器，创建失败时退出程序。
func MustNew(opts ...Option) captcha.Generator {
	gen, err := New(opts...)
	if err != nil {
		log.Fatalf("captcha/base64: create generator failed: %v", err)
	}
	return gen
}

// Generate 生成验证码内容和答案。
func (g *base64Generator) Generate(ctx context.Context) (captcha.ChallengeData, error) {
	if err := ctx.Err(); err != nil {
		return captcha.ChallengeData{}, err
	}

	// 先生成答案
	_, _, answer := g.driver.GenerateIdQuestionAnswer()

	// 使用答案生成图片
	item, err := g.driver.DrawCaptcha(answer)
	if err != nil {
		return captcha.ChallengeData{}, err
	}

	return captcha.ChallengeData{
		Data:   item.EncodeB64string(),
		Answer: answer,
	}, nil
}

// Type 返回生成器类型标识。
func (g *base64Generator) Type() string {
	return "base64"
}

// createDriver 根据配置创建对应的验证码驱动。
func createDriver(opts Options) base64Captcha.Driver {
	const defaultMaxSkew = 0.7 // 数字验证码的最大倾斜度

	switch opts.Type {
	case TypeString:
		return base64Captcha.NewDriverString(
			opts.Height,
			opts.Width,
			opts.NoiseCount,
			base64Captcha.OptionShowHollowLine,
			opts.Length,
			base64Captcha.TxtNumbers+base64Captcha.TxtAlphabet,
			nil,
			nil,
			nil,
		)
	case TypeMath:
		return base64Captcha.NewDriverMath(
			opts.Height,
			opts.Width,
			opts.NoiseCount,
			base64Captcha.OptionShowHollowLine,
			nil,
			nil,
			nil,
		)
	case TypeAudio:
		return base64Captcha.NewDriverAudio(
			opts.Length,
			opts.Language,
		)
	default: // TypeDigit
		return base64Captcha.NewDriverDigit(
			opts.Height,
			opts.Width,
			opts.Length,
			defaultMaxSkew,
			opts.NoiseCount,
		)
	}
}
