package metrics

import (
	"sync"
	"time"
)

// MemoryCollector 是用于测试的简单内存指标收集器
type MemoryCollector struct {
	mu       sync.RWMutex
	requests map[string]int
	durations map[string][]time.Duration
	sizes    map[string][]int64
	errors   map[string]int
	custom   map[string][]float64
}

// NewMemoryCollector 创建新的内存指标收集器
func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{
		requests:  make(map[string]int),
		durations: make(map[string][]time.Duration),
		sizes:     make(map[string][]int64),
		errors:    make(map[string]int),
		custom:    make(map[string][]float64),
	}
}

// RecordRequest 记录 HTTP 请求
func (m *MemoryCollector) RecordRequest(method, path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.makeKey(method, path)
	m.requests[key]++
}

// RecordDuration 记录 HTTP 请求的持续时间
func (m *MemoryCollector) RecordDuration(method, path string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.makeKey(method, path)
	m.durations[key] = append(m.durations[key], duration)
}

// RecordResponseSize 记录 HTTP 响应的大小
func (m *MemoryCollector) RecordResponseSize(method, path string, size int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.makeKey(method, path)
	m.sizes[key] = append(m.sizes[key], size)
}

// RecordError 记录请求处理过程中发生的错误
func (m *MemoryCollector) RecordError(method, path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.makeKey(method, path)
	m.errors[key]++
}

// RecordCustomMetric 记录带标签的自定义指标
func (m *MemoryCollector) RecordCustomMetric(name string, value float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.custom[name] = append(m.custom[name], value)
}

// GetRequestCount 返回给定方法和路径的请求数
func (m *MemoryCollector) GetRequestCount(method, path string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := m.makeKey(method, path)
	return m.requests[key]
}

// GetDurations 返回给定方法和路径的所有记录持续时间
func (m *MemoryCollector) GetDurations(method, path string) []time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := m.makeKey(method, path)
	return m.durations[key]
}

// GetResponseSizes 返回给定方法和路径的所有记录响应大小
func (m *MemoryCollector) GetResponseSizes(method, path string) []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := m.makeKey(method, path)
	return m.sizes[key]
}

// GetErrorCount 返回给定方法和路径的错误数
func (m *MemoryCollector) GetErrorCount(method, path string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := m.makeKey(method, path)
	return m.errors[key]
}

// GetCustomMetrics 返回自定义指标的所有记录值
func (m *MemoryCollector) GetCustomMetrics(name string) []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.custom[name]
}

func (m *MemoryCollector) makeKey(method, path string) string {
	return method + ":" + path
}
