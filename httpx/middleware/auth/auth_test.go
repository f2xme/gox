package auth

import (
	"errors"
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
	queries     map[string]string
	respHeaders map[string]string
}

func newMockContext() *mockContext {
	return &mockContext{
		method:      "GET",
		path:        "/profile",
		store:       make(map[string]any),
		headers:     make(map[string]string),
		queries:     make(map[string]string),
		respHeaders: make(map[string]string),
	}
}

func (m *mockContext) Request() *http.Request  { return nil }
func (m *mockContext) Param(string) string     { return "" }
func (m *mockContext) Query(key string) string { return m.queries[key] }
func (m *mockContext) QueryDefault(key, def string) string {
	if v, ok := m.queries[key]; ok {
		return v
	}
	return def
}
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
func (m *mockContext) Unauthorized(msg ...string) error {
	return m.JSON(401, getMsg(msg, "Unauthorized"))
}
func (m *mockContext) Forbidden(msg ...string) error { return m.JSON(403, getMsg(msg, "Forbidden")) }
func (m *mockContext) NotFound(msg ...string) error  { return m.JSON(404, getMsg(msg, "Not Found")) }
func (m *mockContext) TooManyRequests(msg ...string) error {
	return m.JSON(429, getMsg(msg, "Too Many Requests"))
}
func (m *mockContext) InternalError(msg ...string) error {
	return m.JSON(500, getMsg(msg, "Internal Server Error"))
}
func (m *mockContext) ServiceUnavailable(msg ...string) error {
	return m.JSON(503, getMsg(msg, "Service Unavailable"))
}
func (m *mockContext) Set(key string, value any)           { m.store[key] = value }
func (m *mockContext) Get(key string) (any, bool)          { v, ok := m.store[key]; return v, ok }
func (m *mockContext) MustGet(key string) any              { return m.store[key] }
func (m *mockContext) ResponseWriter() http.ResponseWriter { return nil }
func (m *mockContext) Raw() any                            { return nil }

type mockClaims struct {
	uid    int64
	values map[string]any
}

func (c *mockClaims) GetUID() int64 {
	return c.uid
}

func (c *mockClaims) Get(key string) (any, bool) {
	v, ok := c.values[key]
	return v, ok
}

type mockValidator struct {
	tokenToClaims map[string]Claims
	calledWith    []string
}

func (v *mockValidator) Validate(token string) (Claims, error) {
	v.calledWith = append(v.calledWith, token)
	if claims, ok := v.tokenToClaims[token]; ok {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// mockStatusChecker is a test implementation of UserChecker
type mockStatusChecker struct {
	bannedUsers map[int64]bool
	checkError  error
}

func (m *mockStatusChecker) CheckUser(uid int64) error {
	if m.checkError != nil {
		return m.checkError
	}
	if m.bannedUsers[uid] {
		return errors.New("user is banned")
	}
	return nil
}

func TestAuth_MissingTokenReturnsUnauthorized(t *testing.T) {
	middleware := New()

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("next handler should not be called")
		return nil
	})

	ctx := newMockContext()

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got := ctx.respCode; got != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, got)
	}
}

func TestAuth_ValidBearerTokenStoresClaimsAndCallsNext(t *testing.T) {
	validator := &mockValidator{
		tokenToClaims: map[string]Claims{
			"good-token": &mockClaims{uid: 123, values: map[string]any{"role": "admin"}},
		},
	}

	middleware := New(WithValidator(validator))

	handler := middleware(func(ctx httpx.Context) error {
		claims := GetClaims(ctx)
		if claims == nil {
			t.Fatal("expected claims in context")
		}
		if claims.GetUID() != 123 {
			t.Fatalf("expected uid 123, got %d", claims.GetUID())
		}
		role, ok := claims.Get("role")
		if !ok || role != "admin" {
			t.Fatalf("expected role admin, got %v, %v", role, ok)
		}
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	ctx := newMockContext()
	ctx.headers["Authorization"] = "Bearer good-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.respCode)
	}
}

func TestAuth_InvalidTokenReturnsUnauthorized(t *testing.T) {
	validator := &mockValidator{tokenToClaims: map[string]Claims{}}

	middleware := New(WithValidator(validator))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("next handler should not be called for invalid token")
		return nil
	})

	ctx := newMockContext()
	ctx.headers["Authorization"] = "Bearer bad-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, ctx.respCode)
	}
}

func TestAuth_SkipPathsBypassAuthentication(t *testing.T) {
	validator := &mockValidator{tokenToClaims: map[string]Claims{}}

	middleware := New(WithValidator(validator), WithSkipPaths("/health", "/public/*"))

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	healthCtx := newMockContext()
	healthCtx.path = "/health"
	if err := handler(healthCtx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if healthCtx.respCode != http.StatusOK {
		t.Fatalf("expected /health status %d, got %d", http.StatusOK, healthCtx.respCode)
	}

	publicCtx := newMockContext()
	publicCtx.path = "/public/info"
	if err := handler(publicCtx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if publicCtx.respCode != http.StatusOK {
		t.Fatalf("expected /public/info status %d, got %d", http.StatusOK, publicCtx.respCode)
	}

	if len(validator.calledWith) != 0 {
		t.Fatalf("expected validator not called for skipped paths, got %#v", validator.calledWith)
	}
}

func TestAuth_CustomTokenExtractorReadsQueryToken(t *testing.T) {
	validator := &mockValidator{
		tokenToClaims: map[string]Claims{
			"query-token": &mockClaims{uid: 456, values: map[string]any{}},
		},
	}

	middleware := New(
		WithValidator(validator),
		WithTokenExtractor(func(c httpx.Context) string { return c.Query("token") }),
	)

	handler := middleware(func(ctx httpx.Context) error {
		claims := GetClaims(ctx)
		if claims == nil {
			t.Fatal("expected claims from custom extractor")
		}
		if claims.GetUID() != 456 {
			t.Fatalf("expected uid 456, got %d", claims.GetUID())
		}
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	ctx := newMockContext()
	ctx.queries["token"] = "query-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.respCode)
	}
}

func TestAuth_CustomErrorHandlerOverridesDefaultResponse(t *testing.T) {
	middleware := New(
		WithErrorHandler(func(c httpx.Context) {
			c.Status(http.StatusForbidden)
			_ = c.JSON(http.StatusForbidden, map[string]any{"error": "forbidden"})
		}),
	)

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not be called when auth fails")
		return nil
	})

	ctx := newMockContext()

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, ctx.respCode)
	}
}

