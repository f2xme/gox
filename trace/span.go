package trace

import (
	"context"
	"maps"
	"time"
)

// SpanKind 表示 Span 的类型
type SpanKind string

const (
	SpanKindHTTP    SpanKind = "http"
	SpanKindService SpanKind = "service"
	SpanKindDAO     SpanKind = "dao"
	SpanKindCache   SpanKind = "cache"
	SpanKindRPC     SpanKind = "rpc"
	SpanKindMQ      SpanKind = "mq"
)

// Span 表示一个追踪单元
type Span struct {
	ctx       context.Context
	name      string
	kind      SpanKind
	startTime time.Time
	attrs     map[string]any
}

// StartSpan 开始一个新的 Span
func StartSpan(ctx context.Context, kind SpanKind, name string) *Span {
	return &Span{
		ctx:       ctx,
		name:      name,
		kind:      kind,
		startTime: time.Now(),
	}
}

// Set 设置属性
func (s *Span) Set(key string, value any) *Span {
	if s.attrs == nil {
		s.attrs = make(map[string]any)
	}
	s.attrs[key] = value
	return s
}

// Context 返回 Span 的 context
func (s *Span) Context() context.Context {
	return s.ctx
}

// Name 返回 Span 名称
func (s *Span) Name() string {
	return s.name
}

// Kind 返回 Span 类型
func (s *Span) Kind() SpanKind {
	return s.kind
}

// Duration 返回 Span 持续时间
func (s *Span) Duration() time.Duration {
	return time.Since(s.startTime)
}

// Attrs 返回所有属性的副本，防止外部修改
func (s *Span) Attrs() map[string]any {
	if s.attrs == nil {
		return nil
	}
	// 返回副本以保护内部状态
	return maps.Clone(s.attrs)
}

// End 结束 Span，返回 SpanResult
func (s *Span) End(err error) *SpanResult {
	return &SpanResult{
		Span:  s,
		Error: err,
	}
}

// SpanResult Span 执行结果
type SpanResult struct {
	*Span
	Error error
}

// Success 是否成功
func (r *SpanResult) Success() bool {
	return r.Error == nil
}

// DurationMs 返回毫秒级耗时
func (r *SpanResult) DurationMs() int64 {
	return r.Duration().Milliseconds()
}
