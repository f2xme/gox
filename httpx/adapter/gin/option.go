package gin

import (
	"errors"

	ginframework "github.com/gin-gonic/gin"
)

// Options 定义 Gin 适配器的配置选项
type Options struct {
	// Mode 设置 Gin 运行模式
	// 可选值：debug, release, test
	// 默认值：release
	Mode string
}

// Option 定义配置选项函数
type Option func(*Options)

// Validate 检查配置选项是否有效
func (o *Options) Validate() error {
	if o.Mode != "" && o.Mode != ginframework.DebugMode && o.Mode != ginframework.ReleaseMode && o.Mode != ginframework.TestMode {
		return errors.New("gin: invalid mode, must be debug, release, or test")
	}
	return nil
}

// defaultOptions 返回默认配置
func defaultOptions() *Options {
	return &Options{
		Mode: ginframework.ReleaseMode,
	}
}

// WithMode 设置 Gin 运行模式（debug/release/test）
//
// 示例：
//
//	New(WithMode("debug"))
func WithMode(mode string) Option {
	return func(o *Options) {
		o.Mode = mode
	}
}
