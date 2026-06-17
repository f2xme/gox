package httpx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/f2xme/gox/validator"
)

func TestStatusError_Error(t *testing.T) {
	e := NewStatusError(http.StatusBadRequest, "bad request")
	expected := "status=400, message=bad request"
	if e.Error() != expected {
		t.Errorf("expected %q, got %q", expected, e.Error())
	}
}

func TestStatusError_WithError(t *testing.T) {
	inner := fmt.Errorf("inner error")
	e := NewStatusError(http.StatusBadRequest, "bad request").WithError(inner)
	if !errors.Is(e, inner) {
		t.Error("expected errors.Is to find inner error")
	}
	if e.Unwrap() != inner {
		t.Error("expected Unwrap to return inner error")
	}
}

func TestNewStatusErrorCause(t *testing.T) {
	inner := errors.New("inner error")
	e := NewStatusErrorCause(http.StatusBadGateway, "upstream failed", inner)

	if e.Status != http.StatusBadGateway {
		t.Fatalf("status: want %d, got %d", http.StatusBadGateway, e.Status)
	}
	if e.Message != "upstream failed" {
		t.Fatalf("message: want %q, got %q", "upstream failed", e.Message)
	}
	if !errors.Is(e, inner) {
		t.Fatal("expected errors.Is to find inner error")
	}
}

func TestStatusError_WithErrorDoesNotMutateOriginal(t *testing.T) {
	inner := errors.New("inner error")
	base := NewStatusError(http.StatusBadRequest, "bad request")
	wrapped := base.WithError(inner)

	if base.Err != nil {
		t.Fatalf("base Err: want nil, got %v", base.Err)
	}
	if !errors.Is(wrapped, inner) {
		t.Fatal("expected wrapped error to preserve inner error")
	}
}

func TestStatusError_ErrorsAs(t *testing.T) {
	e := NewStatusError(http.StatusNotFound, "not found")
	wrapped := fmt.Errorf("wrapped: %w", e)

	var he *StatusError
	if !errors.As(wrapped, &he) {
		t.Error("expected errors.As to find StatusError")
	}
	if he.Status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", he.Status)
	}
}

// TestResolveArgs 验证 resolveArgs 的三种调用场景。
func TestResolveArgs(t *testing.T) {
	inner := errors.New("db error")

	tests := []struct {
		name    string
		def     string
		args    []any
		wantMsg string
		wantErr error
	}{
		{
			name:    "无参数使用默认消息",
			def:     "Bad Request",
			args:    nil,
			wantMsg: "Bad Request",
			wantErr: nil,
		},
		{
			name:    "仅传 string 覆盖消息",
			def:     "Bad Request",
			args:    []any{"参数错误"},
			wantMsg: "参数错误",
			wantErr: nil,
		},
		{
			name:    "仅传 error 消息取 err.Error()",
			def:     "Bad Request",
			args:    []any{inner},
			wantMsg: inner.Error(),
			wantErr: inner,
		},
		{
			name:    "string + error 消息取 string，err 保留",
			def:     "Bad Request",
			args:    []any{"参数错误", inner},
			wantMsg: "参数错误",
			wantErr: inner,
		},
		{
			name:    "空字符串不覆盖默认消息",
			def:     "Bad Request",
			args:    []any{""},
			wantMsg: "Bad Request",
			wantErr: nil,
		},
		{
			name:    "nil 忽略",
			def:     "Bad Request",
			args:    []any{nil},
			wantMsg: "Bad Request",
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := resolveArgs(tc.def, tc.args)
			if msg != tc.wantMsg {
				t.Errorf("msg: want %q, got %q", tc.wantMsg, msg)
			}
			if err != tc.wantErr {
				t.Errorf("err: want %v, got %v", tc.wantErr, err)
			}
		})
	}
}

// TestErrFactories 验证工厂函数的三种调用方式。
func TestErrFactories(t *testing.T) {
	inner := errors.New("upstream failure")

	t.Run("无参数返回默认消息", func(t *testing.T) {
		e := ErrBadRequest()
		if e.Status != http.StatusBadRequest {
			t.Errorf("status: want 400, got %d", e.Status)
		}
		if e.Message != "Bad Request" {
			t.Errorf("message: want 'Bad Request', got %q", e.Message)
		}
		if e.Err != nil {
			t.Errorf("err: want nil, got %v", e.Err)
		}
	})

	t.Run("传 string 覆盖消息", func(t *testing.T) {
		e := ErrBadRequest("用户名不能为空")
		if e.Message != "用户名不能为空" {
			t.Errorf("message: want '用户名不能为空', got %q", e.Message)
		}
		if e.Err != nil {
			t.Errorf("err: want nil, got %v", e.Err)
		}
	})

	t.Run("传 error 消息取 err.Error()", func(t *testing.T) {
		e := ErrInternalError(inner)
		if e.Message != inner.Error() {
			t.Errorf("message: want %q, got %q", inner.Error(), e.Message)
		}
		if !errors.Is(e, inner) {
			t.Error("expected errors.Is to find inner error")
		}
	})

	t.Run("string + error 消息取 string，err 保留", func(t *testing.T) {
		e := ErrInternalError("服务异常，请稍后重试", inner)
		if e.Message != "服务异常，请稍后重试" {
			t.Errorf("message: want '服务异常，请稍后重试', got %q", e.Message)
		}
		if !errors.Is(e, inner) {
			t.Error("expected errors.Is to find inner error")
		}
	})

	t.Run("ErrNotFound 仅传 error", func(t *testing.T) {
		e := ErrNotFound(inner)
		if e.Status != http.StatusNotFound {
			t.Errorf("status: want 404, got %d", e.Status)
		}
		if !errors.Is(e, inner) {
			t.Error("expected errors.Is to find inner error")
		}
	})
}

