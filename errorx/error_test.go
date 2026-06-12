package errorx

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("test error")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got %q", err.Error())
	}
	if err.Kind != KindUnknown {
		t.Errorf("expected KindUnknown, got %v", err.Kind)
	}
}

func TestNewWithKind(t *testing.T) {
	err := NewWithKind(KindValidation, "validation failed")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Error() != "validation failed" {
		t.Errorf("expected 'validation failed', got %q", err.Error())
	}
	if err.Kind != KindValidation {
		t.Errorf("expected KindValidation, got %v", err.Kind)
	}
}

func TestNewCode(t *testing.T) {
	err := NewCode("ERR001", "test error")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Code != "ERR001" {
		t.Errorf("expected 'ERR001', got %q", err.Code)
	}
	if err.Message != "test error" {
		t.Errorf("expected 'test error', got %q", err.Message)
	}
}

func TestWrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, "wrapped")

	if wrapped == nil {
		t.Fatal("expected non-nil error")
	}
	if wrapped.Error() != "wrapped: original error" {
		t.Errorf("expected 'wrapped: original error', got %q", wrapped.Error())
	}
	if wrapped.Cause != original {
		t.Error("expected Cause to be original error")
	}

	// Test unwrap
	if errors.Unwrap(wrapped) != original {
		t.Error("Unwrap should return original error")
	}
}

func TestWrapNil(t *testing.T) {
	wrapped := Wrap(nil, "message")
	if wrapped != nil {
		t.Error("wrapping nil should return nil")
	}
}

func TestFromNil(t *testing.T) {
	err := From(nil)
	if err != nil {
		t.Error("from nil should return nil")
	}
}

func TestFromError(t *testing.T) {
	original := errors.New("standard error")
	err := From(original)

	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Message != "standard error" {
		t.Errorf("expected 'standard error', got %q", err.Message)
	}
	if err.Kind != KindUnknown {
		t.Errorf("expected KindUnknown, got %v", err.Kind)
	}
	if err.Cause != original {
		t.Error("expected Cause to be original error")
	}
	if errors.Unwrap(err) != original {
		t.Error("Unwrap should return original error")
	}
	if len(err.Stack) == 0 {
		t.Error("expected non-empty stack")
	}
}

func TestFromErrorxError(t *testing.T) {
	original := New("errorx error")
	err := From(original)
	if err != original {
		t.Error("expected From to return original *Error")
	}
}

func TestErrorWithKind(t *testing.T) {
	err := New("test").WithKind(KindNotFound)
	if err.Kind != KindNotFound {
		t.Errorf("expected KindNotFound, got %v", err.Kind)
	}
}

func TestErrorWithMetadata(t *testing.T) {
	err := New("test").WithMetadata("key", "value")
	if err.Metadata == nil {
		t.Fatal("expected non-nil Metadata")
	}
	if err.Metadata["key"] != "value" {
		t.Errorf("expected 'value', got %v", err.Metadata["key"])
	}
}

func TestErrorWithCode(t *testing.T) {
	err := New("test").WithCode("ERR002")
	if err.Code != "ERR002" {
		t.Errorf("expected 'ERR002', got %q", err.Code)
	}
}

func TestErrorChaining(t *testing.T) {
	err := New("test").
		WithKind(KindValidation).
		WithCode("VAL001").
		WithMetadata("field", "email")

	if err.Kind != KindValidation {
		t.Errorf("expected KindValidation, got %v", err.Kind)
	}
	if err.Code != "VAL001" {
		t.Errorf("expected 'VAL001', got %q", err.Code)
	}
	if err.Metadata["field"] != "email" {
		t.Errorf("expected 'email', got %v", err.Metadata["field"])
	}
}

func TestErrorFormat(t *testing.T) {
	err := NewCode("ERR001", "test error")

	// Test %s format
	s := fmt.Sprintf("%s", err)
	if s != "[ERR001] test error" {
		t.Errorf("expected '[ERR001] test error', got %q", s)
	}

	// Test %v format
	v := fmt.Sprintf("%v", err)
	if v != "[ERR001] test error" {
		t.Errorf("expected '[ERR001] test error', got %q", v)
	}
}

func TestErrorFormatWithCause(t *testing.T) {
	original := errors.New("original")
	wrapped := Wrap(original, "wrapped")

	s := wrapped.Error()
	if s != "wrapped: original" {
		t.Errorf("expected 'wrapped: original', got %q", s)
	}
}

func TestIs(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "errorx.Error",
			err:  New("test error"),
			want: true,
		},
		{
			name: "errorx.Error with code",
			err:  NewCode("ERR001", "test error"),
			want: true,
		},
		{
			name: "wrapped errorx.Error",
			err:  Wrap(New("inner"), "outer"),
			want: true,
		},
		{
			name: "standard error",
			err:  errors.New("standard error"),
			want: false,
		},
		{
			name: "fmt.Errorf",
			err:  fmt.Errorf("formatted error"),
			want: false,
		},
		{
			name: "wrapped standard error",
			err:  fmt.Errorf("wrap: %w", errors.New("original")),
			want: false,
		},
		{
			name: "errorx.Error wrapped in standard error",
			err:  fmt.Errorf("wrap: %w", New("errorx error")),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Is(tt.err)
			if got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}
