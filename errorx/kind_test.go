package errorx

import (
	"testing"
)

func TestKindString(t *testing.T) {
	tests := []struct {
		kind     Kind
		expected string
	}{
		{KindUnknown, "Unknown"},
		{KindValidation, "Validation"},
		{KindNotFound, "NotFound"},
		{KindConflict, "Conflict"},
		{KindUnauthorized, "Unauthorized"},
		{KindForbidden, "Forbidden"},
		{KindInternal, "Internal"},
		{KindTimeout, "Timeout"},
		{KindRetryable, "Retryable"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.kind.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestKindIsRetryable(t *testing.T) {
	tests := []struct {
		kind     Kind
		expected bool
	}{
		{KindRetryable, true},
		{KindTimeout, true},
		{KindValidation, false},
		{KindNotFound, false},
		{KindInternal, false},
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			result := tt.kind.IsRetryable()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
