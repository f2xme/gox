package volcengine

import (
	"context"
	"errors"
	"testing"

	"github.com/f2xme/gox/sms"
)

func TestNewReturnsNotImplemented(t *testing.T) {
	client, err := New(
		WithAccessKeyID("test-key-id"),
		WithAccessKeySecret("test-key-secret"),
		WithSignName("test-sign"),
	)
	if client != nil {
		t.Fatalf("New() client = %v, want nil", client)
	}
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("New() error = %v, want %v", err, ErrNotImplemented)
	}
}

func TestNewValidatesOptionsBeforeNotImplemented(t *testing.T) {
	client, err := New()
	if client != nil {
		t.Fatalf("New() client = %v, want nil", client)
	}
	if err == nil {
		t.Fatal("New() error is nil")
	}
	if errors.Is(err, ErrNotImplemented) {
		t.Fatalf("New() error = %v, want validation error", err)
	}
}

func TestSendReturnsNotImplemented(t *testing.T) {
	client := &volcengineSMS{}
	err := client.Send(context.Background(), sms.Message{
		Phone:        "13800138000",
		TemplateCode: "SMS_123456789",
	})
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("Send() error = %v, want %v", err, ErrNotImplemented)
	}
}
