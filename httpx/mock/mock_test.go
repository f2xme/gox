package mock

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/f2xme/gox/httpx"
)

func TestNewMockContext(t *testing.T) {
	ctx := NewMockContext("POST", "/api/users")

	if ctx.Method() != "POST" {
		t.Errorf("expected POST, got %s", ctx.Method())
	}
	if ctx.Path() != "/api/users" {
		t.Errorf("expected /api/users, got %s", ctx.Path())
	}
	if ctx.ClientIP() != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1, got %s", ctx.ClientIP())
	}
}

func TestMockContext_QueryParams(t *testing.T) {
	ctx := NewMockContext("GET", "/search")
	ctx.QueryParams["q"] = []string{"golang"}
	ctx.QueryParams["page"] = []string{"2"}

	if ctx.Query("q") != "golang" {
		t.Error("expected q=golang")
	}
	if ctx.Query("page").Or("1") != "2" {
		t.Error("expected page=2")
	}
	if ctx.Query("limit").Or("10") != "10" {
		t.Error("expected default limit=10")
	}
	if got := ctx.Query("page").IntOr(0); got != 2 {
		t.Errorf("expected typed page=2, got %d", got)
	}
}

func TestMockContext_PathParams(t *testing.T) {
	ctx := NewMockContext("GET", "/users/:id")
	ctx.PathParams["id"] = "123"

	if ctx.Param("id") != "123" {
		t.Error("expected id=123")
	}
	id, err := ctx.Param("id").Int64()
	if err != nil || id != 123 {
		t.Errorf("expected typed id=123, got id=%d err=%v", id, err)
	}
}

func TestMockContext_Headers(t *testing.T) {
	ctx := NewMockContext("GET", "/")
	ctx.Headers["Authorization"] = "Bearer token"
	ctx.Headers["Content-Type"] = "application/json"

	if ctx.Header("Authorization") != "Bearer token" {
		t.Error("expected Authorization header")
	}
	if ctx.Header("Content-Type") != "application/json" {
		t.Error("expected Content-Type header")
	}
}

func TestMockContext_JSON(t *testing.T) {
	ctx := NewMockContext("GET", "/")
	data := map[string]string{"status": "ok"}

	if err := ctx.JSON(200, data); err != nil {
		t.Fatal(err)
	}

	if ctx.RespCode != 200 {
		t.Errorf("expected 200, got %d", ctx.RespCode)
	}
	if ctx.RespBody == nil {
		t.Error("expected response body to be set")
	}
}

func TestMockContext_String(t *testing.T) {
	ctx := NewMockContext("GET", "/")

	if err := ctx.String(201, "created"); err != nil {
		t.Fatal(err)
	}

	if ctx.RespCode != 201 {
		t.Errorf("expected 201, got %d", ctx.RespCode)
	}
	if ctx.RespBody != "created" {
		t.Errorf("expected 'created', got %v", ctx.RespBody)
	}
}

func TestHTTPErrors_DefaultMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      *httpx.HTTPError
		wantCode int
	}{
		{"BadRequest", httpx.ErrBadRequest(), 400},
		{"Unauthorized", httpx.ErrUnauthorized(), 401},
		{"Forbidden", httpx.ErrForbidden(), 403},
		{"NotFound", httpx.ErrNotFound(), 404},
		{"TooManyRequests", httpx.ErrTooManyRequests(), 429},
		{"InternalError", httpx.ErrInternalError(), 500},
		{"ServiceUnavailable", httpx.ErrServiceUnavailable(), 503},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewMockContext("GET", "/")
			httpx.DefaultErrorHandler(ctx, tt.err)
			if ctx.RespCode != tt.wantCode {
				t.Errorf("expected %d, got %d", tt.wantCode, ctx.RespCode)
			}
		})
	}
}

func TestHTTPError_CustomMessage(t *testing.T) {
	ctx := NewMockContext("GET", "/")
	httpx.DefaultErrorHandler(ctx, httpx.ErrNotFound("用户不存在"))

	if ctx.RespCode != 404 {
		t.Errorf("expected 404, got %d", ctx.RespCode)
	}

	resp, ok := ctx.RespBody.(*httpx.Response)
	if !ok {
		t.Fatal("expected *httpx.Response")
	}
	if resp.Message != "用户不存在" {
		t.Errorf("expected custom message, got %s", resp.Message)
	}
}

func TestMockContext_Store(t *testing.T) {
	ctx := NewMockContext("GET", "/")
	ctx.Set("user_id", 123)
	ctx.Set("role", "admin")

	if v, ok := ctx.Get("user_id"); !ok || v != 123 {
		t.Error("expected user_id=123")
	}
	if v := ctx.MustGet("role"); v != "admin" {
		t.Error("expected role=admin")
	}
	if _, ok := ctx.Get("missing"); ok {
		t.Error("expected missing key to return false")
	}
}

func TestMockContext_SetHeader(t *testing.T) {
	ctx := NewMockContext("GET", "/")
	ctx.SetHeader("X-Request-ID", "abc123")
	ctx.SetHeader("X-Custom", "value")

	if ctx.RespHeaders.Get("X-Request-ID") != "abc123" {
		t.Error("expected X-Request-ID header")
	}
	if ctx.RespHeaders.Get("X-Custom") != "value" {
		t.Error("expected X-Custom header")
	}
}

