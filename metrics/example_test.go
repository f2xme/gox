package metrics_test

import (
	"context"
	"fmt"

	"github.com/f2xme/gox/metrics"
)

// mockMetrics 是用于示例的简单实现
type mockMetrics struct{}

func (m *mockMetrics) Counter(ctx context.Context, name string, labels metrics.Labels) (metrics.Counter, error) {
	return &mockCounter{name: name, labels: labels}, nil
}

func (m *mockMetrics) Gauge(ctx context.Context, name string, labels metrics.Labels) (metrics.Gauge, error) {
	return &mockGauge{name: name, labels: labels}, nil
}

func (m *mockMetrics) Histogram(ctx context.Context, name string, labels metrics.Labels) (metrics.Histogram, error) {
	return &mockHistogram{name: name, labels: labels}, nil
}

type mockCounter struct {
	name   string
	labels metrics.Labels
	value  float64
}

func (c *mockCounter) Inc(ctx context.Context) error {
	c.value++
	fmt.Printf("Counter %s incremented to %.0f\n", c.name, c.value)
	return nil
}

func (c *mockCounter) Add(ctx context.Context, delta float64) error {
	c.value += delta
	fmt.Printf("Counter %s increased by %.0f to %.0f\n", c.name, delta, c.value)
	return nil
}

type mockGauge struct {
	name   string
	labels metrics.Labels
	value  float64
}

func (g *mockGauge) Set(ctx context.Context, value float64) error {
	g.value = value
	fmt.Printf("Gauge %s set to %.0f\n", g.name, value)
	return nil
}

func (g *mockGauge) Inc(ctx context.Context) error {
	g.value++
	fmt.Printf("Gauge %s incremented to %.0f\n", g.name, g.value)
	return nil
}

func (g *mockGauge) Dec(ctx context.Context) error {
	g.value--
	fmt.Printf("Gauge %s decremented to %.0f\n", g.name, g.value)
	return nil
}

type mockHistogram struct {
	name   string
	labels metrics.Labels
}

func (h *mockHistogram) Observe(ctx context.Context, value float64) error {
	fmt.Printf("Histogram %s observed value %.3f\n", h.name, value)
	return nil
}

// Example_counter 演示如何使用 Counter 记录累计数量
func Example_counter() {
	m := &mockMetrics{}
	ctx := context.Background()

	// 创建计数器
	counter, _ := m.Counter(ctx, "http_requests_total", metrics.Labels{
		"method": "GET",
		"path":   "/api/users",
	})

	// 每次请求增加 1
	counter.Inc(ctx)
	counter.Inc(ctx)

	// 批量增加
	counter.Add(ctx, 5)

	// Output:
	// Counter http_requests_total incremented to 1
	// Counter http_requests_total incremented to 2
	// Counter http_requests_total increased by 5 to 7
}

// Example_gauge 演示如何使用 Gauge 记录当前状态
func Example_gauge() {
	m := &mockMetrics{}
	ctx := context.Background()

	// 创建仪表盘
	gauge, _ := m.Gauge(ctx, "active_connections", nil)

	// 设置当前值
	gauge.Set(ctx, 10)

	// 连接建立时增加
	gauge.Inc(ctx)
	gauge.Inc(ctx)

	// 连接关闭时减少
	gauge.Dec(ctx)

	// Output:
	// Gauge active_connections set to 10
	// Gauge active_connections incremented to 11
	// Gauge active_connections incremented to 12
	// Gauge active_connections decremented to 11
}

// Example_histogram 演示如何使用 Histogram 记录数值分布
func Example_histogram() {
	m := &mockMetrics{}
	ctx := context.Background()

	// 创建直方图
	histogram, _ := m.Histogram(ctx, "http_request_duration_seconds", metrics.Labels{
		"method": "GET",
	})

	// 记录不同的请求耗时
	histogram.Observe(ctx, 0.010)
	histogram.Observe(ctx, 0.050)
	histogram.Observe(ctx, 0.120)

	// Output:
	// Histogram http_request_duration_seconds observed value 0.010
	// Histogram http_request_duration_seconds observed value 0.050
	// Histogram http_request_duration_seconds observed value 0.120
}

// Example_labels 演示如何使用标签为指标添加维度
func Example_labels() {
	m := &mockMetrics{}
	ctx := context.Background()

	// 为不同的 HTTP 方法和状态码创建计数器
	getSuccess, _ := m.Counter(ctx, "api_requests_total", metrics.Labels{
		"method": "GET",
		"status": "200",
	})

	postSuccess, _ := m.Counter(ctx, "api_requests_total", metrics.Labels{
		"method": "POST",
		"status": "201",
	})

	getError, _ := m.Counter(ctx, "api_requests_total", metrics.Labels{
		"method": "GET",
		"status": "404",
	})

	// 记录不同类型的请求
	getSuccess.Inc(ctx)
	postSuccess.Inc(ctx)
	getError.Inc(ctx)

	// Output:
	// Counter api_requests_total incremented to 1
	// Counter api_requests_total incremented to 1
	// Counter api_requests_total incremented to 1
}

// Example_httpMiddleware 演示如何在 HTTP 中间件中使用指标
func Example_httpMiddleware() {
	m := &mockMetrics{}
	ctx := context.Background()

	// 创建指标
	requestsTotal, _ := m.Counter(ctx, "http_requests_total", nil)
	requestDuration, _ := m.Histogram(ctx, "http_request_duration_seconds", nil)
	activeRequests, _ := m.Gauge(ctx, "http_requests_active", nil)

	// 模拟处理请求
	activeRequests.Inc(ctx)

	// ... 处理请求 ...
	// 假设请求耗时 0.005 秒
	duration := 0.005

	requestDuration.Observe(ctx, duration)
	requestsTotal.Inc(ctx)
	activeRequests.Dec(ctx)

	// Output:
	// Gauge http_requests_active incremented to 1
	// Histogram http_request_duration_seconds observed value 0.005
	// Counter http_requests_total incremented to 1
	// Gauge http_requests_active decremented to 0
}

// Example_businessMetrics 演示如何记录业务指标
func Example_businessMetrics() {
	m := &mockMetrics{}
	ctx := context.Background()

	// 创建业务指标
	ordersTotal, _ := m.Counter(ctx, "orders_created_total", nil)
	revenue, _ := m.Counter(ctx, "revenue_total", metrics.Labels{
		"currency": "USD",
	})
	inventoryLevel, _ := m.Gauge(ctx, "inventory_level", metrics.Labels{
		"product": "widget",
	})

	// 记录订单创建
	ordersTotal.Inc(ctx)

	// 记录收入（单位：美分）
	revenue.Add(ctx, 9999) // $99.99

	// 更新库存
	inventoryLevel.Set(ctx, 150)

	// Output:
	// Counter orders_created_total incremented to 1
	// Counter revenue_total increased by 9999 to 9999
	// Gauge inventory_level set to 150
}
