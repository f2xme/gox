package validator

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/f2xme/gox/httpx"
)

type mockContext struct {
	method      string
	path        string
	body        io.ReadCloser
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
		body:        io.NopCloser(strings.NewReader("")),
		store:       make(map[string]any),
		headers:     make(map[string]string),
		queries:     make(map[string]string),
		cookies:     make(map[string]*http.Cookie),
		respHeaders: make(map[string]string),
	}
}

func (m *mockContext) Request() *http.Request {
	req, _ := http.NewRequest(m.method, "https://example.com"+m.path, m.body)
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
func (m *mockContext) Set(key string, value any)                { m.store[key] = value }
func (m *mockContext) Get(key string) (any, bool)               { v, ok := m.store[key]; return v, ok }
func (m *mockContext) MustGet(key string) any                   { return m.store[key] }
func (m *mockContext) ResponseWriter() http.ResponseWriter      { return nil }
func (m *mockContext) Raw() any                                 { return nil }

func TestValidator_ExceedsMaxBodySize(t *testing.T) {
	middleware := New(WithMaxBodySize(10))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run when body size exceeds limit")
		return nil
	})

	ctx := newMockContext()
	ctx.method = http.MethodPost
	ctx.body = io.NopCloser(strings.NewReader("this is more than 10 bytes"))
	ctx.headers["Content-Length"] = "27"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, ctx.respCode)
	}
}

func TestValidator_DisallowedContentType(t *testing.T) {
	middleware := New(WithAllowedContentTypes("application/json", "application/xml"))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run when content type is not allowed")
		return nil
	})

	ctx := newMockContext()
	ctx.method = http.MethodPost
	ctx.headers["Content-Type"] = "text/plain"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, ctx.respCode)
	}
}

func TestValidator_AllowedContentType(t *testing.T) {
	middleware := New(WithAllowedContentTypes("application/json", "application/xml"))

	handlerCalled := false
	handler := middleware(func(ctx httpx.Context) error {
		handlerCalled = true
		return nil
	})

	ctx := newMockContext()
	ctx.method = http.MethodPost
	ctx.headers["Content-Type"] = "application/json"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !handlerCalled {
		t.Fatal("handler should be called when content type is allowed")
	}
}

func TestValidator_MissingRequiredHeader(t *testing.T) {
	middleware := New(WithRequiredHeaders("X-API-Key", "X-Request-ID"))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run when required header is missing")
		return nil
	})

	ctx := newMockContext()
	ctx.headers["X-API-Key"] = "secret"
	// X-Request-ID is missing

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.respCode)
	}
}

func TestValidator_AllRequiredHeadersPresent(t *testing.T) {
	middleware := New(WithRequiredHeaders("X-API-Key", "X-Request-ID"))

	handlerCalled := false
	handler := middleware(func(ctx httpx.Context) error {
		handlerCalled = true
		return nil
	})

	ctx := newMockContext()
	ctx.headers["X-API-Key"] = "secret"
	ctx.headers["X-Request-ID"] = "12345"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !handlerCalled {
		t.Fatal("handler should be called when all required headers are present")
	}
}

func TestValidator_CustomValidatorFails(t *testing.T) {
	middleware := New(WithCustomValidator(func(ctx httpx.Context) error {
		if ctx.Query("token") == "" {
			return http.ErrMissingFile // any error
		}
		return nil
	}))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run when custom validator fails")
		return nil
	})

	ctx := newMockContext()
	// token query param is missing

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.respCode)
	}
}

func TestValidator_CustomValidatorPasses(t *testing.T) {
	middleware := New(WithCustomValidator(func(ctx httpx.Context) error {
		if ctx.Query("token") == "" {
			return http.ErrMissingFile
		}
		return nil
	}))

	handlerCalled := false
	handler := middleware(func(ctx httpx.Context) error {
		handlerCalled = true
		return nil
	})

	ctx := newMockContext()
	ctx.queries["token"] = "valid-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !handlerCalled {
		t.Fatal("handler should be called when custom validator passes")
	}
}
