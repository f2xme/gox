package metrics

import (
	"context"
	"sync"
	"testing"
)

// mockCounter 是用于测试的 Counter 实现
type mockCounter struct {
	mu    sync.Mutex
	value float64
}

func (m *mockCounter) Inc(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value++
	return nil
}

func (m *mockCounter) Add(ctx context.Context, delta float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value += delta
	return nil
}

// mockGauge 是用于测试的 Gauge 实现
type mockGauge struct {
	mu    sync.Mutex
	value float64
}

func (m *mockGauge) Set(ctx context.Context, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value = value
	return nil
}

func (m *mockGauge) Inc(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value++
	return nil
}

func (m *mockGauge) Dec(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value--
	return nil
}

// mockHistogram 是用于测试的 Histogram 实现
type mockHistogram struct {
	mu     sync.Mutex
	values []float64
}

func (m *mockHistogram) Observe(ctx context.Context, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values = append(m.values, value)
	return nil
}

// mockMetrics 是用于测试的 Metrics 实现
type mockMetrics struct {
	counters   map[string]*mockCounter
	gauges     map[string]*mockGauge
	histograms map[string]*mockHistogram
	mu         sync.Mutex
}

func newMockMetrics() *mockMetrics {
	return &mockMetrics{
		counters:   make(map[string]*mockCounter),
		gauges:     make(map[string]*mockGauge),
		histograms: make(map[string]*mockHistogram),
	}
}

func (m *mockMetrics) Counter(ctx context.Context, name string, labels Labels) (Counter, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := name
	if c, ok := m.counters[key]; ok {
		return c, nil
	}

	c := &mockCounter{}
	m.counters[key] = c
	return c, nil
}

func (m *mockMetrics) Gauge(ctx context.Context, name string, labels Labels) (Gauge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := name
	if g, ok := m.gauges[key]; ok {
		return g, nil
	}

	g := &mockGauge{}
	m.gauges[key] = g
	return g, nil
}

func (m *mockMetrics) Histogram(ctx context.Context, name string, labels Labels) (Histogram, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := name
	if h, ok := m.histograms[key]; ok {
		return h, nil
	}

	h := &mockHistogram{}
	m.histograms[key] = h
	return h, nil
}

func TestCounter(t *testing.T) {
	tests := []struct {
		name      string
		ops       func(*mockCounter) error
		wantValue float64
	}{
		{
			name: "Inc increments by 1",
			ops: func(c *mockCounter) error {
				return c.Inc(context.Background())
			},
			wantValue: 1,
		},
		{
			name: "Add increments by delta",
			ops: func(c *mockCounter) error {
				return c.Add(context.Background(), 5.5)
			},
			wantValue: 5.5,
		},
		{
			name: "Multiple Inc calls",
			ops: func(c *mockCounter) error {
				if err := c.Inc(context.Background()); err != nil {
					return err
				}
				if err := c.Inc(context.Background()); err != nil {
					return err
				}
				return c.Inc(context.Background())
			},
			wantValue: 3,
		},
		{
			name: "Inc and Add combined",
			ops: func(c *mockCounter) error {
				if err := c.Inc(context.Background()); err != nil {
					return err
				}
				return c.Add(context.Background(), 2.5)
			},
			wantValue: 3.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &mockCounter{}
			if err := tt.ops(c); err != nil {
				t.Fatalf("operation failed: %v", err)
			}
			if c.value != tt.wantValue {
				t.Errorf("got value %v, want %v", c.value, tt.wantValue)
			}
		})
	}
}

func TestGauge(t *testing.T) {
	tests := []struct {
		name      string
		ops       func(*mockGauge) error
		wantValue float64
	}{
		{
			name: "Set sets value",
			ops: func(g *mockGauge) error {
				return g.Set(context.Background(), 42.5)
			},
			wantValue: 42.5,
		},
		{
			name: "Inc increments by 1",
			ops: func(g *mockGauge) error {
				if err := g.Set(context.Background(), 10); err != nil {
					return err
				}
				return g.Inc(context.Background())
			},
			wantValue: 11,
		},
		{
			name: "Dec decrements by 1",
			ops: func(g *mockGauge) error {
				if err := g.Set(context.Background(), 10); err != nil {
					return err
				}
				return g.Dec(context.Background())
			},
			wantValue: 9,
		},
		{
			name: "Inc and Dec combined",
			ops: func(g *mockGauge) error {
				if err := g.Set(context.Background(), 5); err != nil {
					return err
				}
				if err := g.Inc(context.Background()); err != nil {
					return err
				}
				if err := g.Inc(context.Background()); err != nil {
					return err
				}
				return g.Dec(context.Background())
			},
			wantValue: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &mockGauge{}
			if err := tt.ops(g); err != nil {
				t.Fatalf("operation failed: %v", err)
			}
			if g.value != tt.wantValue {
				t.Errorf("got value %v, want %v", g.value, tt.wantValue)
			}
		})
	}
}

func TestHistogram(t *testing.T) {
	tests := []struct {
		name       string
		ops        func(*mockHistogram) error
		wantValues []float64
	}{
		{
			name: "Single observation",
			ops: func(h *mockHistogram) error {
				return h.Observe(context.Background(), 1.5)
			},
			wantValues: []float64{1.5},
		},
		{
			name: "Multiple observations",
			ops: func(h *mockHistogram) error {
				if err := h.Observe(context.Background(), 1.0); err != nil {
					return err
				}
				if err := h.Observe(context.Background(), 2.5); err != nil {
					return err
				}
				return h.Observe(context.Background(), 3.7)
			},
			wantValues: []float64{1.0, 2.5, 3.7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &mockHistogram{}
			if err := tt.ops(h); err != nil {
				t.Fatalf("operation failed: %v", err)
			}
			if len(h.values) != len(tt.wantValues) {
				t.Fatalf("got %d values, want %d", len(h.values), len(tt.wantValues))
			}
			for i, v := range h.values {
				if v != tt.wantValues[i] {
					t.Errorf("value[%d] = %v, want %v", i, v, tt.wantValues[i])
				}
			}
		})
	}
}

func TestMetrics(t *testing.T) {
	ctx := context.Background()
	m := newMockMetrics()

	t.Run("Counter creation", func(t *testing.T) {
		c, err := m.Counter(ctx, "test_counter", nil)
		if err != nil {
			t.Fatalf("Counter() error = %v", err)
		}
		if c == nil {
			t.Fatal("Counter() returned nil")
		}
	})

	t.Run("Gauge creation", func(t *testing.T) {
		g, err := m.Gauge(ctx, "test_gauge", nil)
		if err != nil {
			t.Fatalf("Gauge() error = %v", err)
		}
		if g == nil {
			t.Fatal("Gauge() returned nil")
		}
	})

	t.Run("Histogram creation", func(t *testing.T) {
		h, err := m.Histogram(ctx, "test_histogram", nil)
		if err != nil {
			t.Fatalf("Histogram() error = %v", err)
		}
		if h == nil {
			t.Fatal("Histogram() returned nil")
		}
	})

	t.Run("Counter with labels", func(t *testing.T) {
		labels := Labels{"method": "GET", "path": "/api"}
		c, err := m.Counter(ctx, "http_requests", labels)
		if err != nil {
			t.Fatalf("Counter() error = %v", err)
		}
		if c == nil {
			t.Fatal("Counter() returned nil")
		}
	})
}
