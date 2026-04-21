package cors

import (
	"net/http"
	"testing"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/mock"
)

func TestNew_NoOrigin(t *testing.T) {
	mw := New(WithOrigins([]string{"http://example.com"}))
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := mock.NewMockContext("GET", "/test")
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.RespHeaders.Get("Access-Control-Allow-Origin") != "" {
		t.Error("should not set CORS headers without Origin")
	}
}

func TestNew_AllowedOrigin(t *testing.T) {
	mw := New(WithOrigins([]string{"http://example.com"}))
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := mock.NewMockContext("GET", "/test")
	ctx.Headers["Origin"] = "http://example.com"
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.RespHeaders.Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Error("expected Access-Control-Allow-Origin header")
	}
}

func TestNew_DisallowedOrigin(t *testing.T) {
	mw := New(WithOrigins([]string{"http://example.com"}))
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := mock.NewMockContext("GET", "/test")
	ctx.Headers["Origin"] = "http://evil.com"
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.RespHeaders.Get("Access-Control-Allow-Origin") != "" {
		t.Error("should not set CORS headers for disallowed origin")
	}
}

func TestNew_DefaultWildcard(t *testing.T) {
	mw := New()
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := mock.NewMockContext("GET", "/test")
	ctx.Headers["Origin"] = "http://any-site.com"
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.RespHeaders.Get("Access-Control-Allow-Origin") != "http://any-site.com" {
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
	ctx := mock.NewMockContext("OPTIONS", "/test")
	ctx.Headers["Origin"] = "http://example.com"
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if nextCalled {
		t.Error("next should not be called for preflight")
	}
	if ctx.RespCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", ctx.RespCode)
	}
	if ctx.RespHeaders.Get("Access-Control-Allow-Methods") != "GET, POST" {
		t.Errorf("unexpected Allow-Methods: %q", ctx.RespHeaders.Get("Access-Control-Allow-Methods"))
	}
	if ctx.RespHeaders.Get("Access-Control-Max-Age") != "3600" {
		t.Errorf("unexpected Max-Age: %q", ctx.RespHeaders.Get("Access-Control-Max-Age"))
	}
}

func TestNew_WithCredentials(t *testing.T) {
	mw := New(WithCredentials(true))
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := mock.NewMockContext("GET", "/test")
	ctx.Headers["Origin"] = "http://example.com"
	handler(ctx)
	if ctx.RespHeaders.Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("expected Allow-Credentials header")
	}
}

func TestNew_WithExposeHeaders(t *testing.T) {
	mw := New(WithExposeHeaders([]string{"X-Custom", "X-Other"}))
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := mock.NewMockContext("GET", "/test")
	ctx.Headers["Origin"] = "http://example.com"
	handler(ctx)
	if ctx.RespHeaders.Get("Access-Control-Expose-Headers") != "X-Custom, X-Other" {
		t.Errorf("unexpected Expose-Headers: %q", ctx.RespHeaders.Get("Access-Control-Expose-Headers"))
	}
}
