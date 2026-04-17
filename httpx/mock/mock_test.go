package mock

import (
	"net/http"
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
	ctx.QueryParams["q"] = "golang"
	ctx.QueryParams["page"] = "2"

	if ctx.Query("q") != "golang" {
		t.Error("expected q=golang")
	}
	if ctx.QueryDefault("page", "1") != "2" {
		t.Error("expected page=2")
	}
	if ctx.QueryDefault("limit", "10") != "10" {
		t.Error("expected default limit=10")
	}
}

func TestMockContext_PathParams(t *testing.T) {
	ctx := NewMockContext("GET", "/users/:id")
	ctx.PathParams["id"] = "123"

	if ctx.Param("id") != "123" {
		t.Error("expected id=123")
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

func TestMockContext_StatusMethods(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(*MockContext) error
		wantCode int
	}{
		{"BadRequest", func(c *MockContext) error { return c.BadRequest() }, 400},
		{"Unauthorized", func(c *MockContext) error { return c.Unauthorized() }, 401},
		{"Forbidden", func(c *MockContext) error { return c.Forbidden() }, 403},
		{"NotFound", func(c *MockContext) error { return c.NotFound() }, 404},
		{"TooManyRequests", func(c *MockContext) error { return c.TooManyRequests() }, 429},
		{"InternalError", func(c *MockContext) error { return c.InternalError() }, 500},
		{"ServiceUnavailable", func(c *MockContext) error { return c.ServiceUnavailable() }, 503},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewMockContext("GET", "/")
			if err := tt.fn(ctx); err != nil {
				t.Fatal(err)
			}
			if ctx.RespCode != tt.wantCode {
				t.Errorf("expected %d, got %d", tt.wantCode, ctx.RespCode)
			}
		})
	}
}

func TestMockContext_CustomMessage(t *testing.T) {
	ctx := NewMockContext("GET", "/")
	if err := ctx.NotFound("用户不存在"); err != nil {
		t.Fatal(err)
	}

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

	if ctx.RespHeaders["X-Request-ID"] != "abc123" {
		t.Error("expected X-Request-ID header")
	}
	if ctx.RespHeaders["X-Custom"] != "value" {
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
	if ctx.RespHeaders["Location"] != "/new" {
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
	if ctx.RespHeaders["Content-Type"] != "application/octet-stream" {
		t.Error("expected Content-Type header")
	}
}

func TestMockContext_SuccessAndFail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := NewMockContext("GET", "/")
		data := map[string]int{"count": 42}

		if err := ctx.Success(data); err != nil {
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

		if err := ctx.Fail("操作失败"); err != nil {
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
