/*
Package prometheus 提供 Prometheus 监控系统的 metrics 适配器实现。

# 概述

prometheus 实现了 metrics.Metrics 接口，将指标收集请求转换为 Prometheus 格式。
它使用 prometheus/client_golang 库，支持 Counter、Gauge、Histogram 三种指标类型。

# 快速开始

基本用法：

	import (
		"context"
		"net/http"
		"github.com/f2xme/gox/metrics"
		"github.com/f2xme/gox/metrics/adapter/prometheus"
	)

	func main() {
		// 创建 Prometheus adapter
		adapter := prometheus.New()

		ctx := context.Background()

		// 创建计数器
		counter, _ := adapter.Counter(ctx, "http_requests_total", metrics.Labels{
			"method": "GET",
			"path":   "/api/users",
		})

		// 增加计数
		counter.Inc(ctx)

		// 暴露 /metrics 端点
		http.Handle("/metrics", adapter.Handler())
		http.ListenAndServe(":8080", nil)
	}

# 配置选项

使用 Option 模式配置 adapter：

	adapter := prometheus.New(
		prometheus.WithNamespace("myapp"),
		prometheus.WithSubsystem("api"),
		prometheus.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 1, 10}),
	)

可用选项：
  - WithNamespace: 设置指标名称前缀（如 "myapp"）
  - WithSubsystem: 设置子系统名称（如 "api"）
  - WithHistogramBuckets: 自定义 Histogram 的 buckets

# 指标类型

## Counter - 计数器

Counter 是只增不减的累加器：

	counter, _ := adapter.Counter(ctx, "requests_total", metrics.Labels{
		"method": "POST",
	})

	counter.Inc(ctx)        // 增加 1
	counter.Add(ctx, 10)    // 增加 10

## Gauge - 仪表盘

Gauge 是可增可减的数值：

	gauge, _ := adapter.Gauge(ctx, "active_connections", nil)

	gauge.Set(ctx, 42)  // 设置为 42
	gauge.Inc(ctx)      // 增加 1
	gauge.Dec(ctx)      // 减少 1

## Histogram - 直方图

Histogram 记录数值分布：

	histogram, _ := adapter.Histogram(ctx, "request_duration_seconds", metrics.Labels{
		"method": "GET",
	})

	histogram.Observe(ctx, 0.123)  // 记录一次观测值

# HTTP 端点集成

使用 Handler() 方法获取 HTTP handler：

	import (
		"net/http"
		"github.com/f2xme/gox/metrics/adapter/prometheus"
	)

	func main() {
		adapter := prometheus.New()

		// 注册 /metrics 端点
		http.Handle("/metrics", adapter.Handler())

		// 启动 HTTP 服务器
		http.ListenAndServe(":8080", nil)
	}

访问 http://localhost:8080/metrics 查看指标。

# 指标命名

Prometheus 会自动组合 namespace、subsystem 和 name：

	adapter := prometheus.New(
		prometheus.WithNamespace("myapp"),
		prometheus.WithSubsystem("api"),
	)

	counter, _ := adapter.Counter(ctx, "requests_total", nil)
	// 实际指标名称: myapp_api_requests_total

# 标签处理

标签用于为指标添加维度：

	counter, _ := adapter.Counter(ctx, "requests_total", metrics.Labels{
		"method": "GET",
		"status": "200",
	})

	// 相同名称和标签键的指标会复用同一个 CounterVec
	counter2, _ := adapter.Counter(ctx, "requests_total", metrics.Labels{
		"method": "POST",
		"status": "201",
	})

注意：
  - 相同名称和标签键的指标共享同一个 Vec
  - 不同的标签值会创建不同的时间序列
  - 避免使用高基数的标签值（如用户 ID）

# 线程安全

所有方法都是线程安全的，可以在多个 goroutine 中并发使用。

# 性能考虑

  - 指标创建时会缓存，相同名称和标签键的指标只创建一次
  - 使用 sync.Map 优化并发读取性能
  - Prometheus 使用 pull 模式，不会阻塞应用
  - 指标收集的开销通常在微秒级

# 最佳实践

1. 在应用启动时创建指标，避免在热路径上创建
2. 使用有限的标签值，避免高基数
3. 遵循 Prometheus 命名规范（小写字母、下划线、单位后缀）
4. 为 Counter 使用 _total 后缀
5. 为时间指标使用 _seconds 后缀
6. 设置合适的 namespace 和 subsystem

# 示例

完整的 HTTP 服务器示例：

	package main

	import (
		"context"
		"net/http"
		"time"

		"github.com/f2xme/gox/metrics"
		"github.com/f2xme/gox/metrics/adapter/prometheus"
	)

	var (
		requestsTotal    metrics.Counter
		requestDuration  metrics.Histogram
		activeConnections metrics.Gauge
	)

	func init() {
		adapter := prometheus.New(
			prometheus.WithNamespace("myapp"),
			prometheus.WithSubsystem("http"),
		)

		ctx := context.Background()
		requestsTotal, _ = adapter.Counter(ctx, "requests_total", metrics.Labels{
			"method": "GET",
		})
		requestDuration, _ = adapter.Histogram(ctx, "request_duration_seconds", nil)
		activeConnections, _ = adapter.Gauge(ctx, "active_connections", nil)

		// 暴露 /metrics 端点
		http.Handle("/metrics", adapter.Handler())
	}

	func handleRequest(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		activeConnections.Inc(context.Background())
		defer func() {
			activeConnections.Dec(context.Background())
			duration := time.Since(start).Seconds()
			requestDuration.Observe(context.Background(), duration)
			requestsTotal.Inc(context.Background())
		}()

		// 处理请求
		w.Write([]byte("Hello, World!"))
	}

	func main() {
		http.HandleFunc("/", handleRequest)
		http.ListenAndServe(":8080", nil)
	}
*/
package prometheus
