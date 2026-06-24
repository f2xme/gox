package memory

// Options 定义内存短信适配器配置选项。
type Options struct {
	// SendError 发送短信时固定返回的错误，用于测试失败分支。
	SendError error
}

// Option 定义配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置选项。
func defaultOptions() Options {
	return Options{}
}

// WithSendError 设置发送短信时固定返回的错误。
//
// 示例：
//
//	New(WithSendError(errors.New("send failed")))
func WithSendError(err error) Option {
	return func(o *Options) {
		o.SendError = err
	}
}
