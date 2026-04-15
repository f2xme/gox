package metrics

import "time"

// Collector 定义收集 HTTP 指标的接口
type Collector interface {
	// RecordRequest 记录 HTTP 请求
	RecordRequest(method, path string)

	// RecordDuration 记录 HTTP 请求的持续时间
	RecordDuration(method, path string, duration time.Duration)

	// RecordResponseSize 记录 HTTP 响应的大小
	RecordResponseSize(method, path string, size int64)

	// RecordError 记录请求处理过程中发生的错误
	RecordError(method, path string)

	// RecordCustomMetric 记录带标签的自定义业务指标
	RecordCustomMetric(name string, value float64, labels map[string]string)
}
