// Package trace 提供轻量级链路追踪功能
// 用于在应用各层之间传递追踪信息，记录操作耗时和属性
package trace

import "context"

// Getter 定义获取值的接口
// gin.Context 等框架的 context 通常实现了此接口
type Getter interface {
	Get(key any) (value any, exists bool)
}

// ContextKey 用于 context 中存储追踪信息的 key 类型
type ContextKey string

// Context key 常量定义
const (
	KeyTraceID   ContextKey = "trace_id"
	KeySpanID    ContextKey = "span_id"
	KeyDeviceID  ContextKey = "device_id"
	KeyRequestID ContextKey = "request_id"
)

// Info 追踪信息结构体
type Info struct {
	TraceID   string `json:"trace_id,omitempty"`
	SpanID    string `json:"span_id,omitempty"`
	DeviceID  string `json:"device_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// FromContext 从 context.Context 中提取追踪信息
// 优化：优先从单个 key 读取完整结构，回退到逐字段读取以保持向后兼容
func FromContext(ctx context.Context) *Info {
	if ctx == nil {
		return &Info{}
	}

	// 优先尝试从新格式读取（单次 Value 调用）
	if v := ctx.Value(infoKey); v != nil {
		if info, ok := v.(*Info); ok {
			return info
		}
	}

	// 回退到旧格式（向后兼容）
	info := &Info{}
	if v := ctx.Value(KeyTraceID); v != nil {
		if s, ok := v.(string); ok {
			info.TraceID = s
		}
	}
	if v := ctx.Value(KeySpanID); v != nil {
		if s, ok := v.(string); ok {
			info.SpanID = s
		}
	}
	if v := ctx.Value(KeyDeviceID); v != nil {
		if s, ok := v.(string); ok {
			info.DeviceID = s
		}
	}
	if v := ctx.Value(KeyRequestID); v != nil {
		if s, ok := v.(string); ok {
			info.RequestID = s
		}
	}

	return info
}

// FromGetter 从实现了 Getter 接口的对象中提取追踪信息
// 适用于 gin.Context 等 HTTP 框架的 context
//
// 使用示例：
//
//	func Handler(ctx *gin.Context) {
//	    info := trace.FromGetter(ctx)
//	}
func FromGetter(g Getter) *Info {
	if g == nil {
		return &Info{}
	}

	info := &Info{}

	if v, ok := g.Get(string(KeyTraceID)); ok {
		if s, ok := v.(string); ok {
			info.TraceID = s
		}
	}

	if v, ok := g.Get(string(KeySpanID)); ok {
		if s, ok := v.(string); ok {
			info.SpanID = s
		}
	}

	if v, ok := g.Get(string(KeyDeviceID)); ok {
		if s, ok := v.(string); ok {
			info.DeviceID = s
		}
	}

	if v, ok := g.Get(string(KeyRequestID)); ok {
		if s, ok := v.(string); ok {
			info.RequestID = s
		}
	}

	return info
}

// GetTraceID 从 Getter 获取 TraceID
func GetTraceID(g Getter) string {
	return getString(g, KeyTraceID)
}

// GetSpanID 从 Getter 获取 SpanID
func GetSpanID(g Getter) string {
	return getString(g, KeySpanID)
}

// GetDeviceID 从 Getter 获取 DeviceID
func GetDeviceID(g Getter) string {
	return getString(g, KeyDeviceID)
}

// GetRequestID 从 Getter 获取 RequestID
func GetRequestID(g Getter) string {
	return getString(g, KeyRequestID)
}

// getString 从 Getter 获取字符串值的通用辅助函数
func getString(g Getter, key ContextKey) string {
	if v, ok := g.Get(string(key)); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// infoKey 用于存储完整 Info 结构的 context key
const infoKey ContextKey = "trace_info"

// ToContext 将追踪信息注入到 context
// 优化：使用单个 context.WithValue 存储整个 Info 结构，避免多次嵌套调用
func ToContext(ctx context.Context, info *Info) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if info == nil {
		return ctx
	}

	// 单次 WithValue 调用，减少内存分配
	return context.WithValue(ctx, infoKey, info)
}

// WithTraceID 设置 TraceID 到 context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, KeyTraceID, traceID)
}

// WithSpanID 设置 SpanID 到 context
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, KeySpanID, spanID)
}

// WithDeviceID 设置 DeviceID 到 context
func WithDeviceID(ctx context.Context, deviceID string) context.Context {
	return context.WithValue(ctx, KeyDeviceID, deviceID)
}

// WithRequestID 设置 RequestID 到 context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, KeyRequestID, requestID)
}
