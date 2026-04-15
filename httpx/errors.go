package httpx

import (
	"errors"
	"fmt"
	"net/http"
)

// HTTPError represents an HTTP error with a status code.
type HTTPError struct {
	Code    int
	Message string
	Err     error
}

// NewHTTPError creates a new HTTPError.
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{Code: code, Message: message}
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d, message=%s, err=%v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

// Unwrap returns the wrapped error.
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// WithError wraps an underlying error.
func (e *HTTPError) WithError(err error) *HTTPError {
	e.Err = err
	return e
}

// DefaultErrorHandler is the default ErrorHandler implementation.
// It maps HTTPError to the appropriate status code, and falls back to 500.
func DefaultErrorHandler(ctx Context, err error) {
	var he *HTTPError
	if errors.As(err, &he) {
		_ = ctx.JSON(he.Code, NewFailResponse(he.Message))
	} else {
		_ = ctx.JSON(http.StatusInternalServerError, NewFailResponse(err.Error()))
	}
}
