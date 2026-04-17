package security

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
	host        string
	respCode    int
	respBody    any
	store       map[string]any
	headers     map[string]string
	queries     map[string]string
	cookies     map[string]*http.Cookie
	respHeaders map[string]string
}

func newMockContext() *mockContext {
	return &mockContext{
		method:      http.MethodGet,
		path:        "/test",
		host:        "example.com",
		store:       make(map[string]any),
		headers:     make(map[string]string),
		queries:     make(map[string]string),
		cookies:     make(map[string]*http.Cookie),
		respHeaders: make(map[string]string),
	}
}

func (m *mockContext) Request() *http.Request {
	req, _ := http.NewRequest(m.method, "https://"+m.host+m.path, nil)
	req.Host = m.host
	q := req.URL.Query()
	for k, v := range m.queries {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()
	for k, v := range m.headers {
		req.Header.Set(k, v)
	}
	return req
}
func (m *mockContext) Param(string) string                      { return "" }
func (m *mockContext) Query(key string) string                  { return m.queries[key] }
func (m *mockContext) QueryDefault(key, def string) string      { if v, ok := m.queries[key]; ok { return v }; return def }
func (m *mockContext) Header(key string) string                 { return m.headers[key] }
func (m *mockContext) Cookie(name string) (*http.Cookie, error) { if c, ok := m.cookies[name]; ok { return c, nil }; return nil, http.ErrNoCookie }
func (m *mockContext) ClientIP() string                         { return "127.0.0.1" }
func (m *mockContext) Method() string                           { return m.method }
func (m *mockContext) Path() string                             { return m.path }
func (m *mockContext) Bind(any) error                           { return nil }
func (m *mockContext) BindJSON(any) error                       { return nil }
func (m *mockContext) BindQuery(any) error                      { return nil }
func (m *mockContext) BindForm(any) error                       { return nil }
func (m *mockContext) JSON(code int, v any) error               { m.respCode = code; m.respBody = v; return nil }
func (m *mockContext) String(code int, s string) error          { m.respCode = code; m.respBody = s; return nil }
func (m *mockContext) HTML(code int, html string) error         { m.respCode = code; m.respBody = html; return nil }
func (m *mockContext) Blob(code int, _ string, _ []byte) error  { m.respCode = code; return nil }
func (m *mockContext) NoContent(code int) error                 { m.respCode = code; return nil }
func (m *mockContext) Redirect(code int, _ string) error        { m.respCode = code; return nil }
func (m *mockContext) SetHeader(key, value string)              { m.respHeaders[key] = value }
func (m *mockContext) SetCookie(c *http.Cookie)                 { m.cookies[c.Name] = c }
func (m *mockContext) Status(code int)                          { m.respCode = code }
func (m *mockContext) Success(data any) error                   { return m.JSON(http.StatusOK, data) }
func (m *mockContext) Fail(msg string) error                    { return m.JSON(http.StatusOK, msg) }
func (m *mockContext) BadRequest(msg ...string) error           { return m.JSON(400, getMsg(msg, "Bad Request")) }
func (m *mockContext) Unauthorized(msg ...string) error         { return m.JSON(401, getMsg(msg, "Unauthorized")) }
func (m *mockContext) Forbidden(msg ...string) error            { return m.JSON(403, getMsg(msg, "Forbidden")) }
func (m *mockContext) NotFound(msg ...string) error             { return m.JSON(404, getMsg(msg, "Not Found")) }
func (m *mockContext) TooManyRequests(msg ...string) error      { return m.JSON(429, getMsg(msg, "Too Many Requests")) }
func (m *mockContext) InternalError(msg ...string) error        { return m.JSON(500, getMsg(msg, "Internal Server Error")) }
func (m *mockContext) ServiceUnavailable(msg ...string) error   { return m.JSON(503, getMsg(msg, "Service Unavailable")) }
func (m *mockContext) Set(key string, value any)                { m.store[key] = value }
func (m *mockContext) Get(key string) (any, bool)               { v, ok := m.store[key]; return v, ok }
func (m *mockContext) MustGet(key string) any                   { return m.store[key] }
func (m *mockContext) ResponseWriter() http.ResponseWriter      { return nil }
func (m *mockContext) Raw() any                                 { return nil }

func TestSecurity_DefaultHeadersAreSet(t *testing.T) {
	middleware := New()

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	ctx := newMockContext()
	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got := ctx.respHeaders["X-Content-Type-Options"]; got != "nosniff" {
		t.Fatalf("expected nosniff, got %q", got)
	}
	if got := ctx.respHeaders["X-Frame-Options"]; got != "DENY" {
		t.Fatalf("expected DENY, got %q", got)
	}
	if got := ctx.respHeaders["X-XSS-Protection"]; got != "1; mode=block" {
		t.Fatalf("expected X-XSS-Protection header, got %q", got)
	}
	if got := ctx.respHeaders["Strict-Transport-Security"]; got != "max-age=31536000; includeSubDomains" {
		t.Fatalf("expected HSTS header, got %q", got)
	}
	if got := ctx.respHeaders["Content-Security-Policy"]; got != "default-src 'self'" {
		t.Fatalf("expected default CSP, got %q", got)
	}
}

func TestSecurity_DisallowedHostReturnsBadRequest(t *testing.T) {
	middleware := New(WithAllowedHosts("api.example.com"))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run for blocked host")
		return nil
	})

	ctx := newMockContext()
	ctx.host = "evil.example.com"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.respCode)
	}
}

func TestSecurity_XSSPatternReturnsBadRequest(t *testing.T) {
	middleware := New(WithXSSProtection(true))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run for XSS payload")
		return nil
	})

	ctx := newMockContext()
	ctx.queries["q"] = "<script>alert(1)</script>"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.respCode)
	}
}

func TestSecurity_SQLPatternReturnsBadRequest(t *testing.T) {
	middleware := New(WithSQLInjectionProtection(true))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run for SQL injection payload")
		return nil
	})

	ctx := newMockContext()
	ctx.queries["q"] = "' OR '1'='1"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.respCode)
	}
}

func TestSecurity_CSRFGetRequestSetsTokenCookie(t *testing.T) {
	middleware := New(WithCSRFProtection(CSRFConfig{TokenLookup: "header:X-CSRF-Token"}))

	handler := middleware(func(ctx httpx.Context) error {
		token := GetCSRFToken(ctx)
		if token == "" {
			t.Fatal("expected csrf token in context")
		}
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	ctx := newMockContext()
	ctx.method = http.MethodGet

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	cookie, ok := ctx.cookies["_csrf"]
	if !ok {
		t.Fatal("expected CSRF cookie to be set")
	}
	if cookie.Value == "" {
		t.Fatal("expected non-empty CSRF cookie value")
	}
}

func TestSecurity_CSRFPostRequestRequiresMatchingToken(t *testing.T) {
	middleware := New(WithCSRFProtection(CSRFConfig{TokenLookup: "header:X-CSRF-Token"}))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run when CSRF token is invalid")
		return nil
	})

	ctx := newMockContext()
	ctx.method = http.MethodPost
	ctx.cookies["_csrf"] = &http.Cookie{Name: "_csrf", Value: "expected-token"}
	ctx.headers["X-CSRF-Token"] = "wrong-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.respCode)
	}
}
