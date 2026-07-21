package traceid

import "github.com/f2xme/gox/httpx"

// New 创建 Trace ID 中间件。
func New(opts ...Option) httpx.Middleware {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			id := ctx.Header(o.headerKey).String()
			if id == "" {
				id = o.generator()
			}

			ctx.Set(contextKey, id)
			ctx.SetHeader(o.headerKey, id)
			return next(ctx)
		}
	}
}

// Get 从上下文获取 Trace ID。
func Get(ctx httpx.Context) string {
	if value, ok := ctx.Get(contextKey); ok {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ctx.Header(defaultHeaderKey).String()
}
