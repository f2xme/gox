/*
Package metrics 提供统一的指标监控抽象层，支持多种监控后端。

# 概述

metrics 包定义了一组标准接口，用于应用程序的指标收集和监控。
通过这些接口，你可以轻松地在不同的监控系统之间切换（如 Prometheus、StatsD），
而无需修改业务代码。

# 快速开始

基本用法：

	import (
		"context"
		"github.com/f2xme/gox/metrics"
		"github.com/f2xme/gox/metrics/adapter/prometheus"
	)

	func main() {
		// 创建 Prometheus 指标收集器
		m := prometheus.New()

		ctx := context.Background()

		// 创建计数器
		counter, _ := m.Counter(ctx, "http_requests_total", metrics.Labels{
			"method": "GET",
			"path":   "/api/users",
		})

		// 增加计数
		counter.Inc(ctx)
	}

# 可用适配器

## prometheus - Prometheus 监控

适用于 Prometheus 监控系统，提供丰富的指标类型和查询能力：

	import "github.com/f2xme/gox/metrics/adapter/prometheus"

	m := prometheus.New(
		prometheus.WithNamespace("myapp"),
		prometheus.WithSubsystem("api"),
	)

特性：
  - 支持 Counter、Gauge、Histogram
  - 自动注册到 Prometheus Registry
  - 支持标签（Labels）
  - HTTP /metrics 端点暴露

## statsdadapter - StatsD 监控

适用于 StatsD 协议的监控系统（如 Graphite、DataDog）：

	import "github.com/f2xme/gox/metrics/adapter/statsdadapter"

	m := statsdadapter.New(
		statsdadapter.WithAddress("localhost:8125"),
		statsdadapter.WithPrefix("myapp"),
	)

特性：
  - UDP 协议，低延迟
  - 支持采样率
  - 自动聚合
  - 无需注册中心

# 核心接口

## Metrics - 指标工厂接口

所有监控适配器都必须实现此接口：

	type Metrics interface {
		Counter(ctx context.Context, name string, labels Labels) (Counter, error)
		Gauge(ctx context.Context, name string, labels Labels) (Gauge, error)
		Histogram(ctx context.Context, name string, labels Labels) (Histogram, error)
	}

## Counter - 计数器

Counter 是一个只增不减的累加器，适用于记录请求数、错误数等：

	type Counter interface {
		Inc(ctx context.Context) error
		Add(ctx context.Context, delta float64) error
	}

使用示例：

	// 创建计数器
	counter, err := m.Counter(ctx, "http_requests_total", metrics.Labels{
		"method": "POST",
		"status": "200",
	})

	// 每次请求增加 1
	counter.Inc(ctx)

	// 批量增加
	counter.Add(ctx, 10)

典型用途：
  - HTTP 请求总数
  - 数据库查询次数
  - 错误发生次数
  - 任务完成数量

## Gauge - 仪表盘

Gauge 是一个可增可减的数值，适用于记录当前状态：

	type Gauge interface {
		Set(ctx context.Context, value float64) error
		Inc(ctx context.Context) error
		Dec(ctx context.Context) error
	}

使用示例：

	// 创建仪表盘
	gauge, err := m.Gauge(ctx, "active_connections", nil)

	// 设置当前值
	gauge.Set(ctx, 42)

	// 连接建立时增加
	gauge.Inc(ctx)

	// 连接关闭时减少
	gauge.Dec(ctx)

典型用途：
  - 当前活跃连接数
  - 内存使用量
  - 队列长度
  - CPU 使用率

## Histogram - 直方图

Histogram 用于记录数值的分布情况，适用于延迟、大小等指标：

	type Histogram interface {
		Observe(ctx context.Context, value float64) error
	}

使用示例：

	// 创建直方图
	histogram, err := m.Histogram(ctx, "http_request_duration_seconds", metrics.Labels{
		"method": "GET",
	})

	// 记录请求耗时
	start := time.Now()
	// ... 处理请求 ...
	duration := time.Since(start).Seconds()
	histogram.Observe(ctx, duration)

典型用途：
  - 请求响应时间
  - 数据库查询耗时
  - 消息大小
  - 批处理数量

# 标签（Labels）

标签用于为指标添加维度，便于分组和过滤：

	type Labels map[string]string

使用示例：

	// 带标签的计数器
	counter, _ := m.Counter(ctx, "api_requests", metrics.Labels{
		"method":   "GET",
		"endpoint": "/users",
		"status":   "200",
	})

	// 不同标签组合创建不同的指标序列
	counter2, _ := m.Counter(ctx, "api_requests", metrics.Labels{
		"method":   "POST",
		"endpoint": "/users",
		"status":   "201",
	})

标签最佳实践：
  - 使用有限的标签值（避免高基数）
  - 标签名使用小写字母和下划线
  - 避免在标签中使用用户 ID、会话 ID 等唯一值
  - 常见标签：method、status、endpoint、region

# 指标命名规范

遵循 Prometheus 命名约定：

## 基本规则

  - 使用小写字母、数字和下划线
  - 以字母开头
  - 使用描述性名称

## 命名模式

	<namespace>_<subsystem>_<name>_<unit>

示例：

	// 好的命名
	http_requests_total           // 总数使用 _total 后缀
	http_request_duration_seconds // 时间使用 _seconds 后缀
	database_connections_active   // 当前状态
	cache_hits_total              // 计数器

	// 不好的命名
	httpRequests                  // 使用驼峰命名
	requests                      // 缺少上下文
	request_time                  // 缺少单位

## 单位后缀

  - _total: 累计总数（Counter）
  - _seconds: 秒
  - _bytes: 字节
  - _ratio: 比率（0-1）
  - _percent: 百分比（0-100）

# 错误处理

包定义了标准错误：

  - ErrInvalidMetricName: 指标名称不合法

示例：

	counter, err := m.Counter(ctx, "invalid-name!", nil)
	if err == metrics.ErrInvalidMetricName {
		// 处理无效的指标名称
	}

# 最佳实践

## 1. 在应用启动时创建指标

	var (
		httpRequestsTotal *metrics.Counter
		activeConnections *metrics.Gauge
	)

	func init() {
		ctx := context.Background()
		m := prometheus.New()

		httpRequestsTotal, _ = m.Counter(ctx, "http_requests_total", nil)
		activeConnections, _ = m.Gauge(ctx, "active_connections", nil)
	}

## 2. 使用 defer 记录耗时

	func handleRequest(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			duration := time.Since(start).Seconds()
			requestDuration.Observe(context.Background(), duration)
		}()

		// 处理请求
	}

## 3. 合理使用标签

	// 推荐：有限的标签值
	counter, _ := m.Counter(ctx, "api_requests", metrics.Labels{
		"method": r.Method,        // GET, POST, PUT, DELETE
		"status": strconv.Itoa(status), // 200, 404, 500
	})

	// 不推荐：高基数标签
	counter, _ := m.Counter(ctx, "api_requests", metrics.Labels{
		"user_id": userID,  // 可能有数百万个不同值
		"request_id": reqID, // 每个请求都不同
	})

## 4. 使用 context 控制超时

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	counter.Inc(ctx)

## 5. 选择合适的指标类型

	// Counter: 只增不减
	requestsTotal.Inc(ctx)

	// Gauge: 可增可减
	activeUsers.Set(ctx, float64(len(users)))

	// Histogram: 记录分布
	requestDuration.Observe(ctx, duration)

## 6. 记录业务指标

除了系统指标，也要记录业务指标：

	// 系统指标
	http_requests_total
	memory_usage_bytes

	// 业务指标
	orders_created_total
	revenue_total
	user_signups_total

# 性能考虑

  - 指标收集的开销通常很小（微秒级）
  - Prometheus 使用 pull 模式，不会阻塞应用
  - StatsD 使用 UDP，不会因网络问题阻塞
  - 避免在热路径上创建新指标（应在启动时创建）
  - 标签组合数量会影响内存使用

# 线程安全

所有指标实现都是线程安全的，可以在多个 goroutine 中并发使用。

# 与监控系统集成

## Prometheus

	import (
		"net/http"
		"github.com/prometheus/client_golang/prometheus/promhttp"
	)

	func main() {
		m := prometheus.New()

		// 暴露 /metrics 端点
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8080", nil)
	}

## StatsD

	m := statsdadapter.New(
		statsdadapter.WithAddress("localhost:8125"),
	)

	// 指标会自动通过 UDP 发送到 StatsD

## 多后端支持

可以同时向多个监控系统发送指标：

	type multiMetrics struct {
		backends []metrics.Metrics
	}

	func (m *multiMetrics) Counter(ctx context.Context, name string, labels metrics.Labels) (metrics.Counter, error) {
		counters := make([]metrics.Counter, len(m.backends))
		for i, backend := range m.backends {
			c, err := backend.Counter(ctx, name, labels)
			if err != nil {
				return nil, err
			}
			counters[i] = c
		}
		return &multiCounter{counters: counters}, nil
	}
*/
package metrics
