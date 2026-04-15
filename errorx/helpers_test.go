package errorx

import (
	"errors"
	"testing"
)

func TestIsKind(t *testing.T) {
	err := NewWithKind(KindValidation, "validation error")

	if !IsKind(err, KindValidation) {
		t.Error("expected IsKind to return true for KindValidation")
	}

	if IsKind(err, KindNotFound) {
		t.Error("expected IsKind to return false for KindNotFound")
	}

	// Test with non-Error type
	stdErr := errors.New("standard error")
	if IsKind(stdErr, KindValidation) {
		t.Error("expected IsKind to return false for standard error")
	}
}

func TestIsRetryable(t *testing.T) {
	retryableErr := NewWithKind(KindRetryable, "retryable error")
	if !IsRetryable(retryableErr) {
		t.Error("expected IsRetryable to return true")
	}

	timeoutErr := NewWithKind(KindTimeout, "timeout error")
	if !IsRetryable(timeoutErr) {
		t.Error("expected IsRetryable to return true for timeout")
	}

	validationErr := NewWithKind(KindValidation, "validation error")
	if IsRetryable(validationErr) {
		t.Error("expected IsRetryable to return false for validation error")
	}

	// Test with non-Error type
	stdErr := errors.New("standard error")
	if IsRetryable(stdErr) {
		t.Error("expected IsRetryable to return false for standard error")
	}
}

func TestIsTimeout(t *testing.T) {
	timeoutErr := NewWithKind(KindTimeout, "timeout error")
	if !IsTimeout(timeoutErr) {
		t.Error("expected IsTimeout to return true")
	}

	validationErr := NewWithKind(KindValidation, "validation error")
	if IsTimeout(validationErr) {
		t.Error("expected IsTimeout to return false")
	}
}

func TestGetCode(t *testing.T) {
	err := NewCode("ERR001", "test error")
	code := GetCode(err)
	if code != "ERR001" {
		t.Errorf("expected 'ERR001', got %q", code)
	}

	// Test with non-Error type
	stdErr := errors.New("standard error")
	code = GetCode(stdErr)
	if code != "" {
		t.Errorf("expected empty string, got %q", code)
	}
}

func TestGetStack(t *testing.T) {
	err := New("test error")
	stack := GetStack(err)
	if len(stack) == 0 {
		t.Error("expected non-empty stack")
	}

	// Test with non-Error type
	stdErr := errors.New("standard error")
	stack = GetStack(stdErr)
	if len(stack) != 0 {
		t.Error("expected empty stack for standard error")
	}
}

func TestFormat(t *testing.T) {
	err := NewCode("ERR001", "test error").
		WithKind(KindValidation).
		WithMetadata("field", "email")

	formatted := Format(err)
	if formatted == "" {
		t.Error("expected non-empty formatted string")
	}

	// Should contain code
	if len(formatted) < 10 {
		t.Error("expected formatted string to contain details")
	}
}

func TestFormatWithStack(t *testing.T) {
	err := New("test error")
	formatted := FormatWithStack(err)

	if formatted == "" {
		t.Error("expected non-empty formatted string")
	}

	// Should contain stack trace
	if len(formatted) < 50 {
		t.Error("expected formatted string to contain stack trace")
	}
}

func TestFormatNonError(t *testing.T) {
	stdErr := errors.New("standard error")
	formatted := Format(stdErr)
	if formatted != "standard error" {
		t.Errorf("expected 'standard error', got %q", formatted)
	}
}
