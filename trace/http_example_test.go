package trace_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/f2xme/gox/trace"
)

func Example_httpIntegration() {
	// 创建 HTTP 处理器
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从 context 中获取追踪信息
		info := trace.FromContext(r.Context())
		fmt.Printf("TraceID exists: %v\n", info.TraceID != "")
		fmt.Printf("SpanID exists: %v\n", info.SpanID != "")
		w.WriteHeader(http.StatusOK)
	})

	// 使用追踪中间件包装
	tracedHandler := trace.HTTPMiddleware(handler)

	// 创建测试请求
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()

	// 处理请求
	tracedHandler.ServeHTTP(rec, req)

	// 检查响应头
	fmt.Printf("Response has TraceID: %v\n", rec.Header().Get(trace.HeaderTraceID) != "")

	// Output:
	// TraceID exists: true
	// SpanID exists: true
	// Response has TraceID: true
}

func Example_httpClientWithTracing() {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 服务端接收到追踪信息
		traceID := r.Header.Get(trace.HeaderTraceID)
		fmt.Printf("Server received TraceID: %v\n", traceID != "")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 客户端创建带追踪信息的 context
	ctx := trace.WithTraceID(context.Background(), "client-trace-123")

	// 使用追踪客户端
	client := trace.HTTPClient(ctx)
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}

	fmt.Printf("Request sent with tracing\n")

	// Output:
	// Server received TraceID: true
	// Request sent with tracing
}

func Example_httpHeaderPropagation() {
	// 模拟上游服务传递的 TraceID
	upstreamTraceID := "upstream-trace-abc123"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := trace.FromContext(r.Context())
		// 验证 TraceID 被正确传递
		fmt.Printf("Received TraceID: %s\n", info.TraceID)
		w.WriteHeader(http.StatusOK)
	})

	tracedHandler := trace.HTTPMiddleware(handler)

	// 创建带有上游 TraceID 的请求
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set(trace.HeaderTraceID, upstreamTraceID)
	rec := httptest.NewRecorder()

	tracedHandler.ServeHTTP(rec, req)

	// 响应头也包含相同的 TraceID
	fmt.Printf("Response TraceID: %s\n", rec.Header().Get(trace.HeaderTraceID))

	// Output:
	// Received TraceID: upstream-trace-abc123
	// Response TraceID: upstream-trace-abc123
}
