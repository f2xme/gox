package security

import (
	"net/http"

	"github.com/f2xme/gox/httpx"
)

// New 创建安全中间件
func New(opts ...Option) httpx.Middleware {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}
	normalizeCSRFConfig(cfg.CSRF)

	return func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			for key, value := range cfg.SecurityHeaders {
				ctx.SetHeader(key, value)
			}

			if len(cfg.AllowedHosts) > 0 {
				host := ctx.Request().Host
				if !cfg.AllowedHosts[host] {
					cfg.ErrorHandler(ctx, http.StatusBadRequest, "Host is not allowed")
					return nil
				}
			}

			if cfg.XSSProtection && hasSuspiciousQuery(ctx, containsXSSPayload) {
				cfg.ErrorHandler(ctx, http.StatusBadRequest, "Request contains suspicious XSS content")
				return nil
			}

			if cfg.SQLInjectionProtection && hasSuspiciousQuery(ctx, containsSQLInjectionPayload) {
				cfg.ErrorHandler(ctx, http.StatusBadRequest, "Request contains suspicious SQL content")
				return nil
			}

			if cfg.CSRF != nil {
				if _, err := ensureCSRFCookie(ctx, cfg.CSRF); err != nil {
					cfg.ErrorHandler(ctx, http.StatusBadRequest, "Failed to create CSRF token")
					return nil
				}
				if requiresCSRFValidation(ctx.Method()) && !validateCSRFToken(ctx, cfg.CSRF) {
					cfg.ErrorHandler(ctx, http.StatusBadRequest, "Invalid CSRF token")
					return nil
				}
			}

			return next(ctx)
		}
	}
}

func hasSuspiciousQuery(ctx httpx.Context, fn func(string) bool) bool {
	rawQuery := ctx.Request().URL.Query()
	for _, values := range rawQuery {
		for _, value := range values {
			if fn(value) {
				return true
			}
		}
	}
	return false
}

func requiresCSRFValidation(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}
