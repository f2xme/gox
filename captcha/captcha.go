// Package captcha 提供验证码生成和验证功能。
//
// 支持多种验证码类型：数字、字母、算术、音频。
// 使用 base64Captcha 库作为底层实现。
package captcha

import (
	"time"

	"github.com/mojocn/base64Captcha"
)

const (
	defaultMaxSkew = 0.7 // 数字验证码的最大倾斜度
)

// Captcha 定义验证码接口
type Captcha interface {
	// Generate 生成验证码，返回验证码 ID 和 base64 编码的图片/音频数据
	// 对于图形验证码，返回 base64 编码的 PNG 图片
	// 对于音频验证码，返回 base64 编码的 WAV 音频
	Generate() (id, b64s string, err error)

	// Verify 验证验证码答案
	// 验证成功后会自动删除验证码，防止重复使用
	// 如果验证码不存在、已过期或答案错误，返回 false
	Verify(id, answer string) bool
}

// captchaImpl 实现 Captcha 接口
type captchaImpl struct {
	store  base64Captcha.Store
	opts   Options
	driver base64Captcha.Driver
}

// New 创建一个新的验证码实例
//
// store 参数指定验证码存储后端，可以使用 NewMemoryStore 创建内存存储，
// 或实现 base64Captcha.Store 接口自定义存储（如 Redis）。
//
// opts 参数用于配置验证码选项，如类型、长度、尺寸等。
//
// 示例：
//
//	store := captcha.NewMemoryStore(100, 5*time.Minute)
//	c := captcha.New(store, captcha.WithType(captcha.TypeDigit), captcha.WithLength(6))
func New(store base64Captcha.Store, opts ...Option) Captcha {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &captchaImpl{
		store:  store,
		opts:   options,
		driver: createDriver(options),
	}
}

// NewMemoryStore 创建一个内存存储实例
//
// gcLimitNumber 指定 GC 时保留的验证码数量，当存储的验证码数量超过此值时，
// 会清理最旧的验证码。
//
// expiration 指定验证码的过期时间，过期后的验证码无法通过验证。
//
// 示例：
//
//	// 保留最多 1000 个验证码，5 分钟后过期
//	store := captcha.NewMemoryStore(1000, 5*time.Minute)
func NewMemoryStore(gcLimitNumber int, expiration time.Duration) base64Captcha.Store {
	return base64Captcha.NewMemoryStore(gcLimitNumber, expiration)
}

// Generate 生成验证码
//
// 根据配置的验证码类型生成相应的验证码：
//   - TypeDigit: 纯数字验证码
//   - TypeString: 字母数字混合验证码
//   - TypeMath: 算术表达式验证码
//   - TypeAudio: 音频验证码
//
// 返回值：
//   - id: 验证码唯一标识符，用于后续验证
//   - b64s: base64 编码的验证码数据（图片或音频）
//   - err: 生成失败时返回错误
func (c *captchaImpl) Generate() (string, string, error) {
	captcha := base64Captcha.NewCaptcha(c.driver, c.store)
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		return "", "", ErrGenerateFailed
	}
	return id, b64s, nil
}

// createDriver 根据配置创建对应的验证码驱动
func createDriver(opts Options) base64Captcha.Driver {
	switch opts.CaptchaType {
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
	default:
		return base64Captcha.NewDriverDigit(
			opts.Height,
			opts.Width,
			opts.Length,
			defaultMaxSkew,
			opts.NoiseCount,
		)
	}
}

// Verify 验证验证码答案
//
// 验证成功后会自动删除验证码，防止重复使用。
//
// 参数：
//   - id: 验证码 ID（由 Generate 返回）
//   - answer: 用户输入的答案
//
// 返回值：
//   - true: 验证成功
//   - false: 验证失败（ID 或答案为空、验证码不存在、已过期或答案错误）
func (c *captchaImpl) Verify(id, answer string) bool {
	if id == "" || answer == "" {
		return false
	}
	return c.store.Verify(id, answer, true)
}
