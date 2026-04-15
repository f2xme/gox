package auth

import (
	"errors"
	"net/http"
	"testing"

	"github.com/f2xme/gox/httpx"
)

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
func (m *mockContext) Set(key string, value any)               { m.store[key] = value }
func (m *mockContext) Get(key string) (any, bool)              { v, ok := m.store[key]; return v, ok }
func (m *mockContext) MustGet(key string) any                  { return m.store[key] }
func (m *mockContext) ResponseWriter() http.ResponseWriter     { return nil }
func (m *mockContext) Raw() any                                { return nil }

type mockClaims struct {
	subject string
	values  map[string]any
}

func (c *mockClaims) GetSubject() string {
	return c.subject
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
			"good-token": &mockClaims{
				subject: "user-123",
				values: map[string]any{
					"role": "admin",
				},
			},
		},
	}

	middleware := New(WithValidator(validator))

	handler := middleware(func(ctx httpx.Context) error {
		claims := GetClaims(ctx)
		if claims == nil {
			t.Fatal("expected claims in context")
		}
		if claims.GetSubject() != "user-123" {
			t.Fatalf("expected subject user-123, got %q", claims.GetSubject())
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
	if len(validator.calledWith) != 1 || validator.calledWith[0] != "good-token" {
		t.Fatalf("expected validator called with good-token, got %#v", validator.calledWith)
	}
}

func TestAuth_InvalidTokenReturnsUnauthorized(t *testing.T) {
	validator := &mockValidator{
		tokenToClaims: map[string]Claims{},
	}

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
	if len(validator.calledWith) != 1 || validator.calledWith[0] != "bad-token" {
		t.Fatalf("expected validator called with bad-token, got %#v", validator.calledWith)
	}
}

func TestAuth_SkipPathsBypassAuthentication(t *testing.T) {
	validator := &mockValidator{
		tokenToClaims: map[string]Claims{},
	}

	middleware := New(
		WithValidator(validator),
		WithSkipPaths("/health", "/public/*"),
	)

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
			"query-token": &mockClaims{subject: "query-user", values: map[string]any{}},
		},
	}

	middleware := New(
		WithValidator(validator),
		WithTokenExtractor(func(c httpx.Context) string {
			return c.Query("token")
		}),
	)

	handler := middleware(func(ctx httpx.Context) error {
		claims := GetClaims(ctx)
		if claims == nil {
			t.Fatal("expected claims from custom extractor")
		}
		if claims.GetSubject() != "query-user" {
			t.Fatalf("expected subject query-user, got %q", claims.GetSubject())
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
			_ = c.JSON(http.StatusForbidden, map[string]any{
				"error": "forbidden",
			})
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
			"banned-token": &mockClaims{subject: "banned-user", values: map[string]any{}},
			"normal-token": &mockClaims{subject: "normal-user", values: map[string]any{}},
		},
	}

	statusChecker := &mockStatusChecker{
		bannedUsers: map[string]bool{
			"banned-user": true,
		},
	}

	middleware := New(
		WithValidator(validator),
		WithUserStatusChecker(statusChecker),
	)

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	// Test banned user
	bannedCtx := newMockContext()
	bannedCtx.headers["Authorization"] = "Bearer banned-token"

	if err := handler(bannedCtx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if bannedCtx.respCode != http.StatusForbidden {
		t.Fatalf("expected banned user status %d, got %d", http.StatusForbidden, bannedCtx.respCode)
	}

	// Test normal user
	normalCtx := newMockContext()
	normalCtx.headers["Authorization"] = "Bearer normal-token"

	if err := handler(normalCtx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if normalCtx.respCode != http.StatusOK {
		t.Fatalf("expected normal user status %d, got %d", http.StatusOK, normalCtx.respCode)
	}
}

func TestAuth_CustomBanHandlerOverridesDefaultResponse(t *testing.T) {
	validator := &mockValidator{
		tokenToClaims: map[string]Claims{
			"banned-token": &mockClaims{subject: "banned-user", values: map[string]any{}},
		},
	}

	statusChecker := &mockStatusChecker{
		bannedUsers: map[string]bool{
			"banned-user": true,
		},
	}

	middleware := New(
		WithValidator(validator),
		WithUserStatusChecker(statusChecker),
		WithBanHandler(func(c httpx.Context) {
			c.Status(http.StatusGone)
			_ = c.JSON(http.StatusGone, map[string]any{
				"error": "account suspended",
			})
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

// mockStatusChecker is a test implementation of UserStatusChecker
type mockStatusChecker struct {
	bannedUsers map[string]bool
	checkError  error
}

func (m *mockStatusChecker) IsBanned(userID string) (bool, error) {
	if m.checkError != nil {
		return false, m.checkError
	}
	return m.bannedUsers[userID], nil
}

func TestAuth_UserStatusChecker_BannedUserReturnsForbidden(t *testing.T) {
	validator := &mockValidator{
		tokenToClaims: map[string]Claims{
			"user-token": &mockClaims{subject: "user123", values: map[string]any{}},
		},
	}

	statusChecker := &mockStatusChecker{
		bannedUsers: map[string]bool{
			"user123": true,
		},
	}

	middleware := New(
		WithValidator(validator),
		WithUserStatusChecker(statusChecker),
	)

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not be called for banned user")
		return nil
	})

	ctx := newMockContext()
	ctx.headers["Authorization"] = "Bearer user-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, ctx.respCode)
	}
}

func TestAuth_UserStatusChecker_NormalUserPasses(t *testing.T) {
	validator := &mockValidator{
		tokenToClaims: map[string]Claims{
			"user-token": &mockClaims{subject: "user456", values: map[string]any{}},
		},
	}

	statusChecker := &mockStatusChecker{
		bannedUsers: map[string]bool{
			"user123": true, // different user is banned
		},
	}

	middleware := New(
		WithValidator(validator),
		WithUserStatusChecker(statusChecker),
	)

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	ctx := newMockContext()
	ctx.headers["Authorization"] = "Bearer user-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.respCode)
	}
}

func TestAuth_UserStatusChecker_ErrorAllowsAccess(t *testing.T) {
	validator := &mockValidator{
		tokenToClaims: map[string]Claims{
			"user-token": &mockClaims{subject: "user789", values: map[string]any{}},
		},
	}

	statusChecker := &mockStatusChecker{
		checkError: errors.New("database connection failed"),
	}

	middleware := New(
		WithValidator(validator),
		WithUserStatusChecker(statusChecker),
	)

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	ctx := newMockContext()
	ctx.headers["Authorization"] = "Bearer user-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusOK {
		t.Fatalf("expected status %d for checker error (fail open), got %d", http.StatusOK, ctx.respCode)
	}
}
