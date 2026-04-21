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

// HTTPMiddleware 返回标准库 http.Handler 中间件：从请求头提取或补全
// 追踪信息，注入 context，并回写到响应头。
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := processTraceMiddleware(w, r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HTTPMiddlewareFunc 是 HTTPMiddleware 针对 http.HandlerFunc 的便捷形式。
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

	h := w.Header()
	h.Set(HeaderTraceID, info.TraceID)
	h.Set(HeaderSpanID, info.SpanID)
	h.Set(HeaderRequestID, info.RequestID)

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

// InjectToHeaders 将 context 中的追踪信息注入到请求头，
// 用于客户端向下游发起请求时透传链路。
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

// HTTPClient 返回一个自动将 ctx 中追踪信息注入到出站请求头的 http.Client。
func HTTPClient(ctx context.Context) *http.Client {
	return &http.Client{
		Transport: &tracingTransport{ctx: ctx, base: http.DefaultTransport},
	}
}

// tracingTransport 从构造时绑定的 ctx 读取追踪信息并注入出站请求头。
type tracingTransport struct {
	ctx  context.Context
	base http.RoundTripper
}

func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 注入追踪信息
	InjectToHeaders(t.ctx, req.Header)
	return t.base.RoundTrip(req)
}

// generateID 生成 128 位随机十六进制 ID（32 字符）。
// crypto/rand.Read 自 Go 1.24 起保证不返回错误，此处显式丢弃以通过静态检查。
func generateID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
