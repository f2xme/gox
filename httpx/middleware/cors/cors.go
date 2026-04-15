package cors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/f2xme/gox/httpx"
)

// New 创建 CORS 中间件。
func New(opts ...Option) httpx.Middleware {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			origin := ctx.Header("Origin")
			if origin == "" {
				return next(ctx)
			}

			allowed := false
			for _, allowedOrigin := range o.allowOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			if !allowed {
				return next(ctx)
			}

			ctx.SetHeader("Access-Control-Allow-Origin", origin)
			if o.allowCredentials {
				ctx.SetHeader("Access-Control-Allow-Credentials", "true")
			}
			if len(o.exposeHeaders) > 0 {
				ctx.SetHeader("Access-Control-Expose-Headers", strings.Join(o.exposeHeaders, ", "))
			}

			if ctx.Method() == http.MethodOptions {
				if len(o.allowMethods) > 0 {
					ctx.SetHeader("Access-Control-Allow-Methods", strings.Join(o.allowMethods, ", "))
				}
				if len(o.allowHeaders) > 0 {
					ctx.SetHeader("Access-Control-Allow-Headers", strings.Join(o.allowHeaders, ", "))
				}
				if o.maxAge > 0 {
					ctx.SetHeader("Access-Control-Max-Age", strconv.Itoa(o.maxAge))
				}
				return ctx.NoContent(http.StatusNoContent)
			}

			return next(ctx)
		}
	}
}
