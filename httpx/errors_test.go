package httpx

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestHTTPError_Error(t *testing.T) {
	e := NewHTTPError(http.StatusBadRequest, "bad request")
	expected := "code=400, message=bad request"
	if e.Error() != expected {
		t.Errorf("expected %q, got %q", expected, e.Error())
	}
}

func TestHTTPError_WithError(t *testing.T) {
	inner := fmt.Errorf("inner error")
	e := NewHTTPError(http.StatusBadRequest, "bad request").WithError(inner)
	if !errors.Is(e, inner) {
		t.Error("expected errors.Is to find inner error")
	}
	if e.Unwrap() != inner {
		t.Error("expected Unwrap to return inner error")
	}
}

func TestHTTPError_ErrorsAs(t *testing.T) {
	e := NewHTTPError(http.StatusNotFound, "not found")
	wrapped := fmt.Errorf("wrapped: %w", e)

	var he *HTTPError
	if !errors.As(wrapped, &he) {
		t.Error("expected errors.As to find HTTPError")
	}
	if he.Code != http.StatusNotFound {
		t.Errorf("expected code 404, got %d", he.Code)
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
		if e.Code != http.StatusBadRequest {
			t.Errorf("code: want 400, got %d", e.Code)
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
		if e.Code != http.StatusNotFound {
			t.Errorf("code: want 404, got %d", e.Code)
		}
		if !errors.Is(e, inner) {
			t.Error("expected errors.Is to find inner error")
		}
	})
}

// TestNewBizError 验证 BizError 工厂的三种调用方式。
func TestNewBizError(t *testing.T) {
	inner := errors.New("db error")

	t.Run("仅传 string", func(t *testing.T) {
		e := NewBizError("余额不足")
		if e.Message != "余额不足" {
			t.Errorf("message: want '余额不足', got %q", e.Message)
		}
		if e.Err != nil {
			t.Errorf("err: want nil, got %v", e.Err)
		}
	})

	t.Run("仅传 error", func(t *testing.T) {
		e := NewBizError(inner)
		if e.Message != inner.Error() {
			t.Errorf("message: want %q, got %q", inner.Error(), e.Message)
		}
		if !errors.Is(e, inner) {
			t.Error("expected errors.Is to find inner error")
		}
	})

	t.Run("string + error", func(t *testing.T) {
		e := NewBizError("业务校验失败", inner)
		if e.Message != "业务校验失败" {
			t.Errorf("message: want '业务校验失败', got %q", e.Message)
		}
		if !errors.Is(e, inner) {
			t.Error("expected errors.Is to find inner error")
		}
	})
}
