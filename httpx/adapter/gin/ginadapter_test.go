package gin_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/f2xme/gox/httpx"
	ginadapter "github.com/f2xme/gox/httpx/adapter/gin"
	ginframework "github.com/gin-gonic/gin"
)

func init() {
	ginframework.SetMode(ginframework.TestMode)
}

func doRequest(e httpx.Engine, method, path string, body ...string) *httptest.ResponseRecorder {
	var req *http.Request
	if len(body) > 0 {
		req = httptest.NewRequest(method, path, strings.NewReader(body[0]))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	e.Raw().(*ginframework.Engine).ServeHTTP(w, req)
	return w
}

func decodeResponse(t *testing.T, w *httptest.ResponseRecorder) httpx.Response {
	t.Helper()
	var resp httpx.Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp
}

func TestGET_String(t *testing.T) {
	e := ginadapter.New()
	e.GET("/hello", func(ctx httpx.Context) error {
		return ctx.String(200, "hello world")
	})

	w := doRequest(e, "GET", "/hello")
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "hello world" {
		t.Errorf("expected 'hello world', got %q", w.Body.String())
	}
}

func TestGET_JSON(t *testing.T) {
	e := ginadapter.New()
	e.GET("/data", func(ctx httpx.Context) error {
		return ctx.JSON(200, map[string]string{"key": "value"})
	})

	w := doRequest(e, "GET", "/data")
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if result["key"] != "value" {
		t.Errorf("expected key=value, got %v", result)
	}
}

func TestSuccess(t *testing.T) {
	e := ginadapter.New()
	e.GET("/ok", func(ctx httpx.Context) error {
		return ctx.Success([]string{"a", "b"})
	})

	w := doRequest(e, "GET", "/ok")
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Message != "ok" {
		t.Errorf("expected message='ok', got %q", resp.Message)
	}
}

func TestFail(t *testing.T) {
	e := ginadapter.New()
	e.GET("/fail", func(ctx httpx.Context) error {
		return ctx.Fail("something wrong")
	})

	w := doRequest(e, "GET", "/fail")
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Message != "something wrong" {
		t.Errorf("expected message='something wrong', got %q", resp.Message)
	}
}

func TestErrorHandler_HTTPError(t *testing.T) {
	e := ginadapter.New()
	e.GET("/err", func(ctx httpx.Context) error {
		return httpx.NewHTTPError(http.StatusBadRequest, "invalid input")
	})

	w := doRequest(e, "GET", "/err")
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Message != "invalid input" {
		t.Errorf("expected message='invalid input', got %q", resp.Message)
	}
}

func TestErrorHandler_GenericError(t *testing.T) {
	e := ginadapter.New()
	e.GET("/err", func(ctx httpx.Context) error {
		return fmt.Errorf("unexpected failure")
	})

	w := doRequest(e, "GET", "/err")
	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGroup(t *testing.T) {
	e := ginadapter.New()
	api := e.Group("/api")
	api.GET("/users", func(ctx httpx.Context) error {
		return ctx.String(200, "users")
	})

	w := doRequest(e, "GET", "/api/users")
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "users" {
		t.Errorf("expected 'users', got %q", w.Body.String())
	}
}

func TestNestedGroup(t *testing.T) {
	e := ginadapter.New()
	api := e.Group("/api")
	v1 := api.Group("/v1")
	v1.GET("/users", func(ctx httpx.Context) error {
		return ctx.String(200, "v1-users")
	})

	w := doRequest(e, "GET", "/api/v1/users")
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "v1-users" {
		t.Errorf("expected 'v1-users', got %q", w.Body.String())
	}
}

func TestMiddleware_ExecutionOrder(t *testing.T) {
	e := ginadapter.New()

	callOrder := []string{}

	mw1 := func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			callOrder = append(callOrder, "mw1-before")
			err := next(ctx)
			callOrder = append(callOrder, "mw1-after")
			return err
		}
	}

	mw2 := func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			callOrder = append(callOrder, "mw2-before")
			err := next(ctx)
			callOrder = append(callOrder, "mw2-after")
			return err
		}
	}

	e.Use(mw1)
	e.GET("/test", func(ctx httpx.Context) error {
		callOrder = append(callOrder, "handler")
		return ctx.String(200, "ok")
	}, mw2)

	doRequest(e, "GET", "/test")

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(callOrder) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(callOrder), callOrder)
	}
	for i, s := range expected {
		if callOrder[i] != s {
			t.Errorf("call[%d]: expected %q, got %q", i, s, callOrder[i])
		}
	}
}

