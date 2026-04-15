package trace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := FromContext(r.Context())
		if info.TraceID == "" {
			t.Error("expected trace_id in context")
		}
		if info.SpanID == "" {
			t.Error("expected span_id in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := HTTPMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	// 检查响应头
	if rec.Header().Get(HeaderTraceID) == "" {
		t.Error("expected X-Trace-ID in response header")
	}
	if rec.Header().Get(HeaderSpanID) == "" {
		t.Error("expected X-Span-ID in response header")
	}
}

func TestHTTPMiddleware_WithExistingTraceID(t *testing.T) {
	expectedTraceID := "existing-trace-123"
	expectedSpanID := "existing-span-456"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := FromContext(r.Context())
		if info.TraceID != expectedTraceID {
			t.Errorf("expected trace_id %s, got %s", expectedTraceID, info.TraceID)
		}
		if info.SpanID != expectedSpanID {
			t.Errorf("expected span_id %s, got %s", expectedSpanID, info.SpanID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := HTTPMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(HeaderTraceID, expectedTraceID)
	req.Header.Set(HeaderSpanID, expectedSpanID)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	// 检查响应头保留了原有的 TraceID
	if rec.Header().Get(HeaderTraceID) != expectedTraceID {
		t.Errorf("expected trace_id %s in response, got %s", expectedTraceID, rec.Header().Get(HeaderTraceID))
	}
}

func TestHTTPMiddlewareFunc(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		info := FromContext(r.Context())
		if info.TraceID == "" {
			t.Error("expected trace_id in context")
		}
		w.WriteHeader(http.StatusOK)
	}

	middleware := HTTPMiddlewareFunc(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware(rec, req)

	if rec.Header().Get(HeaderTraceID) == "" {
		t.Error("expected X-Trace-ID in response header")
	}
}

func TestExtractFromHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(HeaderTraceID, "trace-123")
	req.Header.Set(HeaderSpanID, "span-456")
	req.Header.Set(HeaderRequestID, "req-789")
	req.Header.Set(HeaderDeviceID, "device-abc")

	info := extractFromHeaders(req)

	if info.TraceID != "trace-123" {
		t.Errorf("expected trace_id 'trace-123', got '%s'", info.TraceID)
	}
	if info.SpanID != "span-456" {
		t.Errorf("expected span_id 'span-456', got '%s'", info.SpanID)
	}
	if info.RequestID != "req-789" {
		t.Errorf("expected request_id 'req-789', got '%s'", info.RequestID)
	}
	if info.DeviceID != "device-abc" {
		t.Errorf("expected device_id 'device-abc', got '%s'", info.DeviceID)
	}
}

func TestInjectToHeaders(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-123")
	ctx = WithSpanID(ctx, "span-456")
	ctx = WithRequestID(ctx, "req-789")

	header := http.Header{}
	InjectToHeaders(ctx, header)

	if header.Get(HeaderTraceID) != "trace-123" {
		t.Errorf("expected X-Trace-ID 'trace-123', got '%s'", header.Get(HeaderTraceID))
	}
	if header.Get(HeaderSpanID) != "span-456" {
		t.Errorf("expected X-Span-ID 'span-456', got '%s'", header.Get(HeaderSpanID))
	}
	if header.Get(HeaderRequestID) != "req-789" {
		t.Errorf("expected X-Request-ID 'req-789', got '%s'", header.Get(HeaderRequestID))
	}
}

func TestHTTPClient(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查请求头是否包含追踪信息
		if r.Header.Get(HeaderTraceID) == "" {
			t.Error("expected X-Trace-ID in request header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 创建带追踪信息的 context
	ctx := WithTraceID(context.Background(), "test-trace-123")

	// 使用 HTTPClient 发起请求
	client := HTTPClient(ctx)
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestTracingTransport(t *testing.T) {
	expectedTraceID := "transport-trace-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(HeaderTraceID) != expectedTraceID {
			t.Errorf("expected trace_id %s, got %s", expectedTraceID, r.Header.Get(HeaderTraceID))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := WithTraceID(context.Background(), expectedTraceID)
	transport := &tracingTransport{
		ctx:  ctx,
		base: http.DefaultTransport,
	}

	client := &http.Client{Transport: transport}
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	if id1 == "" {
		t.Error("expected non-empty ID")
	}
	if id2 == "" {
		t.Error("expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
}
