package mock

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/f2xme/gox/payment"
)

func TestPaymentAndRefundNotifications(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	client, err := New(
		WithClock(func() time.Time { return now }),
		WithPaymentStatus(payment.PaymentStatusSuccess),
		WithRefundStatus(payment.RefundStatusSuccess),
	)
	if err != nil {
		t.Fatal(err)
	}
	order := testOrder("order-1", 100)
	if _, err := client.Pay(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	refund := &payment.RefundRequest{OrderID: order.OrderID, RefundID: "refund-1", Amount: 100, OriginalAmount: 100}
	if _, err := client.Refund(context.Background(), refund); err != nil {
		t.Fatal(err)
	}

	payReq, err := client.PaymentNotificationRequest(order.OrderID)
	if err != nil {
		t.Fatal(err)
	}
	payNotify, err := client.ParsePaymentNotification(context.Background(), payReq)
	if err != nil {
		t.Fatal(err)
	}
	if payNotify.Provider != payment.ProviderMock || payNotify.OrderID != order.OrderID || payNotify.Amount != 100 || payNotify.Status != payment.PaymentStatusRefunded || payNotify.PaidAt == nil {
		t.Fatalf("payment notification = %#v", payNotify)
	}

	refundReq, err := client.RefundNotificationRequest(refund.RefundID)
	if err != nil {
		t.Fatal(err)
	}
	refundNotify, err := client.ParseRefundNotification(context.Background(), refundReq)
	if err != nil {
		t.Fatal(err)
	}
	if refundNotify.Provider != payment.ProviderMock || refundNotify.RefundID != refund.RefundID || refundNotify.ProviderRefundID != "mock-refund-refund-1" || refundNotify.Amount != 100 || refundNotify.Status != payment.RefundStatusSuccess || refundNotify.RefundAt == nil {
		t.Fatalf("refund notification = %#v", refundNotify)
	}

	response := client.SuccessResponse()
	recorder := httptest.NewRecorder()
	if err := response.WriteTo(recorder); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK || recorder.Header().Get("Content-Type") != "application/json; charset=utf-8" || recorder.Body.String() != `{"code":"SUCCESS"}` {
		t.Fatalf("success response = %d %q %q", recorder.Code, recorder.Header().Get("Content-Type"), recorder.Body.String())
	}
}

func TestNotificationParsingLimitsAndErrors(t *testing.T) {
	client, err := New(WithPaymentStatus(payment.PaymentStatusSuccess))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Pay(context.Background(), testOrder("order-1", 1)); err != nil {
		t.Fatal(err)
	}
	req, err := client.PaymentNotificationRequest("order-1")
	if err != nil {
		t.Fatal(err)
	}
	body := readRequestBody(t, req)

	tests := []struct {
		name string
		body string
	}{
		{name: "too large", body: strings.Repeat("x", maxNotificationBody+1)},
		{name: "trailing JSON", body: body + `{}`},
		{name: "unknown field", body: strings.Replace(body, `"version":1`, `"version":1,"unknown":true`, 1)},
		{name: "wrong kind", body: strings.Replace(body, `"kind":"payment"`, `"kind":"refund"`, 1)},
		{name: "wrong provider", body: strings.Replace(body, `"provider":"mock"`, `"provider":"wechat"`, 1)},
		{name: "missing transaction", body: strings.Replace(body, `"transaction_id":"mock-pay-order-1"`, `"transaction_id":""`, 1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/notify", strings.NewReader(tt.body))
			if _, err := client.ParsePaymentNotification(context.Background(), r); !errors.Is(err, payment.ErrInvalidRequest) {
				t.Fatalf("ParsePaymentNotification() error = %v", err)
			}
		})
	}

	sentinel := errors.New("parse failed")
	if err := client.SetOperationError(OperationParsePaymentNotification, sentinel); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/notify", strings.NewReader(body))
	if _, err := client.ParsePaymentNotification(context.Background(), r); !errors.Is(err, sentinel) {
		t.Fatalf("injected parse error = %v", err)
	}
	if _, err := client.PaymentNotificationRequest("missing"); !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("missing notification request error = %v", err)
	}
}

func TestRefundNotificationRejectsMissingIdentifiers(t *testing.T) {
	client, err := New(
		WithPaymentStatus(payment.PaymentStatusSuccess),
		WithRefundStatus(payment.RefundStatusSuccess),
	)
	if err != nil {
		t.Fatal(err)
	}
	order := testOrder("order-1", 100)
	if _, err := client.Pay(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	refund := &payment.RefundRequest{OrderID: order.OrderID, RefundID: "refund-1", Amount: 100, OriginalAmount: 100}
	if _, err := client.Refund(context.Background(), refund); err != nil {
		t.Fatal(err)
	}
	req, err := client.RefundNotificationRequest(refund.RefundID)
	if err != nil {
		t.Fatal(err)
	}
	body := readRequestBody(t, req)
	tests := []struct {
		name string
		body string
	}{
		{name: "missing transaction", body: strings.Replace(body, `"transaction_id":"mock-pay-order-1"`, `"transaction_id":""`, 1)},
		{name: "missing provider refund", body: strings.Replace(body, `"provider_refund_id":"mock-refund-refund-1"`, `"provider_refund_id":""`, 1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/notify", strings.NewReader(tt.body))
			if _, err := client.ParseRefundNotification(context.Background(), r); !errors.Is(err, payment.ErrInvalidRequest) {
				t.Fatalf("ParseRefundNotification() error = %v", err)
			}
		})
	}
}

func readRequestBody(t *testing.T, req *http.Request) string {
	t.Helper()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}
