package security

import (
	"net/http"
	"testing"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/mock"
)

func newCtx() *mock.MockContext {
	ctx := mock.NewMockContext(http.MethodGet, "/test")
	ctx.HostValue = "example.com"
	return ctx
}

func TestSecurity_DefaultHeadersAreSet(t *testing.T) {
	middleware := New()

	handler := middleware(func(ctx httpx.Context) error {
		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	ctx := newCtx()
	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got := ctx.RespHeaders.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff, got %q", got)
	}
	if got := ctx.RespHeaders.Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("expected DENY, got %q", got)
	}
	if got := ctx.RespHeaders.Get("X-XSS-Protection"); got != "1; mode=block" {
		t.Fatalf("expected X-XSS-Protection header, got %q", got)
	}
	if got := ctx.RespHeaders.Get("Strict-Transport-Security"); got != "max-age=31536000; includeSubDomains" {
		t.Fatalf("expected HSTS header, got %q", got)
	}
	if got := ctx.RespHeaders.Get("Content-Security-Policy"); got != "default-src 'self'" {
		t.Fatalf("expected default CSP, got %q", got)
	}
}

func TestSecurity_DisallowedHostReturnsBadRequest(t *testing.T) {
	middleware := New(WithAllowedHosts("api.example.com"))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run for blocked host")
		return nil
	})

	ctx := newCtx()
	ctx.HostValue = "evil.example.com"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.RespCode)
	}
}

func TestSecurity_XSSPatternReturnsBadRequest(t *testing.T) {
	middleware := New(WithXSSProtection(true))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run for XSS payload")
		return nil
	})

	ctx := newCtx()
	ctx.QueryParams["q"] = []string{"<script>alert(1)</script>"}

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.RespCode)
	}
}

func TestSecurity_SQLPatternReturnsBadRequest(t *testing.T) {
	middleware := New(WithSQLInjectionProtection(true))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run for SQL injection payload")
		return nil
	})

	ctx := newCtx()
	ctx.QueryParams["q"] = []string{"' OR '1'='1"}

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.RespCode)
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

	ctx := newCtx()
	ctx.MethodValue = http.MethodGet

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	cookie, ok := ctx.Cookies["_csrf"]
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

	ctx := newCtx()
	ctx.MethodValue = http.MethodPost
	ctx.Cookies["_csrf"] = &http.Cookie{Name: "_csrf", Value: "expected-token"}
	ctx.Headers["X-CSRF-Token"] = "wrong-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.RespCode)
	}
}
