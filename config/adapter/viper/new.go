package viper

import (
	"log"
	"strings"

	"github.com/f2xme/gox/config"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var envKeyReplacer = strings.NewReplacer(".", "_")

// New 创建一个由 viper 支持的 config.Config 实例
//
// 参数：
//   - file: 配置文件路径（必须包含文件扩展名，如 .yaml、.json）
//   - opts: 可选的配置选项函数
//
// 返回值：
//   - config.Config: 配置实例，实现了 config.Config 和 config.Watcher 接口
//   - error: 读取配置文件失败时返回错误
//
// 示例：
//
//	cfg, err := New("config.yaml",
//		WithEnvPrefix("APP"),
//		WithWatch(),
//	)
func New(file string, opts ...Option) (config.Config, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	v := viper.New()
	v.SetConfigFile(file)

	for k, val := range options.defaults {
		v.SetDefault(k, val)
	}

	if options.envPrefix != "" {
		v.SetEnvPrefix(options.envPrefix)
		v.SetEnvKeyReplacer(envKeyReplacer)
		v.AutomaticEnv()
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	vc := &viperConfig{v: v}

	if options.watch {
		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			vc.mu.RLock()
			for _, fn := range vc.watchFns {
				go fn()
			}
			vc.mu.RUnlock()
		})
	}

	return vc, nil
}

// MustNew 创建一个由 viper 支持的 config.Config 实例，失败时终止程序
//
// 该方法是 New 的便捷包装，适用于配置文件必须存在的场景。
// 如果读取配置失败，程序会通过 log.Fatalf 终止。
//
// 参数：
//   - file: 配置文件路径（必须包含文件扩展名）
//   - opts: 可选的配置选项函数
//
// 返回值：
//   - config.Config: 配置实例
//
// 示例：
//
//	cfg := MustNew("config.yaml", WithEnvPrefix("APP"))
func MustNew(file string, opts ...Option) config.Config {
	cfg, err := New(file, opts...)
	if err != nil {
		log.Fatalf("config: failed to read %s: %v", file, err)
	}
	return cfg
}