func TestParam(t *testing.T) {
	e := ginadapter.New()
	e.GET("/users/:id", func(ctx httpx.Context) error {
		return ctx.String(200, ctx.Param("id").String())
	})

	w := doRequest(e, "GET", "/users/42")
	if w.Body.String() != "42" {
		t.Errorf("expected '42', got %q", w.Body.String())
	}
}

func TestParamInt64(t *testing.T) {
	e := ginadapter.New()
	e.GET("/items/:id", func(ctx httpx.Context) error {
		id, err := ctx.Param("id").Int64()
		if err != nil {
			return ctx.BadRequest("invalid id")
		}
		return ctx.JSON(200, map[string]int64{"id": id})
	})

	w := doRequest(e, "GET", "/items/123")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"id":123`) {
		t.Errorf("expected id=123 in body, got %q", w.Body.String())
	}
}

func TestQuery(t *testing.T) {
	e := ginadapter.New()
	e.GET("/search", func(ctx httpx.Context) error {
		q := ctx.Query("q").String()
		page := ctx.Query("page").Or("1")
		return ctx.String(200, q+":"+page)
	})

	w := doRequest(e, "GET", "/search?q=hello")
	if w.Body.String() != "hello:1" {
		t.Errorf("expected 'hello:1', got %q", w.Body.String())
	}
}

func TestQueryTyped(t *testing.T) {
	e := ginadapter.New()
	e.GET("/list", func(ctx httpx.Context) error {
		page := ctx.Query("page").IntOr(1)
		enabled := ctx.Query("enabled").BoolOr(false)
		return ctx.JSON(200, map[string]any{"page": page, "enabled": enabled})
	})

	w := doRequest(e, "GET", "/list?page=3&enabled=true")
	if !strings.Contains(w.Body.String(), `"page":3`) || !strings.Contains(w.Body.String(), `"enabled":true`) {
		t.Errorf("unexpected body: %q", w.Body.String())
	}
}

func TestQueryAll(t *testing.T) {
	e := ginadapter.New()
	e.GET("/tags", func(ctx httpx.Context) error {
		return ctx.JSON(200, map[string][]string{"tags": ctx.QueryAll("tag")})
	})

	w := doRequest(e, "GET", "/tags?tag=a&tag=b&tag=c")
	if !strings.Contains(w.Body.String(), `"a","b","c"`) {
		t.Errorf("expected multi-value tags, got %q", w.Body.String())
	}
}

func TestContextValues(t *testing.T) {
	e := ginadapter.New()
	e.GET("/ctx", func(ctx httpx.Context) error {
		ctx.Set("user", "alice")
		v, ok := ctx.Get("user")
		if !ok {
			return ctx.String(500, "not found")
		}
		return ctx.String(200, v.(string))
	})

	w := doRequest(e, "GET", "/ctx")
	if w.Body.String() != "alice" {
		t.Errorf("expected 'alice', got %q", w.Body.String())
	}
}

func TestBindJSON(t *testing.T) {
	e := ginadapter.New()
	type Req struct {
		Name string `json:"name"`
	}
	e.POST("/bind", func(ctx httpx.Context) error {
		var req Req
		if err := ctx.BindJSON(&req); err != nil {
			return err
		}
		return ctx.Success(req.Name)
	})

	w := doRequest(e, "POST", "/bind", `{"name":"bob"}`)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestGroupMiddleware(t *testing.T) {
	e := ginadapter.New()

	authMw := func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			if ctx.Header("Authorization") == "" {
				return httpx.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}
			return next(ctx)
		}
	}

	api := e.Group("/api", authMw)
	api.GET("/secret", func(ctx httpx.Context) error {
		return ctx.String(200, "secret")
	})

	// Without auth header
	w := doRequest(e, "GET", "/api/secret")
	if w.Code != 401 {
		t.Errorf("expected 401, got %d", w.Code)
	}

	// With auth header
	req := httptest.NewRequest("GET", "/api/secret", nil)
	req.Header.Set("Authorization", "Bearer token")
	w2 := httptest.NewRecorder()
	e.Raw().(*ginframework.Engine).ServeHTTP(w2, req)
	if w2.Code != 200 {
		t.Errorf("expected 200, got %d", w2.Code)
	}
}

func TestCustomErrorHandler(t *testing.T) {
	e := ginadapter.New()
	e.SetErrorHandler(func(ctx httpx.Context, err error) {
		_ = ctx.JSON(418, httpx.NewFailResponse("custom: "+err.Error()))
	})
	e.GET("/err", func(ctx httpx.Context) error {
		return fmt.Errorf("oops")
	})

	w := doRequest(e, "GET", "/err")
	if w.Code != 418 {
		t.Errorf("expected 418, got %d", w.Code)
	}
}
