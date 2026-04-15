package captcha

// CaptchaType 定义验证码类型
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

// Options 定义验证码配置选项
type Options struct {
	// Width 验证码图片宽度（像素），默认 240
	Width int
	// Height 验证码图片高度（像素），默认 80
	Height int
	// Length 验证码长度（字符数），默认 4
	Length int
	// NoiseCount 噪点数量，用于增加识别难度，默认 1
	NoiseCount int
	// CaptchaType 验证码类型，默认 TypeDigit
	CaptchaType CaptchaType
	// Language 音频验证码语言，默认 "en"
	Language string
}

// Option 定义配置选项函数
type Option func(*Options)

// Validate 验证配置选项的有效性
func (o *Options) Validate() error {
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
	if o.CaptchaType < TypeDigit || o.CaptchaType > TypeAudio {
		return ErrInvalidType
	}
	return nil
}

// defaultOptions 返回默认配置
func defaultOptions() Options {
	return Options{
		Width:       240,
		Height:      80,
		Length:      4,
		NoiseCount:  1,
		CaptchaType: TypeDigit,
		Language:    "en",
	}
}

// WithWidth 设置验证码宽度
//
// 示例：
//
//	captcha.New(store, captcha.WithWidth(300))
func WithWidth(width int) Option {
	return func(o *Options) {
		o.Width = width
	}
}

// WithHeight 设置验证码高度
//
// 示例：
//
//	captcha.New(store, captcha.WithHeight(100))
func WithHeight(height int) Option {
	return func(o *Options) {
		o.Height = height
	}
}

// WithLength 设置验证码长度
//
// 示例：
//
//	captcha.New(store, captcha.WithLength(6))
func WithLength(length int) Option {
	return func(o *Options) {
		o.Length = length
	}
}

// WithNoiseCount 设置噪点数量
//
// 噪点数量越多，验证码越难识别，但也会影响用户体验。
//
// 示例：
//
//	captcha.New(store, captcha.WithNoiseCount(5))
func WithNoiseCount(count int) Option {
	return func(o *Options) {
		o.NoiseCount = count
	}
}

// WithType 设置验证码类型
//
// 示例：
//
//	captcha.New(store, captcha.WithType(captcha.TypeMath))
func WithType(t CaptchaType) Option {
	return func(o *Options) {
		o.CaptchaType = t
	}
}

// WithLanguage 设置音频验证码语言
//
// 示例：
//
//	captcha.New(store, captcha.WithLanguage("zh"))
func WithLanguage(lang string) Option {
	return func(o *Options) {
		o.Language = lang
	}
}
