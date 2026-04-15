package tracing

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/f2xme/gox/httpx"
	ginadapter "github.com/f2xme/gox/httpx/adapter/gin"
	"github.com/gin-gonic/gin"
)

// mockTracer is a mock tracer for testing.
type mockTracer struct {
	spans []*mockSpan
}

func newMockTracer() *mockTracer {
	return &mockTracer{
		spans: make([]*mockSpan, 0),
	}
}

func (t *mockTracer) Extract(ctx httpx.Context) context.Context {
	return ctx.Request().Context()
}

func (t *mockTracer) StartSpan(ctx context.Context, operationName string) Span {
	span := &mockSpan{
		operationName: operationName,
		tags:          make(map[string]any),
		baggage:       make(map[string]string),
		finished:      false,
	}
	t.spans = append(t.spans, span)
	return span
}

func (t *mockTracer) Inject(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, "span", span)
}

// mockSpan is a mock span for testing.
type mockSpan struct {
	operationName string
	tags          map[string]any
	baggage       map[string]string
	finished      bool
}

func (s *mockSpan) SetTag(key string, value any) {
	s.tags[key] = value
}

func (s *mockSpan) SetBaggageItem(key, value string) {
	s.baggage[key] = value
}

func (s *mockSpan) Finish() {
	s.finished = true
}

func (s *mockSpan) Context() context.Context {
	return context.Background()
}

func TestTracing_NoTracer(t *testing.T) {
	engine := ginadapter.New()
	engine.Use(New()) // No tracer provided

	engine.GET("/test", func(c httpx.Context) error {
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestTracing_WithTracer(t *testing.T) {
	tracer := newMockTracer()
	engine := ginadapter.New()
	engine.Use(New(WithTracer(tracer)))

	engine.GET("/api/users", func(c httpx.Context) error {
		return c.JSON(200, map[string]any{"users": []string{}})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if len(tracer.spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(tracer.spans))
	}

	span := tracer.spans[0]

	if span.operationName != "GET /api/users" {
		t.Errorf("expected operation name 'GET /api/users', got '%s'", span.operationName)
	}

	if !span.finished {
		t.Error("expected span to be finished")
	}

	if span.tags["http.method"] != "GET" {
		t.Errorf("expected http.method tag to be 'GET', got '%v'", span.tags["http.method"])
	}

	if span.tags["http.url"] != "/api/users" {
		t.Errorf("expected http.url tag to be '/api/users', got '%v'", span.tags["http.url"])
	}

	if _, ok := span.tags["http.duration_ms"]; !ok {
		t.Error("expected http.duration_ms tag to be set")
	}
}

func TestTracing_WithError(t *testing.T) {
	tracer := newMockTracer()
	engine := ginadapter.New()
	engine.Use(New(WithTracer(tracer)))

	expectedErr := errors.New("test error")

	engine.GET("/error", func(c httpx.Context) error {
		return expectedErr
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if len(tracer.spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(tracer.spans))
	}

	span := tracer.spans[0]

	if span.tags["error"] != true {
		t.Error("expected error tag to be true")
	}

	if span.tags["error.message"] != "test error" {
		t.Errorf("expected error.message tag to be 'test error', got '%v'", span.tags["error.message"])
	}
}

func TestTracing_CustomOperationName(t *testing.T) {
	tracer := newMockTracer()
	engine := ginadapter.New()
	engine.Use(New(
		WithTracer(tracer),
		WithOperationName(func(ctx httpx.Context) string {
			return "custom-" + ctx.Path()
		}),
	))

	engine.GET("/test", func(c httpx.Context) error {
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if len(tracer.spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(tracer.spans))
	}

	span := tracer.spans[0]

	if span.operationName != "custom-/test" {
		t.Errorf("expected operation name 'custom-/test', got '%s'", span.operationName)
	}
}

func TestTracing_CustomHandler(t *testing.T) {
	tracer := newMockTracer()
	handlerCalled := false

	engine := ginadapter.New()
	engine.Use(New(
		WithTracer(tracer),
		WithHandler(func(ctx httpx.Context, span Span) {
			handlerCalled = true
			span.SetTag("custom.tag", "custom value")
		}),
	))

	engine.GET("/test", func(c httpx.Context) error {
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("expected custom handler to be called")
	}

	if len(tracer.spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(tracer.spans))
	}

	span := tracer.spans[0]

	if span.tags["custom.tag"] != "custom value" {
		t.Errorf("expected custom.tag to be 'custom value', got '%v'", span.tags["custom.tag"])
	}
}

func TestTracing_GetSpan(t *testing.T) {
	tracer := newMockTracer()
	engine := ginadapter.New()
	engine.Use(New(WithTracer(tracer)))

	var retrievedSpan Span

	engine.GET("/test", func(c httpx.Context) error {
		retrievedSpan = GetSpan(c)
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if retrievedSpan == nil {
		t.Error("expected to retrieve span from context")
	}

	if retrievedSpan != tracer.spans[0] {
		t.Error("retrieved span does not match created span")
	}
}

func TestTracing_GetSpan_NoSpan(t *testing.T) {
	engine := ginadapter.New()
	// No tracing middleware

	var retrievedSpan Span

	engine.GET("/test", func(c httpx.Context) error {
		retrievedSpan = GetSpan(c)
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if retrievedSpan != nil {
		t.Error("expected nil span when no tracing middleware is used")
	}
}
