package metrics

import (
	"strconv"
	"time"

	"github.com/f2xme/gox/httpx"
)

// New 创建指标中间件。
func New(opts ...Option) httpx.Middleware {
	o := &Options{
		Collector:      NewMemoryCollector(),
		EnableDetailed: false,
		SkipPaths:      make(map[string]bool),
	}
	for _, opt := range opts {
		opt(o)
	}

	return func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			// Check if path should be skipped
			path := ctx.Path()
			if o.SkipPaths[path] {
				return next(ctx)
			}

			// Normalize path if normalizer is provided
			if o.PathNormalizer != nil {
				path = o.PathNormalizer(path)
			}

			method := ctx.Method()

			// Record start time
			start := time.Now()

			// Call next handler and capture error
			err := next(ctx)

			// Calculate duration
			duration := time.Since(start)

			// Record request and duration
			o.Collector.RecordRequest(method, path)
			o.Collector.RecordDuration(method, path, duration)

			// Record error if occurred
			if err != nil {
				o.Collector.RecordError(method, path)
			}

			// Record response size if detailed metrics enabled
			if o.EnableDetailed {
				// Try to get Content-Length from response
				if rw := ctx.ResponseWriter(); rw != nil {
					if cl := rw.Header().Get("Content-Length"); cl != "" {
						if size, parseErr := strconv.ParseInt(cl, 10, 64); parseErr == nil {
							o.Collector.RecordResponseSize(method, path, size)
						}
					}
				}
			}

			// Call business metrics function if provided
			if o.BusinessMetrics != nil {
				o.BusinessMetrics(ctx, o.Collector)
			}

			return err
		}
	}
}
