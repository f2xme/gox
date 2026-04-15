package prometheus

import (
	"context"
	"testing"

	"github.com/f2xme/gox/metrics"
)

func TestNew(t *testing.T) {
	a := New()
	if a == nil {
		t.Fatal("New() returned nil")
	}
	if a.registry == nil {
		t.Error("registry is nil")
	}
}

func TestWithNamespace(t *testing.T) {
	a := New(WithNamespace("test"))
	if a.namespace != "test" {
		t.Errorf("expected namespace 'test', got '%s'", a.namespace)
	}
}

func TestWithSubsystem(t *testing.T) {
	a := New(WithSubsystem("api"))
	if a.subsystem != "api" {
		t.Errorf("expected subsystem 'api', got '%s'", a.subsystem)
	}
}

func TestWithHistogramBuckets(t *testing.T) {
	buckets := []float64{0.001, 0.01, 0.1, 1, 10}
	a := New(WithHistogramBuckets(buckets))
	if len(a.buckets) != len(buckets) {
		t.Errorf("expected %d buckets, got %d", len(buckets), len(a.buckets))
	}
}

func TestCounter(t *testing.T) {
	a := New(WithNamespace("test"))
	ctx := context.Background()

	counter, err := a.Counter(ctx, "requests_total", metrics.Labels{"method": "GET"})
	if err != nil {
		t.Fatalf("Counter() error: %v", err)
	}
	if counter == nil {
		t.Fatal("Counter() returned nil")
	}

	// Test Inc
	if err := counter.Inc(ctx); err != nil {
		t.Errorf("Inc() error: %v", err)
	}

	// Test Add
	if err := counter.Add(ctx, 5.0); err != nil {
		t.Errorf("Add() error: %v", err)
	}

	// Test different labels create different counters
	counter2, err := a.Counter(ctx, "requests_total", metrics.Labels{"method": "POST"})
	if err != nil {
		t.Fatalf("Counter() error: %v", err)
	}
	if counter2 == nil {
		t.Fatal("Counter() returned nil for different labels")
	}
}

func TestCounterWithoutLabels(t *testing.T) {
	a := New()
	ctx := context.Background()

	counter, err := a.Counter(ctx, "total_requests", nil)
	if err != nil {
		t.Fatalf("Counter() error: %v", err)
	}

	if err := counter.Inc(ctx); err != nil {
		t.Errorf("Inc() error: %v", err)
	}
}

func TestGauge(t *testing.T) {
	a := New(WithNamespace("test"))
	ctx := context.Background()

	gauge, err := a.Gauge(ctx, "active_connections", metrics.Labels{"type": "http"})
	if err != nil {
		t.Fatalf("Gauge() error: %v", err)
	}
	if gauge == nil {
		t.Fatal("Gauge() returned nil")
	}

	// Test Set
	if err := gauge.Set(ctx, 10.0); err != nil {
		t.Errorf("Set() error: %v", err)
	}

	// Test Inc
	if err := gauge.Inc(ctx); err != nil {
		t.Errorf("Inc() error: %v", err)
	}

	// Test Dec
	if err := gauge.Dec(ctx); err != nil {
		t.Errorf("Dec() error: %v", err)
	}
}

func TestGaugeWithoutLabels(t *testing.T) {
	a := New()
	ctx := context.Background()

	gauge, err := a.Gauge(ctx, "memory_usage", nil)
	if err != nil {
		t.Fatalf("Gauge() error: %v", err)
	}

	if err := gauge.Set(ctx, 100.0); err != nil {
		t.Errorf("Set() error: %v", err)
	}
}

func TestHistogram(t *testing.T) {
	a := New(WithNamespace("test"))
	ctx := context.Background()

	histogram, err := a.Histogram(ctx, "request_duration_seconds", metrics.Labels{"endpoint": "/api"})
	if err != nil {
		t.Fatalf("Histogram() error: %v", err)
	}
	if histogram == nil {
		t.Fatal("Histogram() returned nil")
	}

	// Test Observe
	if err := histogram.Observe(ctx, 0.5); err != nil {
		t.Errorf("Observe() error: %v", err)
	}
	if err := histogram.Observe(ctx, 1.2); err != nil {
		t.Errorf("Observe() error: %v", err)
	}
}

func TestHistogramWithCustomBuckets(t *testing.T) {
	buckets := []float64{0.001, 0.01, 0.1, 1, 10}
	a := New(WithHistogramBuckets(buckets))
	ctx := context.Background()

	histogram, err := a.Histogram(ctx, "latency", nil)
	if err != nil {
		t.Fatalf("Histogram() error: %v", err)
	}

	if err := histogram.Observe(ctx, 0.05); err != nil {
		t.Errorf("Observe() error: %v", err)
	}
}

func TestHandler(t *testing.T) {
	a := New()
	handler := a.Handler()
	if handler == nil {
		t.Error("Handler() returned nil")
	}
}

func TestMultipleMetrics(t *testing.T) {
	a := New(WithNamespace("app"), WithSubsystem("http"))
	ctx := context.Background()

	// Create multiple counters
	c1, _ := a.Counter(ctx, "requests_total", metrics.Labels{"method": "GET"})
	c2, _ := a.Counter(ctx, "requests_total", metrics.Labels{"method": "POST"})
	c3, _ := a.Counter(ctx, "errors_total", metrics.Labels{"code": "500"})

	c1.Inc(ctx)
	c2.Add(ctx, 2)
	c3.Inc(ctx)

	// Create multiple gauges
	g1, _ := a.Gauge(ctx, "connections", metrics.Labels{"type": "active"})
	g2, _ := a.Gauge(ctx, "connections", metrics.Labels{"type": "idle"})

	g1.Set(ctx, 10)
	g2.Set(ctx, 5)

	// Create multiple histograms
	h1, _ := a.Histogram(ctx, "latency_seconds", metrics.Labels{"endpoint": "/api"})
	h2, _ := a.Histogram(ctx, "latency_seconds", metrics.Labels{"endpoint": "/health"})

	h1.Observe(ctx, 0.5)
	h2.Observe(ctx, 0.1)
}

func TestMetricCaching(t *testing.T) {
	a := New()
	ctx := context.Background()

	// Create counter with same name and labels twice
	c1, _ := a.Counter(ctx, "test_counter", metrics.Labels{"label": "value"})
	c2, _ := a.Counter(ctx, "test_counter", metrics.Labels{"label": "value"})

	// Both should work without error (cached)
	c1.Inc(ctx)
	c2.Inc(ctx)

	// Create gauge with same name and labels twice
	g1, _ := a.Gauge(ctx, "test_gauge", metrics.Labels{"label": "value"})
	g2, _ := a.Gauge(ctx, "test_gauge", metrics.Labels{"label": "value"})

	g1.Set(ctx, 10)
	g2.Set(ctx, 20)

	// Create histogram with same name and labels twice
	h1, _ := a.Histogram(ctx, "test_histogram", metrics.Labels{"label": "value"})
	h2, _ := a.Histogram(ctx, "test_histogram", metrics.Labels{"label": "value"})

	h1.Observe(ctx, 0.5)
	h2.Observe(ctx, 1.0)
}
