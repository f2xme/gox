package httpx_test

import (
	"net/http"
	"testing"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/mock"
)

func TestCookie(t *testing.T) {
	ctx := mock.NewMockContext(http.MethodGet, "/")
	ctx.SetCookie(&http.Cookie{Name: "session", Value: "token"})

	if got := httpx.Cookie(ctx, "session"); got != "token" {
		t.Fatalf("Cookie() = %q, want token", got)
	}
	if got := httpx.Cookie(ctx, "missing"); got != "" {
		t.Fatalf("Cookie() missing = %q, want empty", got)
	}
}
