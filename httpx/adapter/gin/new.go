package gin

import (
	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/httpx"
	ginframework "github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// New 创建一个基于 Gin 的 httpx.Engine 实例
//
// 示例：
//
//	engine := New()
//	engine := New(WithMode("debug"))
//	engine := New(WithValidator(myValidator))
func New(opts ...Option) httpx.Engine {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	ginframework.SetMode(o.Mode)

	// 设置全局 validator
	if o.Validator != nil {
		binding.Validator = o.Validator
	}

	return &ginEngine{
		engine:       ginframework.New(),
		errorHandler: httpx.DefaultErrorHandler,
	}
}

// NewWithConfig 从 config.Config 创建 httpx.Engine 实例
//
// 配置键：
//   - httpx.gin.mode (string): Gin 运行模式 - "debug", "release" 或 "test"（默认："release"）
//
// 示例：
//
//	engine := NewWithConfig(cfg)
func NewWithConfig(cfg config.Config) httpx.Engine {
	opts := []Option{}

	if mode := cfg.GetString("gin.mode"); mode != "" {
		opts = append(opts, WithMode(mode))
	}

	return New(opts...)
}
