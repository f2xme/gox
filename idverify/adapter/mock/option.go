package mock

// Options 定义 mock 核验器配置。
type Options struct {
	// MismatchNames 这些姓名返回姓名与证件不一致。
	MismatchNames []string
	// InvalidIDNames 这些姓名返回证件无效 / 查无。
	InvalidIDNames []string
	// VerifyError 固定返回的系统错误（优先于姓名规则）。
	VerifyError error
}

// Option 配置函数。
type Option func(*Options)

func defaultOptions() Options {
	return Options{}
}

// WithMismatchNames 设置触发姓名不一致的姓名列表。
func WithMismatchNames(names ...string) Option {
	return func(o *Options) {
		o.MismatchNames = append([]string(nil), names...)
	}
}

// WithInvalidIDNames 设置触发证件无效的姓名列表。
func WithInvalidIDNames(names ...string) Option {
	return func(o *Options) {
		o.InvalidIDNames = append([]string(nil), names...)
	}
}

// WithVerifyError 设置 Verify 固定返回的系统错误。
func WithVerifyError(err error) Option {
	return func(o *Options) {
		o.VerifyError = err
	}
}
