package logx

type initOptions struct {
	async           bool
	asyncBufferSize int
}

// Option 定义全局 logger 初始化选项。
type Option func(*initOptions)

func defaultInitOptions() initOptions {
	return initOptions{asyncBufferSize: defaultAsyncBufferSize}
}

// WithAsync 启用全局日志异步打印。
//
// 启用后 Info、Warn、Error 会先复制日志字段并投递到后台队列，由后台
// worker 串行调用底层 Logger。队列未满时调用方会快速返回，队列满时会
// 等待可用空间以避免丢日志。Fatal 会等待队列处理完成后同步调用底层
// Logger，避免致命日志丢失。
func WithAsync() Option {
	return func(o *initOptions) {
		o.async = true
	}
}

// WithAsyncBufferSize 设置异步打印队列长度。
//
// size 小于等于 0 时使用默认队列长度。
func WithAsyncBufferSize(size int) Option {
	return func(o *initOptions) {
		if size > 0 {
			o.asyncBufferSize = size
		}
	}
}
