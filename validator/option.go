package validator

// Options 定义验证器配置选项。
type Options struct {
	// FieldNameTag 指定用于错误消息字段名的结构体标签，默认使用 label。
	FieldNameTag string
	// DefaultLang 默认语言，Validate() 使用该语言。默认 zh。
	DefaultLang string
	// Locales 要注册的语言列表。空则注册默认集合（zh、en）。
	Locales []string
}

// Option 定义验证器配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		FieldNameTag: "label",
		DefaultLang:  LangZH,
		Locales:      nil, // nil → defaultLocales
	}
}

// WithFieldNameTag 设置错误消息中使用的字段名标签。
//
// 传入空字符串时，错误消息使用结构体字段名。
// 多语言接口建议使用 json 等稳定字段名；展示名由业务 i18n 层处理。
//
// 示例：
//
//	v := validator.New(validator.WithFieldNameTag("json"))
func WithFieldNameTag(tag string) Option {
	return func(o *Options) {
		o.FieldNameTag = tag
	}
}

// WithDefaultLang 设置 Validate() 使用的默认语言。
//
// 语言码会归一化（zh-CN → zh）。未注册的语言在 New 时回退到已注册语言之一。
//
// 示例：
//
//	v := validator.New(validator.WithDefaultLang(validator.LangEN))
func WithDefaultLang(lang string) Option {
	return func(o *Options) {
		o.DefaultLang = lang
	}
}

// WithLocales 设置要注册的语言列表。
//
// 当前内置支持 zh、en。未知语言会被忽略。
// 空切片或未设置时注册默认集合（zh、en）。
//
// 示例：
//
//	v := validator.New(validator.WithLocales(validator.LangZH, validator.LangEN))
func WithLocales(langs ...string) Option {
	return func(o *Options) {
		o.Locales = append([]string(nil), langs...)
	}
}