func TestAuth_BannedUserReturnsForbidden(t *testing.T) {
	validator := &mockValidator{
		tokenToClaims: map[string]Claims{
			"banned-token": &mockClaims{uid: 1, values: map[string]any{}},
			"normal-token": &mockClaims{uid: 2, values: map[string]any{}},
		},
	}

	statusChecker := &mockStatusChecker{
		bannedUsers: map[int64]bool{1: true},
	}

	middleware := New(
		WithValidator(validator),
		WithUserChecker(statusChecker),
		WithCheckHandler(func(c httpx.Context, _ error) { c.Forbidden("user is banned") }),
	)

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	bannedCtx := newMockContext()
	bannedCtx.headers["Authorization"] = "Bearer banned-token"
	if err := handler(bannedCtx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if bannedCtx.respCode != http.StatusForbidden {
		t.Fatalf("expected banned user status %d, got %d", http.StatusForbidden, bannedCtx.respCode)
	}

	normalCtx := newMockContext()
	normalCtx.headers["Authorization"] = "Bearer normal-token"
	if err := handler(normalCtx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if normalCtx.respCode != http.StatusOK {
		t.Fatalf("expected normal user status %d, got %d", http.StatusOK, normalCtx.respCode)
	}
}

func TestAuth_CustomCheckHandlerOverridesDefaultResponse(t *testing.T) {
	validator := &mockValidator{
		tokenToClaims: map[string]Claims{
			"banned-token": &mockClaims{uid: 1, values: map[string]any{}},
		},
	}

	statusChecker := &mockStatusChecker{
		bannedUsers: map[int64]bool{1: true},
	}

	middleware := New(
		WithValidator(validator),
		WithUserChecker(statusChecker),
		WithCheckHandler(func(c httpx.Context, _ error) {
			c.Status(http.StatusGone)
			_ = c.JSON(http.StatusGone, map[string]any{"error": "account suspended"})
		}),
	)

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not be called for banned user")
		return nil
	})

	ctx := newMockContext()
	ctx.headers["Authorization"] = "Bearer banned-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusGone {
		t.Fatalf("expected status %d, got %d", http.StatusGone, ctx.respCode)
	}
}

func TestAuth_CheckerErrorReturnsForbidden(t *testing.T) {
	validator := &mockValidator{
		tokenToClaims: map[string]Claims{
			"user-token": &mockClaims{uid: 789, values: map[string]any{}},
		},
	}

	statusChecker := &mockStatusChecker{
		checkError: errors.New("database connection failed"),
	}

	middleware := New(
		WithValidator(validator),
		WithUserChecker(statusChecker),
		WithCheckHandler(func(c httpx.Context, _ error) { c.Forbidden("check failed") }),
	)

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	ctx := newMockContext()
	ctx.headers["Authorization"] = "Bearer user-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusForbidden {
		t.Fatalf("expected status %d for checker error (fail closed), got %d", http.StatusForbidden, ctx.respCode)
	}
}

func TestAuth_OptionalPaths(t *testing.T) {
	validator := &mockValidator{
		tokenToClaims: map[string]Claims{
			"valid-token": &mockClaims{uid: 1, values: map[string]any{}},
		},
	}
	middleware := New(
		WithValidator(validator),
		WithOptionalPaths("/feed", "/public/*"),
	)

	// 有 token：注入 claims，继续
	t.Run("with valid token", func(t *testing.T) {
		ctx := newMockContext()
		ctx.path = "/feed"
		ctx.headers["Authorization"] = "Bearer valid-token"
		called := false
		if err := middleware(func(c httpx.Context) error {
			called = true
			if GetUID(c) != 1 {
				t.Fatal("expected uid 1")
			}
			return nil
		})(ctx); err != nil {
			t.Fatal(err)
		}
		if !called {
			t.Fatal("handler should be called")
		}
	})

	// 无 token：放行，claims 为 nil
	t.Run("without token", func(t *testing.T) {
		ctx := newMockContext()
		ctx.path = "/feed"
		called := false
		if err := middleware(func(c httpx.Context) error {
			called = true
			if GetUID(c) != 0 {
				t.Fatal("expected uid 0")
			}
			return nil
		})(ctx); err != nil {
			t.Fatal(err)
		}
		if !called {
			t.Fatal("handler should be called")
		}
	})

	// 无效 token：放行（optional 模式不拦截）
	t.Run("with invalid token", func(t *testing.T) {
		ctx := newMockContext()
		ctx.path = "/public/news"
		ctx.headers["Authorization"] = "Bearer bad-token"
		called := false
		if err := middleware(func(c httpx.Context) error {
			called = true
			return nil
		})(ctx); err != nil {
			t.Fatal(err)
		}
		if !called {
			t.Fatal("handler should be called")
		}
	})
}
