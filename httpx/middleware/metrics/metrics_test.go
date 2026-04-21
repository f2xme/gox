package metrics

import (
	"testing"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/mock"
)

func TestMetrics_RecordsRequestCount(t *testing.T) {
	collector := NewMemoryCollector()
	middleware := New(WithCollector(collector))

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	ctx := mock.NewMockContext("GET", "/api/users")

	if err := handler(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	count := collector.GetRequestCount("GET", "/api/users")
	if count != 1 {
		t.Errorf("expected request count 1, got %d", count)
	}
}

func TestMetrics_SkipsConfiguredPaths(t *testing.T) {
	collector := NewMemoryCollector()
	middleware := New(
		WithCollector(collector),
		WithSkipPaths("/health", "/metrics"),
	)

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	ctx := mock.NewMockContext("GET", "/health")

	if err := handler(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	count := collector.GetRequestCount("GET", "/health")
	if count != 0 {
		t.Errorf("expected request count 0 for skipped path, got %d", count)
	}
}

func TestMetrics_NormalizesPath(t *testing.T) {
	collector := NewMemoryCollector()
	middleware := New(
		WithCollector(collector),
		WithPathNormalizer(func(path string) string {
			// Simple normalizer: replace numeric IDs with {id}
			if len(path) > 11 && path[:11] == "/api/users/" {
				return "/api/users/{id}"
			}
			return path
		}),
	)

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	ctx := mock.NewMockContext("GET", "/api/users/123")

	if err := handler(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	count := collector.GetRequestCount("GET", "/api/users/{id}")
	if count != 1 {
		t.Errorf("expected request count 1 for normalized path, got %d", count)
	}
}

func TestMetrics_RecordsDuration(t *testing.T) {
	collector := NewMemoryCollector()
	middleware := New(WithCollector(collector))

	handler := middleware(func(ctx httpx.Context) error {
		// Simulate some work
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	ctx := mock.NewMockContext("POST", "/api/orders")

	if err := handler(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	durations := collector.GetDurations("POST", "/api/orders")
	if len(durations) != 1 {
		t.Errorf("expected 1 duration record, got %d", len(durations))
	}

	if durations[0] < 0 {
		t.Errorf("expected positive duration, got %v", durations[0])
	}
}

func TestMetrics_RecordsErrors(t *testing.T) {
	collector := NewMemoryCollector()
	middleware := New(WithCollector(collector))

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(500, map[string]string{"error": "internal error"})
	})

	ctx := mock.NewMockContext("GET", "/api/fail")

	// Handler returns nil because JSON() doesn't return error in mock
	_ = handler(ctx)

	count := collector.GetRequestCount("GET", "/api/fail")
	if count != 1 {
		t.Errorf("expected request count 1, got %d", count)
	}
}

func TestMetrics_RecordsResponseSize(t *testing.T) {
	collector := NewMemoryCollector()
	middleware := New(
		WithCollector(collector),
		WithDetailedMetrics(true),
	)

	handler := middleware(func(ctx httpx.Context) error {
		ctx.SetHeader("Content-Length", "1024")
		return ctx.JSON(200, map[string]string{"data": "response"})
	})

	ctx := mock.NewMockContext("GET", "/api/data")

	if err := handler(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	sizes := collector.GetResponseSizes("GET", "/api/data")
	if len(sizes) != 1 {
		t.Errorf("expected 1 size record, got %d", len(sizes))
	}

	if sizes[0] != 1024 {
		t.Errorf("expected size 1024, got %d", sizes[0])
	}
}

func TestMetrics_CustomLabels(t *testing.T) {
	collector := NewMemoryCollector()
	middleware := New(
		WithCollector(collector),
		WithCustomLabels(func(ctx any) map[string]string {
			return map[string]string{
				"tenant": "test-tenant",
				"region": "us-west",
			}
		}),
	)

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	ctx := mock.NewMockContext("GET", "/api/data")

	if err := handler(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	// 自定义标签已记录，但当前 MemoryCollector 未暴露查询接口；
	// 本用例仅验证开启自定义标签后中间件运行不崩溃。
	count := collector.GetRequestCount("GET", "/api/data")
	if count != 1 {
		t.Errorf("expected request count 1, got %d", count)
	}
}

func TestMetrics_BusinessMetrics(t *testing.T) {
	collector := NewMemoryCollector()
	middleware := New(
		WithCollector(collector),
		WithBusinessMetrics(func(ctx any, c Collector) {
			c.RecordCustomMetric("order_value", 99.99, map[string]string{
				"currency": "USD",
			})
		}),
	)

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	ctx := mock.NewMockContext("POST", "/api/orders")

	if err := handler(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	metrics := collector.GetCustomMetrics("order_value")
	if len(metrics) != 1 {
		t.Errorf("expected 1 custom metric, got %d", len(metrics))
	}
	if len(metrics) > 0 && metrics[0] != 99.99 {
		t.Errorf("expected metric value 99.99, got %f", metrics[0])
	}
}
