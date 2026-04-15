package logger

// Options 定义日志中间件配置
type Options struct {
	Logger      Logger
	SkipPaths   map[string]bool
	SkipMethods map[string]bool
}

func defaultOptions() *Options {
	return &Options{
		SkipPaths:   make(map[string]bool),
		SkipMethods: make(map[string]bool),
	}
}

// Option 配置日志中间件
type Option func(*Options)

// WithLogger 设置日志记录器实现
func WithLogger(l Logger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

// WithSkipPath 添加跳过日志记录的路径
func WithSkipPath(paths ...string) Option {
	return func(o *Options) {
		for _, p := range paths {
			o.SkipPaths[p] = true
		}
	}
}

// WithSkipMethod 添加跳过日志记录的 HTTP 方法
func WithSkipMethod(methods ...string) Option {
	return func(o *Options) {
		for _, m := range methods {
			o.SkipMethods[m] = true
		}
	}
}
