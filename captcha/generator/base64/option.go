package base64

import "errors"

var (
	// ErrInvalidWidth 表示宽度无效。
	ErrInvalidWidth = errors.New("captcha/base64: width must be positive")

	// ErrInvalidHeight 表示高度无效。
	ErrInvalidHeight = errors.New("captcha/base64: height must be positive")

	// ErrInvalidLength 表示长度无效。
	ErrInvalidLength = errors.New("captcha/base64: length must be positive")

	// ErrInvalidNoiseCount 表示噪点数量无效。
	ErrInvalidNoiseCount = errors.New("captcha/base64: noiseCount cannot be negative")

	// ErrInvalidType 表示验证码类型无效。
	ErrInvalidType = errors.New("captcha/base64: invalid captcha type")
)

// CaptchaType 定义验证码类型。
type CaptchaType int

const (
	// TypeDigit 数字验证码（0-9）
	TypeDigit CaptchaType = iota
	// TypeString 字母数字混合验证码（a-z, A-Z, 0-9）
	TypeString
	// TypeMath 算术表达式验证码（如 1+2=?）
	TypeMath
	// TypeAudio 音频验证码（语音朗读数字）
	TypeAudio
)

// String 返回验证码类型的字符串表示。
func (t CaptchaType) String() string {
	switch t {
	case TypeDigit:
		return "digit"
	case TypeString:
		return "string"
	case TypeMath:
		return "math"
	case TypeAudio:
		return "audio"
	default:
		return "unknown"
	}
}

// Options 定义 base64 生成器的配置选项。
type Options struct {
	// Type 验证码类型，默认 TypeDigit
	Type CaptchaType
	// Width 验证码图片宽度（像素），默认 240
	Width int
	// Height 验证码图片高度（像素），默认 80
	Height int
	// Length 验证码长度（字符数），默认 4
	Length int
	// NoiseCount 噪点数量，用于增加识别难度，默认 1
	NoiseCount int
	// Language 音频验证码语言，默认 "en"
	Language string
}

// Option 定义配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		Type:       TypeDigit,
		Width:      240,
		Height:     80,
		Length:     4,
		NoiseCount: 1,
		Language:   "en",
	}
}

// validate 校验生成器配置。
func (o Options) validate() error {
	if o.Type < TypeDigit || o.Type > TypeAudio {
		return ErrInvalidType
	}
	if o.Width <= 0 {
		return ErrInvalidWidth
	}
	if o.Height <= 0 {
		return ErrInvalidHeight
	}
	if o.Length <= 0 {
		return ErrInvalidLength
	}
	if o.NoiseCount < 0 {
		return ErrInvalidNoiseCount
	}
	return nil
}

// WithType 设置验证码类型。
//
// 示例：
//
//	base64.New(base64.WithType(base64.TypeMath))
func WithType(t CaptchaType) Option {
	return func(o *Options) {
		o.Type = t
	}
}

// WithSize 设置验证码尺寸。
//
// 示例：
//
//	base64.New(base64.WithSize(300, 100))
func WithSize(width, height int) Option {
	return func(o *Options) {
		o.Width = width
		o.Height = height
	}
}

// WithLength 设置验证码长度。
//
// 示例：
//
//	base64.New(base64.WithLength(6))
func WithLength(length int) Option {
	return func(o *Options) {
		o.Length = length
	}
}

// WithNoiseCount 设置噪点数量。
//
// 示例：
//
//	base64.New(base64.WithNoiseCount(5))
func WithNoiseCount(count int) Option {
	return func(o *Options) {
		o.NoiseCount = count
	}
}

// WithLanguage 设置音频验证码语言。
//
// 示例：
//
//	base64.New(base64.WithLanguage("en"))
func WithLanguage(lang string) Option {
	return func(o *Options) {
		o.Language = lang
	}
}
