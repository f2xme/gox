package base64

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

// WithType 设置验证码类型。
func WithType(t CaptchaType) Option {
	return func(o *Options) {
		o.Type = t
	}
}

// WithSize 设置验证码尺寸。
func WithSize(width, height int) Option {
	return func(o *Options) {
		o.Width = width
		o.Height = height
	}
}

// WithLength 设置验证码长度。
func WithLength(length int) Option {
	return func(o *Options) {
		o.Length = length
	}
}

// WithNoiseCount 设置噪点数量。
func WithNoiseCount(count int) Option {
	return func(o *Options) {
		o.NoiseCount = count
	}
}

// WithLanguage 设置音频验证码语言。
func WithLanguage(lang string) Option {
	return func(o *Options) {
		o.Language = lang
	}
}
