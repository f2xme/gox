package prometheus

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// prometheusCounter 实现 metrics.Counter 接口
type prometheusCounter struct {
	vec    *prometheus.CounterVec
	labels []string
}

// Inc 增加计数器的值（增加 1）
func (c *prometheusCounter) Inc(ctx context.Context) error {
	c.vec.WithLabelValues(c.labels...).Inc()
	return nil
}

// Add 增加计数器的值（增加 delta）
func (c *prometheusCounter) Add(ctx context.Context, delta float64) error {
	c.vec.WithLabelValues(c.labels...).Add(delta)
	return nil
}
