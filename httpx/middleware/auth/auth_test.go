package auth

import (
	"errors"
	"net/http"
	"testing"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/mock"
)

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

func newCtx() *mock.MockContext {
	return mock.NewMockContext("GET", "/profile")
}

func TestAuth_MissingTokenReturnsUnauthorized(t *testing.T) {
	middleware := New()

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("next handler should not be called")
		return nil
	})

	ctx := newCtx()

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got := ctx.RespCode; got != http.StatusUnauthorized {
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

	ctx := newCtx()
	ctx.Headers["Authorization"] = "Bearer good-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.RespCode)
	}
}

func TestAuth_InvalidTokenReturnsUnauthorized(t *testing.T) {
	validator := &mockValidator{tokenToClaims: map[string]Claims{}}

	middleware := New(WithValidator(validator))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("next handler should not be called for invalid token")
		return nil
	})

	ctx := newCtx()
	ctx.Headers["Authorization"] = "Bearer bad-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, ctx.RespCode)
	}
}

func TestAuth_SkipPathsBypassAuthentication(t *testing.T) {
	validator := &mockValidator{tokenToClaims: map[string]Claims{}}

	middleware := New(WithValidator(validator), WithSkipPaths("/health", "/public/*"))

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	healthCtx := newCtx()
	healthCtx.PathValue = "/health"
	if err := handler(healthCtx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if healthCtx.RespCode != http.StatusOK {
		t.Fatalf("expected /health status %d, got %d", http.StatusOK, healthCtx.RespCode)
	}

	publicCtx := newCtx()
	publicCtx.PathValue = "/public/info"
	if err := handler(publicCtx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if publicCtx.RespCode != http.StatusOK {
		t.Fatalf("expected /public/info status %d, got %d", http.StatusOK, publicCtx.RespCode)
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
		WithTokenExtractor(func(c httpx.Context) string { return c.Query("token").String() }),
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

	ctx := newCtx()
	ctx.QueryParams["token"] = []string{"query-token"}

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.RespCode)
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

	ctx := newCtx()

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, ctx.RespCode)
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
		WithCheckHandler(func(c httpx.Context, _ error) { _ = c.JSON(http.StatusForbidden, httpx.NewFailResponse("user is banned")) }),
	)

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	bannedCtx := newCtx()
	bannedCtx.Headers["Authorization"] = "Bearer banned-token"
	if err := handler(bannedCtx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if bannedCtx.RespCode != http.StatusForbidden {
		t.Fatalf("expected banned user status %d, got %d", http.StatusForbidden, bannedCtx.RespCode)
	}

	normalCtx := newCtx()
	normalCtx.Headers["Authorization"] = "Bearer normal-token"
	if err := handler(normalCtx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if normalCtx.RespCode != http.StatusOK {
		t.Fatalf("expected normal user status %d, got %d", http.StatusOK, normalCtx.RespCode)
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

	ctx := newCtx()
	ctx.Headers["Authorization"] = "Bearer banned-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusGone {
		t.Fatalf("expected status %d, got %d", http.StatusGone, ctx.RespCode)
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
		WithCheckHandler(func(c httpx.Context, _ error) { _ = c.JSON(http.StatusForbidden, httpx.NewFailResponse("check failed")) }),
	)

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	ctx := newCtx()
	ctx.Headers["Authorization"] = "Bearer user-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusForbidden {
		t.Fatalf("expected status %d for checker error (fail closed), got %d", http.StatusForbidden, ctx.RespCode)
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
		ctx := newCtx()
		ctx.PathValue = "/feed"
		ctx.Headers["Authorization"] = "Bearer valid-token"
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
		ctx := newCtx()
		ctx.PathValue = "/feed"
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
		ctx := newCtx()
		ctx.PathValue = "/public/news"
		ctx.Headers["Authorization"] = "Bearer bad-token"
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
