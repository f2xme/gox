package metrics

// Options 配置指标中间件的选项。
type Options struct {
	// Collector 设置指标收集器。
	Collector Collector

	// EnableDetailed 启用详细指标收集（响应大小等）。
	EnableDetailed bool

	// CustomLabels 从请求上下文中提取自定义标签的函数。
	CustomLabels func(ctx any) map[string]string

	// BusinessMetrics 记录自定义业务指标的函数。
	BusinessMetrics func(ctx any, collector Collector)

	// SkipPaths 跳过指标收集的路径集合。
	SkipPaths map[string]bool

	// PathNormalizer 规范化请求路径的函数（例如 /users/123 -> /users/{id}）。
	PathNormalizer func(path string) string
}

// Option 配置指标中间件。
type Option func(*Options)

// WithCollector 设置指标收集器。
func WithCollector(collector Collector) Option {
	return func(o *Options) {
		o.Collector = collector
	}
}

// WithDetailedMetrics 启用或禁用详细指标收集。
func WithDetailedMetrics(enabled bool) Option {
	return func(o *Options) {
		o.EnableDetailed = enabled
	}
}

// WithCustomLabels 设置从请求上下文中提取自定义标签的函数。
func WithCustomLabels(fn func(ctx any) map[string]string) Option {
	return func(o *Options) {
		o.CustomLabels = fn
	}
}

// WithBusinessMetrics 设置记录自定义业务指标的函数。
func WithBusinessMetrics(fn func(ctx any, collector Collector)) Option {
	return func(o *Options) {
		o.BusinessMetrics = fn
	}
}

// WithSkipPaths 设置跳过指标收集的路径。
func WithSkipPaths(paths ...string) Option {
	return func(o *Options) {
		o.SkipPaths = make(map[string]bool, len(paths))
		for _, path := range paths {
			o.SkipPaths[path] = true
		}
	}
}

// WithPathNormalizer 设置规范化请求路径的函数。
func WithPathNormalizer(fn func(path string) string) Option {
	return func(o *Options) {
		o.PathNormalizer = fn
	}
}
