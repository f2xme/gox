package validator

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/mock"
)

// bodyContext 保留同一个底层请求，使中间件对 Body 的替换对后续读取可见。
type bodyContext struct {
	*mock.MockContext
	req *http.Request
}

func (c *bodyContext) Request() *http.Request { return c.req }

func TestValidator_BodyReadLimit(t *testing.T) {
	for _, tt := range []struct {
		name          string
		body          string
		limit         int64
		contentLength int64
		custom        bool
		wantStatus    int
		wantBytes     int
	}{
		{name: "below limit", body: "1234", limit: 5, contentLength: -1, wantBytes: 4},
		{name: "at limit", body: "12345", limit: 5, contentLength: -1, wantBytes: 5},
		{name: "chunked over limit", body: "123456", limit: 5, contentLength: -1, wantStatus: 413, wantBytes: 5},
		{name: "declared over limit", body: "123456", limit: 5, contentLength: 6, wantStatus: 413},
		{name: "custom validator over limit", body: "123456", limit: 5, contentLength: -1, custom: true, wantStatus: 413, wantBytes: 5},
		{name: "disabled", body: "123456", contentLength: -1, wantBytes: 6},
		{name: "nil body", limit: 5},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(tt.body))
			req.ContentLength = tt.contentLength
			if tt.contentLength == -1 {
				req.TransferEncoding = []string{"chunked"}
			}
			if tt.body == "" {
				req.Body = nil
			}
			ctx := &bodyContext{MockContext: mock.NewMockContext(http.MethodPost, "/test"), req: req}
			readBytes, responses := 0, 0
			readBody := func(ctx httpx.Context) error {
				if ctx.Request().Body == nil {
					return nil
				}
				body, err := io.ReadAll(ctx.Request().Body)
				readBytes = len(body)
				if err != nil {
					var size *http.MaxBytesError
					if !errors.As(err, &size) || size.Limit != tt.limit {
						t.Fatalf("expected MaxBytesError with limit %d, got %v", tt.limit, err)
					}
					return fmt.Errorf("read body: %w", err)
				}
				return nil
			}
			opts := []Option{WithMaxBodySize(tt.limit), WithErrorHandler(func(ctx httpx.Context, code int, message string) {
				responses++
				if message != "Request body too large" {
					t.Errorf("unexpected message: %q", message)
				}
				defaultErrorHandler(ctx, code, message)
			})}
			next := readBody
			if tt.custom {
				opts = append(opts, WithCustomValidator(readBody))
				next = func(httpx.Context) error {
					t.Fatal("handler should not run after custom validation fails")
					return nil
				}
			}
			if err := New(opts...)(next)(ctx); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ctx.RespCode != tt.wantStatus || readBytes != tt.wantBytes {
				t.Fatalf("status/bytes = %d/%d, want %d/%d", ctx.RespCode, readBytes, tt.wantStatus, tt.wantBytes)
			}
			if (tt.wantStatus == 413 && responses != 1) || (tt.wantStatus == 0 && responses != 0) {
				t.Fatalf("unexpected error response count: %d", responses)
			}
		})
	}
}

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
