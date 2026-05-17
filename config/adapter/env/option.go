package env

import "time"

const defaultWatchInterval = time.Second

// Options 定义环境变量适配器的配置选项。
type Options struct {
	prefix        string
	defaults      map[string]any
	watch         bool
	watchInterval time.Duration
}

// Option 定义配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置选项。
func defaultOptions() Options {
	return Options{}
}

// WithPrefix 设置环境变量前缀。
//
// 前缀会自动转为大写，并去掉末尾的 "_"。
// 例如 prefix 为 "APP" 时，配置键 "server.port" 会读取环境变量 APP_SERVER_PORT。
func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.prefix = prefix
	}
}

// WithDefaults 设置配置键的默认值。
//
// 默认值仅在对应环境变量不存在时生效。
func WithDefaults(defaults map[string]any) Option {
	return func(o *Options) {
		o.defaults = defaults
	}
}

// WithWatch 启用环境变量变更监听。
//
// 环境变量没有文件系统事件，该选项会通过轮询当前进程环境变量快照检测变化。
// 不传 interval 时默认每秒检测一次；interval <= 0 时使用默认值。
func WithWatch(interval ...time.Duration) Option {
	return func(o *Options) {
		o.watch = true
		o.watchInterval = defaultWatchInterval
		if len(interval) > 0 && interval[0] > 0 {
			o.watchInterval = interval[0]
		}
	}
}
