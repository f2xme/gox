package metrics

import "context"

// Labels 表示指标的标签集合
type Labels map[string]string

// Counter 表示一个只增不减的计数器
type Counter interface {
	// Inc 将计数器加 1
	Inc(ctx context.Context) error

	// Add 将计数器增加指定的值
	Add(ctx context.Context, delta float64) error
}

// Gauge 表示一个可增可减的仪表盘
type Gauge interface {
	// Set 设置仪表盘的值
	Set(ctx context.Context, value float64) error

	// Inc 将仪表盘的值加 1
	Inc(ctx context.Context) error

	// Dec 将仪表盘的值减 1
	Dec(ctx context.Context) error
}

// Histogram 表示一个直方图，用于记录数值分布
type Histogram interface {
	// Observe 记录一个观测值
	Observe(ctx context.Context, value float64) error
}

// Metrics 是指标工厂接口，用于创建各种类型的指标
type Metrics interface {
	// Counter 创建或获取一个计数器
	Counter(ctx context.Context, name string, labels Labels) (Counter, error)

	// Gauge 创建或获取一个仪表盘
	Gauge(ctx context.Context, name string, labels Labels) (Gauge, error)

	// Histogram 创建或获取一个直方图
	Histogram(ctx context.Context, name string, labels Labels) (Histogram, error)
}
