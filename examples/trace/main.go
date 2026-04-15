package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/f2xme/gox/trace"
)

func main() {
	fmt.Println("=== trace 包使用示例 ===")

	// 示例 1: 使用 ToContext 注入追踪信息
	fmt.Println("\n示例 1: 使用 ToContext 注入追踪信息")
	ctx := context.Background()
	info := &trace.Info{
		TraceID:   "trace-123456",
		SpanID:    "span-001",
		RequestID: "req-789",
		DeviceID:  "device-abc",
	}
	ctx = trace.ToContext(ctx, info)
	fmt.Println("已注入追踪信息到 context")

	// 示例 2: 从 context 提取追踪信息
	fmt.Println("\n示例 2: 从 context 提取追踪信息")
	extractedInfo := trace.FromContext(ctx)
	fmt.Printf("TraceID:   %s\n", extractedInfo.TraceID)
	fmt.Printf("SpanID:    %s\n", extractedInfo.SpanID)
	fmt.Printf("RequestID: %s\n", extractedInfo.RequestID)
	fmt.Printf("DeviceID:  %s\n", extractedInfo.DeviceID)

	// 示例 3: 使用 Span 记录操作耗时
	fmt.Println("\n示例 3: 使用 Span 记录操作耗时")
	trace.SetCallback(func(r *trace.SpanResult) {
		status := "成功"
		if r.Error != nil {
			status = fmt.Sprintf("失败: %v", r.Error)
		}
		fmt.Printf("[%s] %s 耗时=%dms 状态=%s\n",
			r.Kind(), r.Name(), r.DurationMs(), status)
	})

	spanCtx := trace.WithTraceID(context.Background(), "trace-span-demo")
	getUserFromService(spanCtx, 123)
	getUserFromDAO(spanCtx, 456)
	getUserFromCache(spanCtx, "user:789")

	// 示例 4: HTTP 中间件自动注入追踪信息
	fmt.Println("\n示例 4: HTTP 中间件自动注入追踪信息")
	mux := http.NewServeMux()
	mux.HandleFunc("/api/user", handleUser)

	handler := trace.HTTPMiddleware(mux)
	fmt.Println("HTTP 服务器已配置追踪中间件")
	fmt.Println("访问 /api/user 时会自动注入 TraceID、SpanID、RequestID")

	// 模拟 HTTP 请求
	req, _ := http.NewRequest("GET", "/api/user", nil)
	req.Header.Set(trace.HeaderTraceID, "trace-http-001")
	rr := &mockResponseWriter{header: make(http.Header)}
	handler.ServeHTTP(rr, req)

	// 示例 5: 跨服务调用传递追踪信息
	fmt.Println("\n示例 5: 跨服务调用传递追踪信息")
	callCtx := trace.WithTraceID(context.Background(), "trace-downstream-001")
	callCtx = trace.WithSpanID(callCtx, "span-caller-001")

	fmt.Println("使用 InjectToHeaders 手动注入:")
	req2, _ := http.NewRequest("GET", "http://api.example.com/user", nil)
	trace.InjectToHeaders(callCtx, req2.Header)
	fmt.Printf("  %s: %s\n", trace.HeaderTraceID, req2.Header.Get(trace.HeaderTraceID))
	fmt.Printf("  %s: %s\n", trace.HeaderSpanID, req2.Header.Get(trace.HeaderSpanID))

	fmt.Println("\n使用 HTTPClient 自动注入:")
	client := trace.HTTPClient(callCtx)
	fmt.Printf("  已创建带追踪信息的 HTTP Client: %T\n", client)

	fmt.Println("\n=== 示例结束 ===")
	fmt.Println("\n提示：")
	fmt.Println("- 使用 ToContext/WithTraceID 等函数注入追踪信息")
	fmt.Println("- 使用 FromContext 提取追踪信息")
	fmt.Println("- 使用 Service/DAO/Cache 等函数记录 Span 耗时")
	fmt.Println("- 使用 HTTPMiddleware 自动处理 HTTP 请求的追踪信息")
	fmt.Println("- 使用 InjectToHeaders 或 HTTPClient 传递追踪信息到下游服务")
}

// getUserFromService 模拟 Service 层调用
func getUserFromService(ctx context.Context, userID int64) (err error) {
	defer trace.Service(ctx, "GetUser", "user_id", userID)(&err)
	time.Sleep(50 * time.Millisecond)
	return nil
}

// getUserFromDAO 模拟 DAO 层调用
func getUserFromDAO(ctx context.Context, userID int64) (err error) {
	defer trace.DAO(ctx, "GetUserByID", "user_id", userID)(&err)
	time.Sleep(30 * time.Millisecond)
	return errors.New("用户不存在")
}

// getUserFromCache 模拟 Cache 层调用
func getUserFromCache(ctx context.Context, key string) (err error) {
	defer trace.Cache(ctx, "Get", "key", key)(&err)
	time.Sleep(10 * time.Millisecond)
	return nil
}

// handleUser HTTP 处理函数
func handleUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	info := trace.FromContext(ctx)

	fmt.Printf("\n[HTTP Handler] 收到请求:\n")
	fmt.Printf("  TraceID:   %s\n", info.TraceID)
	fmt.Printf("  SpanID:    %s\n", info.SpanID)
	fmt.Printf("  RequestID: %s\n", info.RequestID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// mockResponseWriter 模拟 http.ResponseWriter
type mockResponseWriter struct {
	header http.Header
}

func (m *mockResponseWriter) Header() http.Header {
	return m.header
}

func (m *mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {}
