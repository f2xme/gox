package requestid

import "github.com/f2xme/gox/httpx"

// New 创建请求 ID 中间件
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
			ctx.SetHeader(o.headerKey, id)
			ctx.Set("request_id", id)

			if o.handler != nil {
				o.handler(ctx, id)
			}

			return next(ctx)
		}
	}
}

// Get 从上下文头中获取请求 ID
func Get(ctx httpx.Context) string {
	return ctx.Header(defaultHeaderKey).String()
}
