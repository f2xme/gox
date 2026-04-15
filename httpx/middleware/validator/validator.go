package validator

import (
	"net/http"
	"strconv"

	"github.com/f2xme/gox/httpx"
)

// New 创建验证中间件
func New(opts ...Option) httpx.Middleware {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}

	return func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			// Check max body size
			if cfg.MaxBodySize > 0 {
				if contentLength := ctx.Header("Content-Length"); contentLength != "" {
					size, err := strconv.ParseInt(contentLength, 10, 64)
					if err == nil && size > cfg.MaxBodySize {
						cfg.ErrorHandler(ctx, http.StatusRequestEntityTooLarge, "Request body too large")
						return nil
					}
				}
			}

			// Check allowed content types
			if len(cfg.AllowedTypes) > 0 {
				contentType := ctx.Header("Content-Type")
				if contentType == "" {
					cfg.ErrorHandler(ctx, http.StatusUnsupportedMediaType, "Content-Type header is required")
					return nil
				}
				if !cfg.AllowedTypes[contentType] {
					cfg.ErrorHandler(ctx, http.StatusUnsupportedMediaType, "Unsupported Content-Type")
					return nil
				}
			}

			// Check required headers
			for _, header := range cfg.RequiredHeaders {
				if ctx.Header(header) == "" {
					cfg.ErrorHandler(ctx, http.StatusBadRequest, "Missing required header: "+header)
					return nil
				}
			}

			// Run custom validators
			for _, validator := range cfg.CustomValidators {
				if err := validator(ctx); err != nil {
					cfg.ErrorHandler(ctx, http.StatusBadRequest, err.Error())
					return nil
				}
			}

			return next(ctx)
		}
	}
}
