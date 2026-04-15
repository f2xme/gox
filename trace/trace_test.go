package trace

import (
	"context"
	"testing"
)

// mockGetter 模拟实现 Getter 接口
type mockGetter struct {
	data map[any]any
}

func (m *mockGetter) Get(key any) (any, bool) {
	v, ok := m.data[key]
	return v, ok
}

func TestFromGetter(t *testing.T) {
	g := &mockGetter{
		data: map[any]any{
			string(KeyTraceID):   "trace-123",
			string(KeySpanID):    "span-456",
			string(KeyDeviceID):  "device-789",
			string(KeyRequestID): "request-abc",
		},
	}

	info := FromGetter(g)

	if info.TraceID != "trace-123" {
		t.Errorf("expected trace_id 'trace-123', got '%s'", info.TraceID)
	}
	if info.SpanID != "span-456" {
		t.Errorf("expected span_id 'span-456', got '%s'", info.SpanID)
	}
	if info.DeviceID != "device-789" {
		t.Errorf("expected device_id 'device-789', got '%s'", info.DeviceID)
	}
	if info.RequestID != "request-abc" {
		t.Errorf("expected request_id 'request-abc', got '%s'", info.RequestID)
	}
}

func TestFromGetter_Nil(t *testing.T) {
	info := FromGetter(nil)
	if info == nil {
		t.Error("expected non-nil Info for nil getter")
	}
}

func TestFromGetter_Empty(t *testing.T) {
	g := &mockGetter{data: map[any]any{}}
	info := FromGetter(g)

	if info.TraceID != "" || info.SpanID != "" || info.DeviceID != "" || info.RequestID != "" {
		t.Error("expected empty Info for empty getter")
	}
}

func TestGetTraceID(t *testing.T) {
	g := &mockGetter{
		data: map[any]any{
			string(KeyTraceID): "trace-123",
		},
	}

	if got := GetTraceID(g); got != "trace-123" {
		t.Errorf("expected 'trace-123', got '%s'", got)
	}
}

func TestGetSpanID(t *testing.T) {
	g := &mockGetter{
		data: map[any]any{
			string(KeySpanID): "span-456",
		},
	}

	if got := GetSpanID(g); got != "span-456" {
		t.Errorf("expected 'span-456', got '%s'", got)
	}
}

func TestGetDeviceID(t *testing.T) {
	g := &mockGetter{
		data: map[any]any{
			string(KeyDeviceID): "device-789",
		},
	}

	if got := GetDeviceID(g); got != "device-789" {
		t.Errorf("expected 'device-789', got '%s'", got)
	}
}

func TestGetRequestID(t *testing.T) {
	g := &mockGetter{
		data: map[any]any{
			string(KeyRequestID): "request-abc",
		},
	}

	if got := GetRequestID(g); got != "request-abc" {
		t.Errorf("expected 'request-abc', got '%s'", got)
	}
}

func TestFromContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-ctx-123")
	ctx = WithSpanID(ctx, "span-ctx-456")
	ctx = WithDeviceID(ctx, "device-ctx-789")
	ctx = WithRequestID(ctx, "request-ctx-abc")

	info := FromContext(ctx)

	if info.TraceID != "trace-ctx-123" {
		t.Errorf("expected trace_id 'trace-ctx-123', got '%s'", info.TraceID)
	}
	if info.SpanID != "span-ctx-456" {
		t.Errorf("expected span_id 'span-ctx-456', got '%s'", info.SpanID)
	}
	if info.DeviceID != "device-ctx-789" {
		t.Errorf("expected device_id 'device-ctx-789', got '%s'", info.DeviceID)
	}
	if info.RequestID != "request-ctx-abc" {
		t.Errorf("expected request_id 'request-ctx-abc', got '%s'", info.RequestID)
	}
}

func TestToContext(t *testing.T) {
	info := &Info{
		TraceID:   "trace-to-123",
		SpanID:    "span-to-456",
		DeviceID:  "device-to-789",
		RequestID: "request-to-abc",
	}

	ctx := ToContext(context.Background(), info)
	extracted := FromContext(ctx)

	if extracted.TraceID != info.TraceID {
		t.Errorf("expected '%s', got '%s'", info.TraceID, extracted.TraceID)
	}
	if extracted.SpanID != info.SpanID {
		t.Errorf("expected '%s', got '%s'", info.SpanID, extracted.SpanID)
	}
	if extracted.DeviceID != info.DeviceID {
		t.Errorf("expected '%s', got '%s'", info.DeviceID, extracted.DeviceID)
	}
	if extracted.RequestID != info.RequestID {
		t.Errorf("expected '%s', got '%s'", info.RequestID, extracted.RequestID)
	}
}

func TestToContext_NilInfo(t *testing.T) {
	ctx := ToContext(context.Background(), nil)
	if ctx == nil {
		t.Error("expected non-nil context")
	}
}

func TestToContext_NilContext(t *testing.T) {
	info := &Info{TraceID: "trace-123"}
	ctx := ToContext(context.TODO(), info)
	if ctx == nil {
		t.Error("expected non-nil context")
	}

	extracted := FromContext(ctx)
	if extracted.TraceID != "trace-123" {
		t.Errorf("expected 'trace-123', got '%s'", extracted.TraceID)
	}
}
