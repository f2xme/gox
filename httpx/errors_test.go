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
