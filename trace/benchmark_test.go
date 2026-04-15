package trace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// BenchmarkToContext 测试 ToContext 性能
func BenchmarkToContext(b *testing.B) {
	info := &Info{
		TraceID:   "trace-123",
		SpanID:    "span-456",
		DeviceID:  "device-789",
		RequestID: "request-abc",
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ToContext(ctx, info)
	}
}

// BenchmarkFromContext 测试 FromContext 性能
func BenchmarkFromContext(b *testing.B) {
	info := &Info{
		TraceID:   "trace-123",
		SpanID:    "span-456",
		DeviceID:  "device-789",
		RequestID: "request-abc",
	}
	ctx := ToContext(context.Background(), info)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FromContext(ctx)
	}
}

// BenchmarkHTTPMiddleware 测试 HTTP 中间件性能
func BenchmarkHTTPMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := HTTPMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)
	}
}

// BenchmarkSpanAttrs 测试 Span.Attrs 性能
func BenchmarkSpanAttrs(b *testing.B) {
	span := StartSpan(context.Background(), SpanKindService, "test")
	span.Set("key1", "value1")
	span.Set("key2", "value2")
	span.Set("key3", "value3")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = span.Attrs()
	}
}
