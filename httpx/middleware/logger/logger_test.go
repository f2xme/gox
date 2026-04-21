package logger

import (
	"testing"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/mock"
)

type mockLogger struct {
	messages []string
	fields   [][]any
}

func (l *mockLogger) Info(msg string, keysAndValues ...any) {
	l.messages = append(l.messages, msg)
	l.fields = append(l.fields, keysAndValues)
}

func TestNew_LogsRequest(t *testing.T) {
	l := &mockLogger{}
	mw := New(WithLogger(l))
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})

	ctx := mock.NewMockContext("POST", "/api/users")
	ctx.ClientIPValue = "192.168.1.1"
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if len(l.messages) != 1 {
		t.Fatalf("expected 1 log message, got %d", len(l.messages))
	}
	if l.messages[0] != "request" {
		t.Errorf("expected 'request', got %q", l.messages[0])
	}

	fieldMap := make(map[string]any)
	fields := l.fields[0]
	for i := 0; i < len(fields)-1; i += 2 {
		fieldMap[fields[i].(string)] = fields[i+1]
	}
	if fieldMap["method"] != "POST" {
		t.Errorf("expected method=POST, got %v", fieldMap["method"])
	}
	if fieldMap["path"] != "/api/users" {
		t.Errorf("expected path=/api/users, got %v", fieldMap["path"])
	}
	if fieldMap["ip"] != "192.168.1.1" {
		t.Errorf("expected ip=192.168.1.1, got %v", fieldMap["ip"])
	}
}

func TestNew_LogsError(t *testing.T) {
	l := &mockLogger{}
	mw := New(WithLogger(l))
	expectedErr := httpx.NewHTTPError(500, "internal")
	handler := mw(func(ctx httpx.Context) error {
		return expectedErr
	})

	ctx := mock.NewMockContext("GET", "/fail")
	err := handler(ctx)
	if err != expectedErr {
		t.Error("expected error to be propagated")
	}

	fieldMap := make(map[string]any)
	fields := l.fields[0]
	for i := 0; i < len(fields)-1; i += 2 {
		fieldMap[fields[i].(string)] = fields[i+1]
	}
	if fieldMap["error"] != expectedErr {
		t.Error("expected error in log fields")
	}
}

func TestNew_NoLogger_NoOp(t *testing.T) {
	mw := New() // no logger = no-op
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := mock.NewMockContext("GET", "/test")
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestNew_SkipPath(t *testing.T) {
	l := &mockLogger{}
	mw := New(WithLogger(l), WithSkipPath("/health", "/ping"))

	handler := mw(func(ctx httpx.Context) error { return nil })

	// Skipped path
	ctx := mock.NewMockContext("GET", "/health")
	handler(ctx)
	if len(l.messages) != 0 {
		t.Error("expected no log for skipped path")
	}

	// Non-skipped path
	ctx2 := mock.NewMockContext("GET", "/api/users")
	handler(ctx2)
	if len(l.messages) != 1 {
		t.Error("expected log for non-skipped path")
	}
}

func TestNew_SkipMethod(t *testing.T) {
	l := &mockLogger{}
	mw := New(WithLogger(l), WithSkipMethod("OPTIONS", "HEAD"))

	handler := mw(func(ctx httpx.Context) error { return nil })

	// Skipped method
	ctx := mock.NewMockContext("OPTIONS", "/test")
	handler(ctx)
	if len(l.messages) != 0 {
		t.Error("expected no log for skipped method")
	}

	// Non-skipped method
	ctx2 := mock.NewMockContext("GET", "/test")
	handler(ctx2)
	if len(l.messages) != 1 {
		t.Error("expected log for non-skipped method")
	}
}
