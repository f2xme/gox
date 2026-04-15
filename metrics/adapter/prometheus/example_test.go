package prometheus_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/f2xme/gox/metrics"
	prometheus "github.com/f2xme/gox/metrics/adapter/prometheus"
)

func ExampleNew() {
	// 创建一个基本的 Prometheus adapter
	adapter := prometheus.New()

	ctx := context.Background()
	counter, _ := adapter.Counter(ctx, "requests_total", nil)
	counter.Inc(ctx)

	fmt.Println("Prometheus adapter created")
	// Output: Prometheus adapter created
}

func ExampleNew_withOptions() {
	// 使用选项创建 adapter
	adapter := prometheus.New(
		prometheus.WithNamespace("myapp"),
		prometheus.WithSubsystem("api"),
		prometheus.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 1, 10}),
	)
	_ = adapter

	ctx := context.Background()
	counter, _ := adapter.Counter(ctx, "requests_total", metrics.Labels{
		"method": "GET",
	})
	counter.Inc(ctx)

	fmt.Println("Configured adapter created")
	// Output: Configured adapter created
}

func ExamplePrometheusAdapter_Counter() {
	adapter := prometheus.New(prometheus.WithNamespace("myapp"))
	ctx := context.Background()

	// 创建带标签的计数器
	counter, _ := adapter.Counter(ctx, "http_requests_total", metrics.Labels{
		"method": "GET",
		"path":   "/api/users",
	})

	// 增加计数
	counter.Inc(ctx)
	counter.Add(ctx, 5)

	fmt.Println("Counter incremented")
	// Output: Counter incremented
}

func ExamplePrometheusAdapter_Gauge() {
	adapter := prometheus.New(prometheus.WithNamespace("myapp"))
	ctx := context.Background()

	// 创建仪表盘指标
	gauge, _ := adapter.Gauge(ctx, "active_connections", metrics.Labels{
		"type": "http",
	})

	// 设置值
	gauge.Set(ctx, 42)
	gauge.Inc(ctx)
	gauge.Dec(ctx)

	fmt.Println("Gauge updated")
	// Output: Gauge updated
}

func ExamplePrometheusAdapter_Histogram() {
	adapter := prometheus.New(prometheus.WithNamespace("myapp"))
	ctx := context.Background()

	// 创建直方图指标
	histogram, _ := adapter.Histogram(ctx, "request_duration_seconds", metrics.Labels{
		"endpoint": "/api/users",
	})

	// 记录观测值
	histogram.Observe(ctx, 0.123)
	histogram.Observe(ctx, 0.456)

	fmt.Println("Histogram observations recorded")
	// Output: Histogram observations recorded
}

func ExamplePrometheusAdapter_Handler() {
	adapter := prometheus.New()

	// 注册 /metrics 端点
	http.Handle("/metrics", adapter.Handler())

	fmt.Println("Metrics endpoint registered")
	// Output: Metrics endpoint registered
}

func Example_httpServer() {
	// 创建 adapter
	adapter := prometheus.New(
		prometheus.WithNamespace("myapp"),
		prometheus.WithSubsystem("http"),
	)

	ctx := context.Background()

	// 创建指标
	requestsTotal, _ := adapter.Counter(ctx, "requests_total", metrics.Labels{
		"method": "GET",
	})
	requestDuration, _ := adapter.Histogram(ctx, "request_duration_seconds", nil)
	activeConnections, _ := adapter.Gauge(ctx, "active_connections", nil)

	// 模拟请求处理
	activeConnections.Inc(ctx)
	start := time.Now()

	// ... 处理请求 ...

	duration := time.Since(start).Seconds()
	requestDuration.Observe(ctx, duration)
	requestsTotal.Inc(ctx)
	activeConnections.Dec(ctx)

	fmt.Println("Request processed")
	// Output: Request processed
}
