package timeout

import (
	"context"
	"time"

	"github.com/f2xme/gox/httpx"
)

// New 创建超时中间件，取消超过指定时长的请求。
// 如果请求超时，默认返回 503 Service Unavailable 响应，
// 或调用通过 WithHandler 提供的自定义处理函数。
func New(opts ...Option) httpx.Middleware {
	o := &Options{
		Timeout: 30 * time.Second,
		Handler: nil,
	}
	for _, opt := range opts {
		opt(o)
	}

	return func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			// Create a context with timeout
			timeoutCtx, cancel := context.WithTimeout(ctx.Request().Context(), o.Timeout)
			defer cancel()

			// Replace request context with timeout context
			req := ctx.Request()
			*req = *req.WithContext(timeoutCtx)

			// Channel to receive handler result
			done := make(chan error, 1)

			// Run handler in goroutine
			go func() {
				done <- next(ctx)
			}()

			// Wait for either completion or timeout
			select {
			case err := <-done:
				// Handler completed normally
				return err

			case <-timeoutCtx.Done():
				// Request timed out
				if o.Handler != nil {
					o.Handler(ctx)
					return context.DeadlineExceeded
				}

				// Default timeout response
				ctx.Status(503)
				ctx.JSON(503, map[string]any{
					"error":   "Request Timeout",
					"message": "The request took too long to process",
				})
				return context.DeadlineExceeded
			}
		}
	}
}
