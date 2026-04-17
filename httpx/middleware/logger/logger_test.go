package logger

import (
	"net/http"
	"testing"

	"github.com/f2xme/gox/httpx"
)

func getMsg(msg []string, def string) string {
	if len(msg) > 0 && msg[0] != "" {
		return msg[0]
	}
	return def
}

type mockLogger struct {
	messages []string
	fields   [][]any
}

func (l *mockLogger) Info(msg string, keysAndValues ...any) {
	l.messages = append(l.messages, msg)
	l.fields = append(l.fields, keysAndValues)
}

type mockContext struct {
	method      string
	path        string
	clientIP    string
	respCode    int
	respBody    any
	store       map[string]any
	headers     map[string]string
	respHeaders map[string]string
}

func newMockContext(method, path string) *mockContext {
	return &mockContext{
		method:      method,
		path:        path,
		clientIP:    "127.0.0.1",
		store:       make(map[string]any),
		headers:     make(map[string]string),
		respHeaders: make(map[string]string),
	}
}

func (m *mockContext) Request() *http.Request                   { return nil }
func (m *mockContext) Param(string) string                     { return "" }
func (m *mockContext) Query(string) string                     { return "" }
func (m *mockContext) QueryDefault(_, def string) string       { return def }
func (m *mockContext) Header(key string) string                { return m.headers[key] }
func (m *mockContext) Cookie(string) (*http.Cookie, error)     { return nil, http.ErrNoCookie }
func (m *mockContext) ClientIP() string                        { return m.clientIP }
func (m *mockContext) Method() string                          { return m.method }
func (m *mockContext) Path() string                            { return m.path }
func (m *mockContext) Bind(any) error                          { return nil }
func (m *mockContext) BindJSON(any) error                      { return nil }
func (m *mockContext) BindQuery(any) error                     { return nil }
func (m *mockContext) BindForm(any) error                      { return nil }
func (m *mockContext) JSON(code int, v any) error              { m.respCode = code; m.respBody = v; return nil }
func (m *mockContext) String(code int, s string) error         { m.respCode = code; m.respBody = s; return nil }
func (m *mockContext) HTML(code int, _ string) error           { m.respCode = code; return nil }
func (m *mockContext) Blob(code int, _ string, _ []byte) error { m.respCode = code; return nil }
func (m *mockContext) NoContent(code int) error                { m.respCode = code; return nil }
func (m *mockContext) Redirect(code int, _ string) error       { m.respCode = code; return nil }
func (m *mockContext) SetHeader(key, value string)             { m.respHeaders[key] = value }
func (m *mockContext) SetCookie(*http.Cookie)                  {}
func (m *mockContext) Status(code int)                         { m.respCode = code }
func (m *mockContext) Success(data any) error                  { return m.JSON(200, data) }
func (m *mockContext) Fail(msg string) error                   { return m.JSON(200, msg) }
func (m *mockContext) BadRequest(msg ...string) error          { return m.JSON(400, getMsg(msg, "Bad Request")) }
func (m *mockContext) Unauthorized(msg ...string) error        { return m.JSON(401, getMsg(msg, "Unauthorized")) }
func (m *mockContext) Forbidden(msg ...string) error           { return m.JSON(403, getMsg(msg, "Forbidden")) }
func (m *mockContext) NotFound(msg ...string) error            { return m.JSON(404, getMsg(msg, "Not Found")) }
func (m *mockContext) TooManyRequests(msg ...string) error     { return m.JSON(429, getMsg(msg, "Too Many Requests")) }
func (m *mockContext) InternalError(msg ...string) error       { return m.JSON(500, getMsg(msg, "Internal Server Error")) }
func (m *mockContext) ServiceUnavailable(msg ...string) error  { return m.JSON(503, getMsg(msg, "Service Unavailable")) }
func (m *mockContext) Set(key string, value any)               { m.store[key] = value }
func (m *mockContext) Get(key string) (any, bool)              { v, ok := m.store[key]; return v, ok }
func (m *mockContext) MustGet(key string) any                  { return m.store[key] }
func (m *mockContext) ResponseWriter() http.ResponseWriter     { return nil }
func (m *mockContext) Raw() any                                { return nil }

func TestNew_LogsRequest(t *testing.T) {
	l := &mockLogger{}
	mw := New(WithLogger(l))
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})

	ctx := newMockContext("POST", "/api/users")
	ctx.clientIP = "192.168.1.1"
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

	ctx := newMockContext("GET", "/fail")
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
	ctx := newMockContext("GET", "/test")
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestNew_SkipPath(t *testing.T) {
	l := &mockLogger{}
	mw := New(WithLogger(l), WithSkipPath("/health", "/ping"))

	handler := mw(func(ctx httpx.Context) error { return nil })

	// Skipped path
	ctx := newMockContext("GET", "/health")
	handler(ctx)
	if len(l.messages) != 0 {
		t.Error("expected no log for skipped path")
	}

	// Non-skipped path
	ctx2 := newMockContext("GET", "/api/users")
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
	ctx := newMockContext("OPTIONS", "/test")
	handler(ctx)
	if len(l.messages) != 0 {
		t.Error("expected no log for skipped method")
	}

	// Non-skipped method
	ctx2 := newMockContext("GET", "/test")
	handler(ctx2)
	if len(l.messages) != 1 {
		t.Error("expected log for non-skipped method")
	}
}
