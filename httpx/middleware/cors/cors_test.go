package cors

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

func newMockContext(method, path string) *mockContext {
	return &mockContext{
		method:      method,
		path:        path,
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

func TestNew_NoOrigin(t *testing.T) {
	mw := New(WithOrigins([]string{"http://example.com"}))
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := newMockContext("GET", "/test")
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if _, ok := ctx.respHeaders["Access-Control-Allow-Origin"]; ok {
		t.Error("should not set CORS headers without Origin")
	}
}

func TestNew_AllowedOrigin(t *testing.T) {
	mw := New(WithOrigins([]string{"http://example.com"}))
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := newMockContext("GET", "/test")
	ctx.headers["Origin"] = "http://example.com"
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.respHeaders["Access-Control-Allow-Origin"] != "http://example.com" {
		t.Error("expected Access-Control-Allow-Origin header")
	}
}

func TestNew_DisallowedOrigin(t *testing.T) {
	mw := New(WithOrigins([]string{"http://example.com"}))
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := newMockContext("GET", "/test")
	ctx.headers["Origin"] = "http://evil.com"
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if _, ok := ctx.respHeaders["Access-Control-Allow-Origin"]; ok {
		t.Error("should not set CORS headers for disallowed origin")
	}
}

func TestNew_DefaultWildcard(t *testing.T) {
	mw := New() // default allows all origins
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := newMockContext("GET", "/test")
	ctx.headers["Origin"] = "http://any-site.com"
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.respHeaders["Access-Control-Allow-Origin"] != "http://any-site.com" {
		t.Error("expected Access-Control-Allow-Origin for default wildcard")
	}
}

func TestNew_Preflight(t *testing.T) {
	mw := New(
		WithOrigins([]string{"http://example.com"}),
		WithMethods([]string{"GET", "POST"}),
		WithHeaders([]string{"Content-Type"}),
		WithMaxAge(3600),
	)
	nextCalled := false
	handler := mw(func(ctx httpx.Context) error {
		nextCalled = true
		return nil
	})
	ctx := newMockContext("OPTIONS", "/test")
	ctx.headers["Origin"] = "http://example.com"
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if nextCalled {
		t.Error("next should not be called for preflight")
	}
	if ctx.respCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", ctx.respCode)
	}
	if ctx.respHeaders["Access-Control-Allow-Methods"] != "GET, POST" {
		t.Errorf("unexpected Allow-Methods: %q", ctx.respHeaders["Access-Control-Allow-Methods"])
	}
	if ctx.respHeaders["Access-Control-Max-Age"] != "3600" {
		t.Errorf("unexpected Max-Age: %q", ctx.respHeaders["Access-Control-Max-Age"])
	}
}

func TestNew_WithCredentials(t *testing.T) {
	mw := New(WithCredentials(true))
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := newMockContext("GET", "/test")
	ctx.headers["Origin"] = "http://example.com"
	handler(ctx)
	if ctx.respHeaders["Access-Control-Allow-Credentials"] != "true" {
		t.Error("expected Allow-Credentials header")
	}
}

func TestNew_WithExposeHeaders(t *testing.T) {
	mw := New(WithExposeHeaders([]string{"X-Custom", "X-Other"}))
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := newMockContext("GET", "/test")
	ctx.headers["Origin"] = "http://example.com"
	handler(ctx)
	if ctx.respHeaders["Access-Control-Expose-Headers"] != "X-Custom, X-Other" {
		t.Errorf("unexpected Expose-Headers: %q", ctx.respHeaders["Access-Control-Expose-Headers"])
	}
}
