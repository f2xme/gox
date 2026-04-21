package recovery

import (
	"testing"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/mock"
)

func TestNew_NoPanic(t *testing.T) {
	mw := New()
	handler := mw(func(ctx httpx.Context) error {
		return ctx.String(200, "ok")
	})
	ctx := mock.NewMockContext("GET", "/test")
	if err := handler(ctx); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestNew_CatchesPanic(t *testing.T) {
	mw := New()
	handler := mw(func(ctx httpx.Context) error {
		panic("boom")
	})
	ctx := mock.NewMockContext("GET", "/test")
	err := handler(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "boom" {
		t.Errorf("expected 'boom', got %q", err.Error())
	}
}

func TestNew_CatchesErrorPanic(t *testing.T) {
	mw := New()
	expectedErr := httpx.NewHTTPError(500, "internal")
	handler := mw(func(ctx httpx.Context) error {
		panic(expectedErr)
	})
	ctx := mock.NewMockContext("GET", "/test")
	err := handler(ctx)
	if err != expectedErr {
		t.Errorf("expected HTTPError, got %v", err)
	}
}

func TestNew_WithHandler(t *testing.T) {
	var handlerCalled bool
	var handlerErr error
	mw := New(WithHandler(func(ctx httpx.Context, err error) {
		handlerCalled = true
		handlerErr = err
	}))
	handler := mw(func(ctx httpx.Context) error {
		panic("boom")
	})
	ctx := mock.NewMockContext("GET", "/test")
	handler(ctx)
	if !handlerCalled {
		t.Error("expected handler to be called")
	}
	if handlerErr == nil || handlerErr.Error() != "boom" {
		t.Errorf("expected error 'boom', got %v", handlerErr)
	}
}
