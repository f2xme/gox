package requestid

import (
	"testing"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/mock"
)

func TestNew_GeneratesID(t *testing.T) {
	mw := New()
	var capturedID string
	handler := mw(func(ctx httpx.Context) error {
		v, _ := ctx.Get("request_id")
		capturedID = v.(string)
		return nil
	})
	ctx := mock.NewMockContext("GET", "/test")
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if capturedID == "" {
		t.Error("expected request_id to be set")
	}
	if ctx.RespHeaders.Get(defaultHeaderKey) == "" {
		t.Error("expected X-Request-ID response header")
	}
	if ctx.RespHeaders.Get(defaultHeaderKey) != capturedID {
		t.Error("header and context value should match")
	}
}

func TestNew_PreservesExisting(t *testing.T) {
	mw := New()
	handler := mw(func(ctx httpx.Context) error { return nil })
	ctx := mock.NewMockContext("GET", "/test")
	ctx.Headers[defaultHeaderKey] = "existing-id-123"
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.RespHeaders.Get(defaultHeaderKey) != "existing-id-123" {
		t.Errorf("expected preserved ID, got %q", ctx.RespHeaders.Get(defaultHeaderKey))
	}
}

func TestNew_Unique(t *testing.T) {
	mw := New()
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		ctx := mock.NewMockContext("GET", "/test")
		mw(func(ctx httpx.Context) error { return nil })(ctx)
		id := ctx.RespHeaders.Get(defaultHeaderKey)
		if ids[id] {
			t.Fatalf("duplicate request ID: %s", id)
		}
		ids[id] = true
	}
}

func TestNew_WithCustomHeaderKey(t *testing.T) {
	mw := New(WithHeaderKey("X-Custom-ID"))
	handler := mw(func(ctx httpx.Context) error { return nil })
	ctx := mock.NewMockContext("GET", "/test")
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.RespHeaders.Get("X-Custom-ID") == "" {
		t.Error("expected X-Custom-ID response header")
	}
	if ctx.RespHeaders.Get(defaultHeaderKey) != "" {
		t.Error("should not set default header when custom key is used")
	}
}

func TestNew_WithGenerator(t *testing.T) {
	counter := 0
	mw := New(WithGenerator(func() string {
		counter++
		return "custom-id"
	}))
	handler := mw(func(ctx httpx.Context) error { return nil })
	ctx := mock.NewMockContext("GET", "/test")
	handler(ctx)
	if ctx.RespHeaders.Get(defaultHeaderKey) != "custom-id" {
		t.Errorf("expected 'custom-id', got %q", ctx.RespHeaders.Get(defaultHeaderKey))
	}
	if counter != 1 {
		t.Errorf("expected generator called once, got %d", counter)
	}
}

func TestNew_WithHandler(t *testing.T) {
	var handlerID string
	mw := New(WithHandler(func(ctx httpx.Context, id string) {
		handlerID = id
	}))
	handler := mw(func(ctx httpx.Context) error { return nil })
	ctx := mock.NewMockContext("GET", "/test")
	handler(ctx)
	if handlerID == "" {
		t.Error("expected handler to be called with ID")
	}
}

func TestGet(t *testing.T) {
	ctx := mock.NewMockContext("GET", "/test")
	ctx.Headers[defaultHeaderKey] = "test-id"
	if Get(ctx) != "test-id" {
		t.Errorf("expected 'test-id', got %q", Get(ctx))
	}
}
