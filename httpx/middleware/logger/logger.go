package logger

import (
	"time"

	"github.com/f2xme/gox/httpx"
)

// Logger 定义请求日志记录接口
type Logger interface {
	Info(msg string, keysAndValues ...any)
}

// New 创建日志中间件
// 如果未通过 WithLogger 提供日志记录器，中间件将不执行任何操作
func New(opts ...Option) httpx.Middleware {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			if o.Logger == nil {
				return next(ctx)
			}
			if o.SkipPaths[ctx.Path()] {
				return next(ctx)
			}
			if o.SkipMethods[ctx.Method()] {
				return next(ctx)
			}

			start := time.Now()
			err := next(ctx)
			o.Logger.Info("request",
				"method", ctx.Method(),
				"path", ctx.Path(),
				"ip", ctx.ClientIP(),
				"duration", time.Since(start).String(),
				"error", err,
			)
			return err
		}
	}
}
