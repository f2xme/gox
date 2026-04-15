package metrics

import (
	"net/http"
	"testing"

	"github.com/f2xme/gox/httpx"
)

// mockContext implements httpx.Context for testing
type mockContext struct {
	method      string
	path        string
	respCode    int
	respBody    any
	store       map[string]any
	headers     map[string]string
	queries     map[string]string
	respHeaders map[string]string
	respWriter  *mockResponseWriter
}

func (m *mockContext) Request() *http.Request                   { return nil }
func (m *mockContext) Param(string) string                      { return "" }
func (m *mockContext) Query(key string) string                  { return m.queries[key] }
func (m *mockContext) QueryDefault(key, def string) string {
	if v, ok := m.queries[key]; ok {
		return v
	}
	return def
}
func (m *mockContext) Header(key string) string                 { return m.headers[key] }
func (m *mockContext) Cookie(string) (*http.Cookie, error)      { return nil, http.ErrNoCookie }
func (m *mockContext) ClientIP() string                         { return "127.0.0.1" }
func (m *mockContext) Method() string                           { return m.method }
func (m *mockContext) Path() string                             { return m.path }
func (m *mockContext) Bind(any) error                           { return nil }
func (m *mockContext) BindJSON(any) error                       { return nil }
func (m *mockContext) BindQuery(any) error                      { return nil }
func (m *mockContext) BindForm(any) error                       { return nil }
func (m *mockContext) JSON(code int, v any) error               { m.respCode = code; m.respBody = v; return nil }
func (m *mockContext) String(code int, s string) error          { m.respCode = code; m.respBody = s; return nil }
func (m *mockContext) HTML(code int, _ string) error            { m.respCode = code; return nil }
func (m *mockContext) Blob(code int, _ string, _ []byte) error  { m.respCode = code; return nil }
func (m *mockContext) NoContent(code int) error                 { m.respCode = code; return nil }
func (m *mockContext) Redirect(code int, _ string) error        { m.respCode = code; return nil }
func (m *mockContext) SetHeader(key, value string) {
	m.respHeaders[key] = value
	if m.respWriter != nil {
		m.respWriter.headers[key] = value
	}
}
func (m *mockContext) SetCookie(*http.Cookie)                   {}
func (m *mockContext) Status(code int)                          { m.respCode = code }
func (m *mockContext) Success(data any) error                   { return m.JSON(200, data) }
func (m *mockContext) Fail(msg string) error                    { return m.JSON(200, msg) }
func (m *mockContext) Set(key string, value any)                { m.store[key] = value }
func (m *mockContext) Get(key string) (any, bool)               { v, ok := m.store[key]; return v, ok }
func (m *mockContext) MustGet(key string) any                   { return m.store[key] }
func (m *mockContext) ResponseWriter() http.ResponseWriter {
	if m.respWriter == nil {
		m.respWriter = &mockResponseWriter{headers: m.respHeaders}
	}
	return m.respWriter
}
func (m *mockContext) Raw() any { return nil }

type mockResponseWriter struct {
	headers map[string]string
}

func (m *mockResponseWriter) Header() http.Header {
	h := http.Header{}
	for k, v := range m.headers {
		h.Set(k, v)
	}
	return h
}

func (m *mockResponseWriter) Write([]byte) (int, error) { return 0, nil }
func (m *mockResponseWriter) WriteHeader(int)           {}

func TestMetrics_RecordsRequestCount(t *testing.T) {
	collector := NewMemoryCollector()
	middleware := New(WithCollector(collector))

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	ctx := &mockContext{
		method:      "GET",
		path:        "/api/users",
		store:       make(map[string]any),
		headers:     make(map[string]string),
		queries:     make(map[string]string),
		respHeaders: make(map[string]string),
	}

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

	ctx := &mockContext{
		method:      "GET",
		path:        "/health",
		store:       make(map[string]any),
		headers:     make(map[string]string),
		queries:     make(map[string]string),
		respHeaders: make(map[string]string),
	}

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

	ctx := &mockContext{
		method:      "GET",
		path:        "/api/users/123",
		store:       make(map[string]any),
		headers:     make(map[string]string),
		queries:     make(map[string]string),
		respHeaders: make(map[string]string),
	}

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

	ctx := &mockContext{
		method:      "POST",
		path:        "/api/orders",
		store:       make(map[string]any),
		headers:     make(map[string]string),
		queries:     make(map[string]string),
		respHeaders: make(map[string]string),
	}

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

	ctx := &mockContext{
		method:      "GET",
		path:        "/api/fail",
		store:       make(map[string]any),
		headers:     make(map[string]string),
		queries:     make(map[string]string),
		respHeaders: make(map[string]string),
	}

	// Handler returns nil because JSON() doesn't return error in mock
	_ = handler(ctx)

	// Check that request was recorded
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

	ctx := &mockContext{
		method:      "GET",
		path:        "/api/data",
		store:       make(map[string]any),
		headers:     make(map[string]string),
		queries:     make(map[string]string),
		respHeaders: make(map[string]string),
	}

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

	ctx := &mockContext{
		method:      "GET",
		path:        "/api/data",
		store:       make(map[string]any),
		headers:     make(map[string]string),
		queries:     make(map[string]string),
		respHeaders: make(map[string]string),
	}

	if err := handler(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	// Custom labels are recorded but not directly testable with current MemoryCollector
	// This test verifies the middleware doesn't crash with custom labels configured
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
			// Record custom business metric
			c.RecordCustomMetric("order_value", 99.99, map[string]string{
				"currency": "USD",
			})
		}),
	)

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	ctx := &mockContext{
		method:      "POST",
		path:        "/api/orders",
		store:       make(map[string]any),
		headers:     make(map[string]string),
		queries:     make(map[string]string),
		respHeaders: make(map[string]string),
	}

	if err := handler(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	// Verify custom metric was recorded
	metrics := collector.GetCustomMetrics("order_value")
	if len(metrics) != 1 {
		t.Errorf("expected 1 custom metric, got %d", len(metrics))
	}
	if len(metrics) > 0 && metrics[0] != 99.99 {
		t.Errorf("expected metric value 99.99, got %f", metrics[0])
	}
}
