package requestid

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

type mockContext struct {
	method      string
	path        string
	respCode    int
	respBody    any
	store       map[string]any
	headers     map[string]string
	respHeaders map[string]string
}

func newMockContext() *mockContext {
	return &mockContext{
		method:      "GET",
		path:        "/test",
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
func (m *mockContext) ClientIP() string                        { return "127.0.0.1" }
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

func TestNew_GeneratesID(t *testing.T) {
	mw := New()
	var capturedID string
	handler := mw(func(ctx httpx.Context) error {
		v, _ := ctx.Get("request_id")
		capturedID = v.(string)
		return nil
	})
	ctx := newMockContext()
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if capturedID == "" {
		t.Error("expected request_id to be set")
	}
	if ctx.respHeaders[defaultHeaderKey] == "" {
		t.Error("expected X-Request-ID response header")
	}
	if ctx.respHeaders[defaultHeaderKey] != capturedID {
		t.Error("header and context value should match")
	}
}

func TestNew_PreservesExisting(t *testing.T) {
	mw := New()
	handler := mw(func(ctx httpx.Context) error { return nil })
	ctx := newMockContext()
	ctx.headers[defaultHeaderKey] = "existing-id-123"
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.respHeaders[defaultHeaderKey] != "existing-id-123" {
		t.Errorf("expected preserved ID, got %q", ctx.respHeaders[defaultHeaderKey])
	}
}

func TestNew_Unique(t *testing.T) {
	mw := New()
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		ctx := newMockContext()
		mw(func(ctx httpx.Context) error { return nil })(ctx)
		id := ctx.respHeaders[defaultHeaderKey]
		if ids[id] {
			t.Fatalf("duplicate request ID: %s", id)
		}
		ids[id] = true
	}
}

func TestNew_WithCustomHeaderKey(t *testing.T) {
	mw := New(WithHeaderKey("X-Custom-ID"))
	handler := mw(func(ctx httpx.Context) error { return nil })
	ctx := newMockContext()
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.respHeaders["X-Custom-ID"] == "" {
		t.Error("expected X-Custom-ID response header")
	}
	if ctx.respHeaders[defaultHeaderKey] != "" {
		t.Error("should not set default header when custom key is used")
	}
}

func TestNew_WithGenerator(t *testing.T) {
	counter := 0
	mw := New(WithGenerator(func() string {
		counter++
		return "custom-id"
	}))
	handler := mw(func(ctx httpx.Context) error { return nil })
	ctx := newMockContext()
	handler(ctx)
	if ctx.respHeaders[defaultHeaderKey] != "custom-id" {
		t.Errorf("expected 'custom-id', got %q", ctx.respHeaders[defaultHeaderKey])
	}
	if counter != 1 {
		t.Errorf("expected generator called once, got %d", counter)
	}
}

func TestNew_WithHandler(t *testing.T) {
	var handlerID string
	mw := New(WithHandler(func(ctx httpx.Context, id string) {
		handlerID = id
	}))
	handler := mw(func(ctx httpx.Context) error { return nil })
	ctx := newMockContext()
	handler(ctx)
	if handlerID == "" {
		t.Error("expected handler to be called with ID")
	}
}

func TestGet(t *testing.T) {
	ctx := newMockContext()
	ctx.headers[defaultHeaderKey] = "test-id"
	if Get(ctx) != "test-id" {
		t.Errorf("expected 'test-id', got %q", Get(ctx))
	}
}
