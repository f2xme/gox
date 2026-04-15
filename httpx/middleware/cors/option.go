package cors

// Options 配置 CORS 中间件。
type Options struct {
	allowOrigins     []string
	allowMethods     []string
	allowHeaders     []string
	exposeHeaders    []string
	allowCredentials bool
	maxAge           int
}

// Option 配置函数。
type Option func(*Options)

// WithOrigins 设置允许的来源。
func WithOrigins(origins []string) Option {
	return func(o *Options) {
		o.allowOrigins = origins
	}
}

// WithMethods 设置允许的 HTTP 方法。
func WithMethods(methods []string) Option {
	return func(o *Options) {
		o.allowMethods = methods
	}
}

// WithHeaders 设置允许的请求头。
func WithHeaders(headers []string) Option {
	return func(o *Options) {
		o.allowHeaders = headers
	}
}

// WithExposeHeaders 设置暴露给浏览器的响应头。
func WithExposeHeaders(headers []string) Option {
	return func(o *Options) {
		o.exposeHeaders = headers
	}
}

// WithCredentials 启用或禁用凭证支持。
func WithCredentials(allow bool) Option {
	return func(o *Options) {
		o.allowCredentials = allow
	}
}

// WithMaxAge 设置预检请求的缓存时间（秒）。
func WithMaxAge(maxAge int) Option {
	return func(o *Options) {
		o.maxAge = maxAge
	}
}

func defaultOptions() *Options {
	return &Options{
		allowOrigins: []string{"*"},
		allowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		allowHeaders: []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
	}
}
