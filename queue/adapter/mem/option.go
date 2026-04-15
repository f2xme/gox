package mem

// Options 定义内存队列的配置选项
type Options struct {
	// BufferSize 每个主题订阅的通道缓冲区大小
	// 默认值为 64
	BufferSize int
}

// defaultOptions 返回默认配置
func defaultOptions() Options {
	return Options{
		BufferSize: 64,
	}
}

// Option 配置选项函数
type Option func(*Options)

// WithBufferSize 设置每个主题订阅的通道缓冲区大小
// 值 <= 0 时默认为 64
//
// 示例：
//
//	mem.New(mem.WithBufferSize(128))
func WithBufferSize(size int) Option {
	return func(o *Options) {
		if size <= 0 {
			size = 64
		}
		o.BufferSize = size
	}
}
