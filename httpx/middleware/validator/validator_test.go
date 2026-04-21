package validator

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/mock"
)

func TestValidator_ExceedsMaxBodySize(t *testing.T) {
	middleware := New(WithMaxBodySize(10))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run when body size exceeds limit")
		return nil
	})

	ctx := mock.NewMockContext(http.MethodPost, "/test")
	ctx.BodyValue = io.NopCloser(strings.NewReader("this is more than 10 bytes"))
	ctx.Headers["Content-Length"] = "27"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, ctx.RespCode)
	}
}

func TestValidator_DisallowedContentType(t *testing.T) {
	middleware := New(WithAllowedContentTypes("application/json", "application/xml"))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run when content type is not allowed")
		return nil
	})

	ctx := mock.NewMockContext(http.MethodPost, "/test")
	ctx.Headers["Content-Type"] = "text/plain"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, ctx.RespCode)
	}
}

func TestValidator_AllowedContentType(t *testing.T) {
	middleware := New(WithAllowedContentTypes("application/json", "application/xml"))

	handlerCalled := false
	handler := middleware(func(ctx httpx.Context) error {
		handlerCalled = true
		return nil
	})

	ctx := mock.NewMockContext(http.MethodPost, "/test")
	ctx.Headers["Content-Type"] = "application/json"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !handlerCalled {
		t.Fatal("handler should be called when content type is allowed")
	}
}

func TestValidator_MissingRequiredHeader(t *testing.T) {
	middleware := New(WithRequiredHeaders("X-API-Key", "X-Request-ID"))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run when required header is missing")
		return nil
	})

	ctx := mock.NewMockContext(http.MethodGet, "/test")
	ctx.Headers["X-API-Key"] = "secret"
	// X-Request-ID 缺失

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.RespCode)
	}
}

func TestValidator_AllRequiredHeadersPresent(t *testing.T) {
	middleware := New(WithRequiredHeaders("X-API-Key", "X-Request-ID"))

	handlerCalled := false
	handler := middleware(func(ctx httpx.Context) error {
		handlerCalled = true
		return nil
	})

	ctx := mock.NewMockContext(http.MethodGet, "/test")
	ctx.Headers["X-API-Key"] = "secret"
	ctx.Headers["X-Request-ID"] = "12345"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !handlerCalled {
		t.Fatal("handler should be called when all required headers are present")
	}
}

func TestValidator_CustomValidatorFails(t *testing.T) {
	middleware := New(WithCustomValidator(func(ctx httpx.Context) error {
		if ctx.Query("token") == "" {
			return http.ErrMissingFile // any error
		}
		return nil
	}))

	handler := middleware(func(ctx httpx.Context) error {
		t.Fatal("handler should not run when custom validator fails")
		return nil
	})

	ctx := mock.NewMockContext(http.MethodGet, "/test")
	// token query param 缺失

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.RespCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.RespCode)
	}
}

func TestValidator_CustomValidatorPasses(t *testing.T) {
	middleware := New(WithCustomValidator(func(ctx httpx.Context) error {
		if ctx.Query("token") == "" {
			return http.ErrMissingFile
		}
		return nil
	}))

	handlerCalled := false
	handler := middleware(func(ctx httpx.Context) error {
		handlerCalled = true
		return nil
	})

	ctx := mock.NewMockContext(http.MethodGet, "/test")
	ctx.QueryParams["token"] = []string{"valid-token"}

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !handlerCalled {
		t.Fatal("handler should be called when all required headers are present")
	}
}
