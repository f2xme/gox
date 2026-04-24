package validator

// Options 定义验证器配置选项。
type Options struct {
	// FieldNameTag 指定用于错误消息字段名的结构体标签，默认使用 label。
	FieldNameTag string
}

// Option 定义验证器配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		FieldNameTag: "label",
	}
}

// WithFieldNameTag 设置错误消息中使用的字段名标签。
//
// 传入空字符串时，错误消息使用结构体字段名。
//
// 示例：
//
//	v := validator.New(validator.WithFieldNameTag("json"))
func WithFieldNameTag(tag string) Option {
	return func(o *Options) {
		o.FieldNameTag = tag
	}
}
