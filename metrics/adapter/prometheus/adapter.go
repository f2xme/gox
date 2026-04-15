package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/f2xme/gox/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusAdapter 实现 metrics.Metrics 接口，提供 Prometheus 监控支持
type PrometheusAdapter struct {
	registry  *prometheus.Registry
	namespace string
	subsystem string
	buckets   []float64
	metrics   sync.Map // key: metricKey, value: prometheus.Collector
}

// metricKey 用于缓存已创建的指标
type metricKey struct {
	name      string
	labelKeys string // sorted and joined label keys
}

// New 创建一个新的 PrometheusAdapter
func New(opts ...Option) *PrometheusAdapter {
	adapter := &PrometheusAdapter{
		registry: prometheus.NewRegistry(),
		buckets:  prometheus.DefBuckets,
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

// Counter 创建或获取一个 Counter 指标
func (a *PrometheusAdapter) Counter(ctx context.Context, name string, labels metrics.Labels) (metrics.Counter, error) {
	labelKeys := extractLabelKeys(labels)
	key := metricKey{
		name:      name,
		labelKeys: strings.Join(labelKeys, ","),
	}

	// 尝试从缓存获取
	if cached, ok := a.metrics.Load(key); ok {
		vec := cached.(*prometheus.CounterVec)
		return &prometheusCounter{
			vec:    vec,
			labels: extractLabelValues(labels, labelKeys),
		}, nil
	}

	// 创建新的 CounterVec
	vec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: a.namespace,
			Subsystem: a.subsystem,
			Name:      name,
			Help:      fmt.Sprintf("Counter metric: %s", name),
		},
		labelKeys,
	)

	// 注册到 Registry
	if err := a.registry.Register(vec); err != nil {
		// 检查是否是重复注册错误
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			// 使用已注册的 Collector
			vec = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			return nil, fmt.Errorf("failed to register counter: %w", err)
		}
	}

	// 缓存
	a.metrics.Store(key, vec)

	return &prometheusCounter{
		vec:    vec,
		labels: extractLabelValues(labels, labelKeys),
	}, nil
}

// Gauge 创建或获取一个 Gauge 指标
func (a *PrometheusAdapter) Gauge(ctx context.Context, name string, labels metrics.Labels) (metrics.Gauge, error) {
	labelKeys := extractLabelKeys(labels)
	key := metricKey{
		name:      name,
		labelKeys: strings.Join(labelKeys, ","),
	}

	// 尝试从缓存获取
	if cached, ok := a.metrics.Load(key); ok {
		vec := cached.(*prometheus.GaugeVec)
		return &prometheusGauge{
			vec:    vec,
			labels: extractLabelValues(labels, labelKeys),
		}, nil
	}

	// 创建新的 GaugeVec
	vec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: a.namespace,
			Subsystem: a.subsystem,
			Name:      name,
			Help:      fmt.Sprintf("Gauge metric: %s", name),
		},
		labelKeys,
	)

	// 注册到 Registry
	if err := a.registry.Register(vec); err != nil {
		// 检查是否是重复注册错误
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			// 使用已注册的 Collector
			vec = are.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			return nil, fmt.Errorf("failed to register gauge: %w", err)
		}
	}

	// 缓存
	a.metrics.Store(key, vec)

	return &prometheusGauge{
		vec:    vec,
		labels: extractLabelValues(labels, labelKeys),
	}, nil
}

// Histogram 创建或获取一个 Histogram 指标
func (a *PrometheusAdapter) Histogram(ctx context.Context, name string, labels metrics.Labels) (metrics.Histogram, error) {
	labelKeys := extractLabelKeys(labels)
	key := metricKey{
		name:      name,
		labelKeys: strings.Join(labelKeys, ","),
	}

	// 尝试从缓存获取
	if cached, ok := a.metrics.Load(key); ok {
		vec := cached.(*prometheus.HistogramVec)
		return &prometheusHistogram{
			vec:    vec,
			labels: extractLabelValues(labels, labelKeys),
		}, nil
	}

	// 创建新的 HistogramVec
	vec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: a.namespace,
			Subsystem: a.subsystem,
			Name:      name,
			Help:      fmt.Sprintf("Histogram metric: %s", name),
			Buckets:   a.buckets,
		},
		labelKeys,
	)

	// 注册到 Registry
	if err := a.registry.Register(vec); err != nil {
		// 检查是否是重复注册错误
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			// 使用已注册的 Collector
			vec = are.ExistingCollector.(*prometheus.HistogramVec)
		} else {
			return nil, fmt.Errorf("failed to register histogram: %w", err)
		}
	}

	// 缓存
	a.metrics.Store(key, vec)

	return &prometheusHistogram{
		vec:    vec,
		labels: extractLabelValues(labels, labelKeys),
	}, nil
}

// Handler 返回 HTTP handler，用于暴露 /metrics 端点
func (a *PrometheusAdapter) Handler() http.Handler {
	return promhttp.HandlerFor(a.registry, promhttp.HandlerOpts{})
}

// extractLabelKeys 从 Labels 中提取并排序 label keys
func extractLabelKeys(labels metrics.Labels) []string {
	if len(labels) == 0 {
		return nil
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// extractLabelValues 按照 labelKeys 的顺序提取 label values
func extractLabelValues(labels metrics.Labels, labelKeys []string) []string {
	if len(labelKeys) == 0 {
		return nil
	}

	values := make([]string, len(labelKeys))
	for i, key := range labelKeys {
		values[i] = labels[key]
	}
	return values
}
