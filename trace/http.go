package trace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// HTTP Header 常量
const (
	HeaderTraceID   = "X-Trace-ID"
	HeaderSpanID    = "X-Span-ID"
	HeaderRequestID = "X-Request-ID"
	HeaderDeviceID  = "X-Device-ID"
)

// HTTPMiddleware 返回一个标准库 http.Handler 中间件
// 自动从 HTTP Header 提取或生成追踪信息，并注入到 context
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := processTraceMiddleware(w, r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HTTPMiddlewareFunc 返回一个 http.HandlerFunc 中间件
// 用于包装 http.HandlerFunc 类型的处理器
func HTTPMiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := processTraceMiddleware(w, r)
		next(w, r.WithContext(ctx))
	}
}

// processTraceMiddleware 处理追踪中间件的核心逻辑
func processTraceMiddleware(w http.ResponseWriter, r *http.Request) context.Context {
	info := extractFromHeaders(r)

	if info.TraceID == "" {
		info.TraceID = generateID()
	}
	if info.SpanID == "" {
		info.SpanID = generateID()
	}
	if info.RequestID == "" {
		info.RequestID = generateID()
	}

	w.Header().Set(HeaderTraceID, info.TraceID)
	w.Header().Set(HeaderSpanID, info.SpanID)
	w.Header().Set(HeaderRequestID, info.RequestID)

	return ToContext(r.Context(), info)
}

// extractFromHeaders 从 HTTP Header 中提取追踪信息
func extractFromHeaders(r *http.Request) *Info {
	return &Info{
		TraceID:   r.Header.Get(HeaderTraceID),
		SpanID:    r.Header.Get(HeaderSpanID),
		RequestID: r.Header.Get(HeaderRequestID),
		DeviceID:  r.Header.Get(HeaderDeviceID),
	}
}

// InjectToHeaders 将追踪信息注入到 HTTP Header
// 用于客户端发起请求时传递追踪信息
func InjectToHeaders(ctx context.Context, header http.Header) {
	info := FromContext(ctx)
	setHeaderIfNotEmpty(header, HeaderTraceID, info.TraceID)
	setHeaderIfNotEmpty(header, HeaderSpanID, info.SpanID)
	setHeaderIfNotEmpty(header, HeaderRequestID, info.RequestID)
	setHeaderIfNotEmpty(header, HeaderDeviceID, info.DeviceID)
}

// setHeaderIfNotEmpty 如果值非空则设置 header
func setHeaderIfNotEmpty(header http.Header, key, value string) {
	if value != "" {
		header.Set(key, value)
	}
}

// HTTPClient 返回一个带追踪信息的 http.Client
// 自动在请求中注入追踪信息
func HTTPClient(ctx context.Context) *http.Client {
	return &http.Client{
		Transport: &tracingTransport{
			ctx:  ctx,
			base: http.DefaultTransport,
		},
	}
}

// tracingTransport 实现 http.RoundTripper 接口
// 自动注入追踪信息到请求头
type tracingTransport struct {
	ctx  context.Context
	base http.RoundTripper
}

func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 注入追踪信息
	InjectToHeaders(t.ctx, req.Header)
	return t.base.RoundTrip(req)
}

// generateID 生成唯一 ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