func TestDefaultErrorHandler(t *testing.T) {
	inner := errors.New("sql: connection refused")

	tests := []struct {
		name     string
		err      error
		wantCode int
		wantMsg  string
	}{
		{
			name:     "4xx 返回用户消息",
			err:      ErrBadRequest("参数错误", inner),
			wantCode: http.StatusBadRequest,
			wantMsg:  "参数错误",
		},
		{
			name:     "5xx 不暴露底层错误",
			err:      ErrInternalError(inner),
			wantCode: http.StatusInternalServerError,
			wantMsg:  http.StatusText(http.StatusInternalServerError),
		},
		{
			name:     "普通错误回退为标准 500 消息",
			err:      inner,
			wantCode: http.StatusInternalServerError,
			wantMsg:  http.StatusText(http.StatusInternalServerError),
		},
		{
			name:     "验证错误返回 400",
			err:      fmt.Errorf("bind request: %w", validator.ErrValidation),
			wantCode: http.StatusBadRequest,
			wantMsg:  "bind request: validator: validation failed",
		},
		{
			name:     "非法状态码兜底为 500",
			err:      NewStatusError(0, "bad status"),
			wantCode: http.StatusInternalServerError,
			wantMsg:  http.StatusText(http.StatusInternalServerError),
		},
		{
			name:     "空 4xx 消息回退为状态文本",
			err:      NewStatusError(http.StatusNotFound, ""),
			wantCode: http.StatusNotFound,
			wantMsg:  http.StatusText(http.StatusNotFound),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &errorHandlerContext{}
			DefaultErrorHandler(ctx, tc.err)

			if ctx.status != tc.wantCode {
				t.Fatalf("status: want %d, got %d", tc.wantCode, ctx.status)
			}
			resp, ok := ctx.body.(errorResponse)
			if !ok {
				t.Fatalf("body type: want errorResponse, got %T", ctx.body)
			}
			if resp.Message != tc.wantMsg {
				t.Fatalf("message: want %q, got %q", tc.wantMsg, resp.Message)
			}
		})
	}
}

func TestNormalizeHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		code int
		want int
	}{
		{name: "下界合法", code: 100, want: 100},
		{name: "常用状态码合法", code: http.StatusBadRequest, want: http.StatusBadRequest},
		{name: "上界合法", code: 599, want: 599},
		{name: "低于下界回退 500", code: 99, want: http.StatusInternalServerError},
		{name: "高于上界回退 500", code: 600, want: http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeHTTPStatus(tc.code); got != tc.want {
				t.Fatalf("NormalizeHTTPStatus(%d): want %d, got %d", tc.code, tc.want, got)
			}
		})
	}
}

type errorHandlerContext struct {
	status int
	body   any
}

func (c *errorHandlerContext) Request() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	return req
}

func (c *errorHandlerContext) ReqContext() context.Context { return context.Background() }
func (c *errorHandlerContext) Param(string) Value          { return "" }
func (c *errorHandlerContext) Query(string) Value          { return "" }
func (c *errorHandlerContext) QueryAll(string) []string    { return nil }
func (c *errorHandlerContext) Header(string) Value         { return "" }
func (c *errorHandlerContext) Cookie(string) (*http.Cookie, error) {
	return nil, http.ErrNoCookie
}
func (c *errorHandlerContext) ClientIP() string { return "" }
func (c *errorHandlerContext) Method() string   { return http.MethodGet }
func (c *errorHandlerContext) Path() string     { return "/" }
func (c *errorHandlerContext) Bind(any) error   { return nil }
func (c *errorHandlerContext) BindJSON(any) error {
	return nil
}
func (c *errorHandlerContext) BindQuery(any) error { return nil }
func (c *errorHandlerContext) BindForm(any) error  { return nil }
func (c *errorHandlerContext) JSON(status int, v any) error {
	c.status = status
	c.body = v
	return nil
}
func (c *errorHandlerContext) String(status int, s string) error {
	c.status = status
	c.body = s
	return nil
}
func (c *errorHandlerContext) HTML(status int, html string) error {
	c.status = status
	c.body = html
	return nil
}
func (c *errorHandlerContext) Blob(status int, _ string, data []byte) error {
	c.status = status
	c.body = data
	return nil
}
func (c *errorHandlerContext) NoContent(status int) error {
	c.status = status
	return nil
}
func (c *errorHandlerContext) Redirect(status int, url string) error {
	c.status = status
	c.body = url
	return nil
}
func (c *errorHandlerContext) SetHeader(string, string)            {}
func (c *errorHandlerContext) SetCookie(*http.Cookie)              {}
func (c *errorHandlerContext) Status(status int)                   { c.status = status }
func (c *errorHandlerContext) Set(string, any)                     {}
func (c *errorHandlerContext) Get(string) (any, bool)              { return nil, false }
func (c *errorHandlerContext) MustGet(string) any                  { return nil }
func (c *errorHandlerContext) ResponseWriter() http.ResponseWriter { return nil }
func (c *errorHandlerContext) Raw() any                            { return nil }
