package oss

import (
	"errors"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "error with key",
			err: &Error{
				Code:    ErrCodeNotFound,
				Message: "object not found",
				Key:     "test/file.txt",
			},
			expected: "oss: NotFound (key=test/file.txt): object not found",
		},
		{
			name: "error without key",
			err: &Error{
				Code:    ErrCodeAccessDenied,
				Message: "access denied",
			},
			expected: "oss: AccessDenied: access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := &Error{
		Code:    ErrCodeInternal,
		Message: "internal error",
		Err:     originalErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestNewError(t *testing.T) {
	err := NewError(ErrCodeNotFound, "not found")
	if err.Code != ErrCodeNotFound {
		t.Errorf("Code = %q, want %q", err.Code, ErrCodeNotFound)
	}
	if err.Message != "not found" {
		t.Errorf("Message = %q, want %q", err.Message, "not found")
	}
}

func TestNewErrorWithKey(t *testing.T) {
	err := NewError(ErrCodeNotFound, "not found", "test/file.txt")
	if err.Code != ErrCodeNotFound {
		t.Errorf("Code = %q, want %q", err.Code, ErrCodeNotFound)
	}
	if err.Message != "not found" {
		t.Errorf("Message = %q, want %q", err.Message, "not found")
	}
	if err.Key != "test/file.txt" {
		t.Errorf("Key = %q, want %q", err.Key, "test/file.txt")
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	err := WrapError(ErrCodeInternal, "wrapped error", originalErr)
	if err.Code != ErrCodeInternal {
		t.Errorf("Code = %q, want %q", err.Code, ErrCodeInternal)
	}
	if err.Message != "wrapped error" {
		t.Errorf("Message = %q, want %q", err.Message, "wrapped error")
	}
	if err.Err != originalErr {
		t.Errorf("Err = %v, want %v", err.Err, originalErr)
	}
}

func TestWrapErrorWithKey(t *testing.T) {
	originalErr := errors.New("original error")
	err := WrapError(ErrCodeInternal, "wrapped error", originalErr, "test/file.txt")
	if err.Code != ErrCodeInternal {
		t.Errorf("Code = %q, want %q", err.Code, ErrCodeInternal)
	}
	if err.Message != "wrapped error" {
		t.Errorf("Message = %q, want %q", err.Message, "wrapped error")
	}
	if err.Key != "test/file.txt" {
		t.Errorf("Key = %q, want %q", err.Key, "test/file.txt")
	}
	if err.Err != originalErr {
		t.Errorf("Err = %v, want %v", err.Err, originalErr)
	}
}
