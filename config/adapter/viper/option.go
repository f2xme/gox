package viper

// Options 定义 viper 适配器的配置选项
type Options struct {
	envPrefix string
	defaults  map[string]any
	watch     bool
}

// Option 定义配置选项函数
type Option func(*Options)

// defaultOptions 返回默认配置选项
func defaultOptions() Options {
	return Options{}
}

// WithEnvPrefix 设置环境变量前缀，启用自动环境变量绑定
//
// 环境变量键名会自动将配置键中的 "." 替换为 "_"，并添加前缀。
// 例如：prefix 为 "APP"，配置键 "server.port" 会绑定到环境变量 APP_SERVER_PORT
//
// 示例：
//
//	cfg, _ := New("config.yaml", WithEnvPrefix("MYAPP"))
//	// 配置键 "database.host" 会读取环境变量 MYAPP_DATABASE_HOST
func WithEnvPrefix(prefix string) Option {
	return func(o *Options) {
		o.envPrefix = prefix
	}
}

// WithDefaults 设置配置键的默认值
//
// 默认值仅在配置文件中不存在对应键时生效。
//
// 示例：
//
//	cfg, _ := New("config.yaml", WithDefaults(map[string]any{
//		"server.port": 8080,
//		"debug":       false,
//	}))
func WithDefaults(defaults map[string]any) Option {
	return func(o *Options) {
		o.defaults = defaults
	}
}

// WithWatch 启用配置文件热重载
//
// 通过 fsnotify 监听配置文件变更，当文件修改时会触发已注册的回调函数。
// 需要配合 Watch 方法注册回调函数使用。
//
// 示例：
//
//	cfg, _ := New("config.yaml", WithWatch())
//	cfg.(config.Watcher).Watch(func() {
//		log.Println("配置已更新")
//	})
func WithWatch() Option {
	return func(o *Options) {
		o.watch = true
	}
}
