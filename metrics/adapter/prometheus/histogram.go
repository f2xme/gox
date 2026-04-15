package prometheus

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// prometheusHistogram 实现 metrics.Histogram 接口
type prometheusHistogram struct {
	vec    *prometheus.HistogramVec
	labels []string
}

// Observe 记录一个观测值
func (h *prometheusHistogram) Observe(ctx context.Context, value float64) error {
	h.vec.WithLabelValues(h.labels...).Observe(value)
	return nil
}
