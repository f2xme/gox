package prometheus

// Option 是配置 PrometheusAdapter 的函数类型
type Option func(*PrometheusAdapter)

// WithNamespace 设置指标的 namespace 前缀
//
// 示例：
//
//	adapter := prometheus.New(
//		prometheus.WithNamespace("myapp"),
//	)
//	// 指标名称会变成: myapp_<name>
func WithNamespace(namespace string) Option {
	return func(a *PrometheusAdapter) {
		a.namespace = namespace
	}
}

// WithSubsystem 设置指标的 subsystem 前缀
//
// 示例：
//
//	adapter := prometheus.New(
//		prometheus.WithNamespace("myapp"),
//		prometheus.WithSubsystem("api"),
//	)
//	// 指标名称会变成: myapp_api_<name>
func WithSubsystem(subsystem string) Option {
	return func(a *PrometheusAdapter) {
		a.subsystem = subsystem
	}
}

// WithHistogramBuckets 设置 Histogram 的 buckets
//
// 默认使用 prometheus.DefBuckets: [.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10]
//
// 示例：
//
//	adapter := prometheus.New(
//		prometheus.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 1, 10}),
//	)
func WithHistogramBuckets(buckets []float64) Option {
	return func(a *PrometheusAdapter) {
		a.buckets = buckets
	}
}
