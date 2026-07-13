package payment

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidateOrder(t *testing.T) {
	future := time.Now().Add(time.Minute)
	tests := []struct {
		name  string
		order *Order
		ok    bool
	}{
		{name: "valid", order: &Order{OrderID: "o1", Amount: 1, Subject: "商品", NotifyURL: "https://example.com/notify", ExpireAt: &future}, ok: true},
		{name: "nil"},
		{name: "missing id", order: &Order{Amount: 1, Subject: "商品", NotifyURL: "https://example.com/notify"}},
		{name: "zero amount", order: &Order{OrderID: "o1", Subject: "商品", NotifyURL: "https://example.com/notify"}},
		{name: "missing subject", order: &Order{OrderID: "o1", Amount: 1, NotifyURL: "https://example.com/notify"}},
		{name: "bad notify", order: &Order{OrderID: "o1", Amount: 1, Subject: "商品", NotifyURL: "/notify"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrder(tt.order)
			if (err == nil) != tt.ok {
				t.Fatalf("ValidateOrder() error = %v, ok = %v", err, tt.ok)
			}
			if err != nil && !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("expected ErrInvalidRequest, got %v", err)
			}
		})
	}
}

func TestValidateRefundRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *RefundRequest
		ok   bool
	}{
		{name: "valid", req: &RefundRequest{OrderID: "o1", RefundID: "r1", Amount: 50, OriginalAmount: 100}, ok: true},
		{name: "missing original", req: &RefundRequest{OrderID: "o1", RefundID: "r1", Amount: 50}},
		{name: "too much", req: &RefundRequest{OrderID: "o1", RefundID: "r1", Amount: 101, OriginalAmount: 100}},
		{name: "bad notify", req: &RefundRequest{OrderID: "o1", RefundID: "r1", Amount: 50, OriginalAmount: 100, NotifyURL: "relative"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRefundRequest(tt.req)
			if (err == nil) != tt.ok {
				t.Fatalf("ValidateRefundRequest() error = %v, ok = %v", err, tt.ok)
			}
		})
	}
}

func TestValidateContext(t *testing.T) {
	if err := ValidateContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := ValidateContext(nil); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected ErrInvalidRequest, got %v", err)
	}
}

func TestNotifyResponseWriteTo(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NotifyResponse{StatusCode: http.StatusAccepted, ContentType: "text/plain", Body: []byte("ok")}
	if err := response.WriteTo(recorder); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusAccepted || recorder.Header().Get("Content-Type") != "text/plain" || !bytes.Equal(recorder.Body.Bytes(), []byte("ok")) {
		t.Fatalf("unexpected response: code=%d header=%v body=%q", recorder.Code, recorder.Header(), recorder.Body.String())
	}
}

func TestProviderErrorUnwrap(t *testing.T) {
	cause := context.DeadlineExceeded
	err := &ProviderError{Provider: ProviderWechat, Operation: "pay", Code: "FAIL", Err: cause}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped cause, got %v", err)
	}
	if got := err.Error(); got == "" {
		t.Fatal("expected error message")
	}
}
