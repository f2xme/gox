package prometheus

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// prometheusGauge 实现 metrics.Gauge 接口
type prometheusGauge struct {
	vec    *prometheus.GaugeVec
	labels []string
}

// Set 设置 Gauge 的值
func (g *prometheusGauge) Set(ctx context.Context, value float64) error {
	g.vec.WithLabelValues(g.labels...).Set(value)
	return nil
}

// Inc 增加 Gauge 的值（增加 1）
func (g *prometheusGauge) Inc(ctx context.Context) error {
	g.vec.WithLabelValues(g.labels...).Inc()
	return nil
}

// Dec 减少 Gauge 的值（减少 1）
func (g *prometheusGauge) Dec(ctx context.Context) error {
	g.vec.WithLabelValues(g.labels...).Dec()
	return nil
}