func TestMockContext_Redirect(t *testing.T) {
	ctx := NewMockContext("GET", "/old")
	if err := ctx.Redirect(http.StatusMovedPermanently, "/new"); err != nil {
		t.Fatal(err)
	}

	if ctx.RespCode != http.StatusMovedPermanently {
		t.Errorf("expected 301, got %d", ctx.RespCode)
	}
	if ctx.RespHeaders.Get("Location") != "/new" {
		t.Error("expected Location header")
	}
}

func TestMockContext_NoContent(t *testing.T) {
	ctx := NewMockContext("DELETE", "/users/123")
	if err := ctx.NoContent(http.StatusNoContent); err != nil {
		t.Fatal(err)
	}

	if ctx.RespCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", ctx.RespCode)
	}
	if ctx.RespBody != nil {
		t.Error("expected nil body")
	}
}

func TestMockContext_Blob(t *testing.T) {
	ctx := NewMockContext("GET", "/file")
	data := []byte("binary data")

	if err := ctx.Blob(200, "application/octet-stream", data); err != nil {
		t.Fatal(err)
	}

	if ctx.RespCode != 200 {
		t.Errorf("expected 200, got %d", ctx.RespCode)
	}
	if ctx.RespHeaders.Get("Content-Type") != "application/octet-stream" {
		t.Error("expected Content-Type header")
	}
}

func TestSuccessAndFail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := NewMockContext("GET", "/")
		data := map[string]int{"count": 42}

		if err := httpx.Success(ctx, data); err != nil {
			t.Fatal(err)
		}

		if ctx.RespCode != 200 {
			t.Errorf("expected 200, got %d", ctx.RespCode)
		}

		resp, ok := ctx.RespBody.(*httpx.Response)
		if !ok {
			t.Fatal("expected *httpx.Response")
		}
		if !resp.Success {
			t.Error("expected success=true")
		}
	})

	t.Run("Fail", func(t *testing.T) {
		ctx := NewMockContext("GET", "/")

		if err := httpx.Fail(ctx, "操作失败"); err != nil {
			t.Fatal(err)
		}

		if ctx.RespCode != 200 {
			t.Errorf("expected 200, got %d", ctx.RespCode)
		}

		resp, ok := ctx.RespBody.(*httpx.Response)
		if !ok {
			t.Fatal("expected *httpx.Response")
		}
		if resp.Success {
			t.Error("expected success=false")
		}
		if resp.Message != "操作失败" {
			t.Errorf("expected message, got %s", resp.Message)
		}
	})
}

func TestMockContext_Request(t *testing.T) {
	ctx := NewMockContext("POST", "/api/items")
	ctx.HostValue = "api.example.com"
	ctx.Headers["X-Trace-ID"] = "trace-1"
	ctx.QueryParams["q"] = []string{"golang"}
	ctx.Cookies["session"] = &http.Cookie{Name: "session", Value: "s-123"}
	ctx.BodyValue = io.NopCloser(strings.NewReader(`{"name":"x"}`))

	req := ctx.Request()
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if req.Method != "POST" {
		t.Errorf("expected POST, got %s", req.Method)
	}
	if req.Host != "api.example.com" {
		t.Errorf("expected host, got %s", req.Host)
	}
	if req.URL.Query().Get("q") != "golang" {
		t.Errorf("expected query q=golang, got %s", req.URL.Query().Get("q"))
	}
	if req.Header.Get("X-Trace-ID") != "trace-1" {
		t.Errorf("expected header, got %s", req.Header.Get("X-Trace-ID"))
	}
	c, err := req.Cookie("session")
	if err != nil || c.Value != "s-123" {
		t.Errorf("expected cookie session=s-123, got %v err=%v", c, err)
	}

	body, _ := io.ReadAll(req.Body)
	if string(body) != `{"name":"x"}` {
		t.Errorf("expected body, got %q", string(body))
	}
}

func TestMockContext_CookieReadWrite(t *testing.T) {
	ctx := NewMockContext("GET", "/")
	ctx.SetCookie(&http.Cookie{Name: "token", Value: "abc"})

	got, err := ctx.Cookie("token")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Value != "abc" {
		t.Errorf("expected abc, got %s", got.Value)
	}

	if _, err := ctx.Cookie("missing"); err != http.ErrNoCookie {
		t.Errorf("expected ErrNoCookie, got %v", err)
	}
}

func TestMockContext_ResponseWriter(t *testing.T) {
	ctx := NewMockContext("GET", "/")
	ctx.SetHeader("Content-Length", "1024")

	rw := ctx.ResponseWriter()
	if rw == nil {
		t.Fatal("expected non-nil ResponseWriter")
	}
	if got := rw.Header().Get("Content-Length"); got != "1024" {
		t.Errorf("expected Content-Length=1024, got %q", got)
	}

	rw.WriteHeader(http.StatusTeapot)
	if ctx.RespCode != http.StatusTeapot {
		t.Errorf("expected RespCode synced from WriteHeader, got %d", ctx.RespCode)
	}
}
