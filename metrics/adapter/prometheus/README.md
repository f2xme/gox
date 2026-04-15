# Prometheus Adapter

Prometheus adapter 为 `metrics` 包提供 Prometheus 监控系统的实现。

## 安装

```bash
go get github.com/f2xme/gox/metrics/adapter/prometheus
```

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/f2xme/gox/metrics"
    prometheus "github.com/f2xme/gox/metrics/adapter/prometheus"
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
```

访问 `http://localhost:8080/metrics` 查看指标。

### 配置选项

使用 Option 模式配置 adapter：

```go
adapter := prometheus.New(
    prometheus.WithNamespace("myapp"),
    prometheus.WithSubsystem("api"),
    prometheus.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 1, 10}),
)
```

#### 可用选项

- **WithNamespace(string)**: 设置指标名称前缀
- **WithSubsystem(string)**: 设置子系统名称
- **WithHistogramBuckets([]float64)**: 自定义 Histogram 的 buckets

## 指标类型

### Counter - 计数器

Counter 是只增不减的累加器，适用于记录请求数、错误数等：

```go
counter, _ := adapter.Counter(ctx, "requests_total", metrics.Labels{
    "method": "POST",
    "status": "200",
})

counter.Inc(ctx)        // 增加 1
counter.Add(ctx, 10)    // 增加 10
```

### Gauge - 仪表盘

Gauge 是可增可减的数值，适用于记录当前状态：

```go
gauge, _ := adapter.Gauge(ctx, "active_connections", nil)

gauge.Set(ctx, 42)  // 设置为 42
gauge.Inc(ctx)      // 增加 1
gauge.Dec(ctx)      // 减少 1
```

### Histogram - 直方图

Histogram 记录数值分布，适用于延迟、大小等指标：

```go
histogram, _ := adapter.Histogram(ctx, "request_duration_seconds", metrics.Labels{
    "method": "GET",
})

histogram.Observe(ctx, 0.123)  // 记录一次观测值
```

## 完整示例

### HTTP 服务器集成

```go
package main

import (
    "context"
    "net/http"
    "time"
    
    "github.com/f2xme/gox/metrics"
    prometheus "github.com/f2xme/gox/metrics/adapter/prometheus"
)

var (
    requestsTotal     metrics.Counter
    requestDuration   metrics.Histogram
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
```

### 自定义 Histogram Buckets

```go
// 为 API 延迟优化的 buckets（毫秒级）
adapter := prometheus.New(
    prometheus.WithHistogramBuckets([]float64{
        0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5,
    }),
)

histogram, _ := adapter.Histogram(ctx, "api_latency_seconds", metrics.Labels{
    "endpoint": "/api/users",
})
```

## 指标命名规范

遵循 Prometheus 命名约定：

### 基本规则

- 使用小写字母、数字和下划线
- 以字母开头
- 使用描述性名称

### 命名模式

```
<namespace>_<subsystem>_<name>_<unit>
```

### 示例

```go
// 好的命名
http_requests_total           // 总数使用 _total 后缀
http_request_duration_seconds // 时间使用 _seconds 后缀
database_connections_active   // 当前状态
cache_hits_total              // 计数器

// 不好的命名
httpRequests                  // 使用驼峰命名
requests                      // 缺少上下文
request_time                  // 缺少单位
```

### 单位后缀

- `_total`: 累计总数（Counter）
- `_seconds`: 秒
- `_bytes`: 字节
- `_ratio`: 比率（0-1）
- `_percent`: 百分比（0-100）

## 标签使用

### 推荐做法

```go
// 使用有限的标签值
counter, _ := adapter.Counter(ctx, "api_requests", metrics.Labels{
    "method": r.Method,              // GET, POST, PUT, DELETE
    "status": strconv.Itoa(status),  // 200, 404, 500
    "endpoint": "/api/users",        // 有限的端点数量
})
```

### 避免做法

```go
// 不要使用高基数标签
counter, _ := adapter.Counter(ctx, "api_requests", metrics.Labels{
    "user_id": userID,      // 可能有数百万个不同值
    "request_id": reqID,    // 每个请求都不同
    "timestamp": now,       // 每次都不同
})
```

### 标签最佳实践

- 使用有限的标签值（避免高基数）
- 标签名使用小写字母和下划线
- 避免在标签中使用用户 ID、会话 ID 等唯一值
- 常见标签：method、status、endpoint、region

## Prometheus 集成

### 配置 Prometheus

在 `prometheus.yml` 中添加抓取配置：

```yaml
scrape_configs:
  - job_name: 'myapp'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### 查询示例

```promql
# 每秒请求数
rate(myapp_http_requests_total[5m])

# 按状态码分组的请求数
sum by (status) (myapp_http_requests_total)

# P95 延迟
histogram_quantile(0.95, rate(myapp_http_request_duration_seconds_bucket[5m]))

# 当前活跃连接数
myapp_http_active_connections
```

## 性能考虑

- 指标收集的开销通常很小（微秒级）
- Prometheus 使用 pull 模式，不会阻塞应用
- 避免在热路径上创建新指标（应在启动时创建）
- 标签组合数量会影响内存使用
- 使用 `sync.Map` 缓存指标，避免重复注册

## 线程安全

所有指标实现都是线程安全的，可以在多个 goroutine 中并发使用。

## 测试

运行测试：

```bash
go test -v github.com/f2xme/gox/metrics/adapter/prometheus
```

## 相关链接

- [Prometheus 官方文档](https://prometheus.io/docs/)
- [Prometheus Go Client](https://github.com/prometheus/client_golang)
- [Prometheus 最佳实践](https://prometheus.io/docs/practices/naming/)
- [metrics 包文档](../../README.md)

## License

MIT
