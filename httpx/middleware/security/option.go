package security

import (
	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/middleware/internal"
)

// Options 定义安全中间件配置
type Options struct {
	SecurityHeaders        map[string]string
	XSSProtection          bool
	SQLInjectionProtection bool
	CSRF                   *CSRFConfig
	AllowedHosts           map[string]bool
	ErrorHandler           func(ctx httpx.Context, code int, message string)
}

// Option 配置安全中间件
type Option func(*Options)

// CSRFConfig 配置 CSRF token 生成和验证
type CSRFConfig struct {
	TokenLength    int
	TokenLookup    string
	CookieName     string
	CookiePath     string
	CookieMaxAge   int
	CookieSecure   bool
	CookieSameSite string
}

func defaultOptions() *Options {
	return &Options{
		SecurityHeaders: map[string]string{
			"X-Content-Type-Options":    "nosniff",
			"X-Frame-Options":           "DENY",
			"X-XSS-Protection":          "1; mode=block",
			"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
			"Content-Security-Policy":   "default-src 'self'",
		},
		ErrorHandler: defaultErrorHandler,
	}
}

func defaultErrorHandler(ctx httpx.Context, code int, message string) {
	internal.JSONError(ctx, code, message)
}

func normalizeCSRFConfig(cfg *CSRFConfig) {
	if cfg == nil {
		return
	}
	if cfg.TokenLength == 0 {
		cfg.TokenLength = 32
	}
	if cfg.TokenLookup == "" {
		cfg.TokenLookup = "header:X-CSRF-Token"
	}
	if cfg.CookieName == "" {
		cfg.CookieName = "_csrf"
	}
	if cfg.CookiePath == "" {
		cfg.CookiePath = "/"
	}
	if cfg.CookieMaxAge == 0 {
		cfg.CookieMaxAge = 86400
	}
	if cfg.CookieSameSite == "" {
		cfg.CookieSameSite = "Strict"
	}
}

// WithSecurityHeaders 覆盖默认的安全响应头
func WithSecurityHeaders(headers map[string]string) Option {
	return func(c *Options) {
		copied := make(map[string]string, len(headers))
		for k, v := range headers {
			copied[k] = v
		}
		c.SecurityHeaders = copied
	}
}

// WithXSSProtection 启用或禁用 XSS 模式检测
func WithXSSProtection(enabled bool) Option {
	return func(c *Options) {
		c.XSSProtection = enabled
	}
}

// WithSQLInjectionProtection 启用或禁用 SQL 注入模式检测
func WithSQLInjectionProtection(enabled bool) Option {
	return func(c *Options) {
		c.SQLInjectionProtection = enabled
	}
}

// WithCSRFProtection 启用 CSRF 保护
func WithCSRFProtection(cfg CSRFConfig) Option {
	return func(c *Options) {
		copied := cfg
		c.CSRF = &copied
	}
}

// WithContentSecurityPolicy 覆盖 CSP 头的值
func WithContentSecurityPolicy(policy string) Option {
	return func(c *Options) {
		c.SecurityHeaders["Content-Security-Policy"] = policy
	}
}

// WithAllowedHosts 限制允许的主机名
func WithAllowedHosts(hosts ...string) Option {
	return func(c *Options) {
		c.AllowedHosts = make(map[string]bool, len(hosts))
		for _, host := range hosts {
			c.AllowedHosts[host] = true
		}
	}
}

// WithErrorHandler 覆盖默认的错误处理函数
func WithErrorHandler(fn func(ctx httpx.Context, code int, message string)) Option {
	return func(c *Options) {
		c.ErrorHandler = fn
	}
}
