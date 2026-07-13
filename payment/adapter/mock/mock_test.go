package mock

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/f2xme/gox/payment"
)

type structuredExtra struct {
	Values []int
	Lookup map[string]*int
	When   time.Time
}

type privateReferenceExtra struct {
	values []int
}

var (
	_ payment.Payment         = (*Client)(nil)
	_ payment.PaymentNotifier = (*Client)(nil)
	_ payment.RefundNotifier  = (*Client)(nil)
)

func testOrder(id string, amount int64) *payment.Order {
	return &payment.Order{
		OrderID:   id,
		Amount:    amount,
		Subject:   "测试商品",
		NotifyURL: "https://example.com/payment/notify",
	}
}

func TestNewRejectsInvalidOptions(t *testing.T) {
	tests := []struct {
		name string
		opt  Option
	}{
		{name: "empty provider", opt: WithProvider("")},
		{name: "nil clock", opt: WithClock(nil)},
		{name: "negative delay", opt: WithDelay(-time.Second)},
		{name: "bad payment status", opt: WithPaymentStatus("unknown")},
		{name: "bad refund status", opt: WithRefundStatus("unknown")},
		{name: "bad operation", opt: WithOperationError("unknown", errors.New("fail"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New(tt.opt); !errors.Is(err, payment.ErrInvalidConfig) {
				t.Fatalf("New() error = %v, want ErrInvalidConfig", err)
			}
		})
	}
}

func TestPayQueryStatusAndClose(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	client, err := New(WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	typed := map[string][]int{"values": {1, 2}}
	order := testOrder("order/1", 100)
	order.Extra = map[string]any{"typed": typed}

	result, err := client.Pay(context.Background(), order)
	if err != nil {
		t.Fatal(err)
	}
	if result.OrderID != order.OrderID || result.TransactionID != "mock-pay-order/1" || result.PayURL != "mock://pay/order%2F1" {
		t.Fatalf("Pay() result = %#v", result)
	}
	query, err := client.Query(context.Background(), order.OrderID)
	if err != nil || query.Status != payment.PaymentStatusPending || query.PaidAt != nil {
		t.Fatalf("Query() = %#v, %v", query, err)
	}
	if err := client.SetPaymentStatus(order.OrderID, payment.PaymentStatusSuccess); err != nil {
		t.Fatal(err)
	}
	query, err = client.Query(context.Background(), order.OrderID)
	if err != nil || query.Status != payment.PaymentStatusSuccess || query.PaidAt == nil || !query.PaidAt.Equal(now) {
		t.Fatalf("successful Query() = %#v, %v", query, err)
	}
	if err := client.Close(context.Background(), order.OrderID); !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("Close(success) error = %v", err)
	}
	if _, err := client.Pay(context.Background(), order); !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("duplicate Pay() error = %v", err)
	}

	pending := testOrder("pending", 10)
	if _, err := client.Pay(context.Background(), pending); err != nil {
		t.Fatal(err)
	}
	if err := client.Close(context.Background(), pending.OrderID); err != nil {
		t.Fatal(err)
	}
	if err := client.Close(context.Background(), pending.OrderID); err != nil {
		t.Fatalf("repeated Close() error = %v", err)
	}
	closed, err := client.Query(context.Background(), pending.OrderID)
	if err != nil || closed.Status != payment.PaymentStatusClosed {
		t.Fatalf("closed Query() = %#v, %v", closed, err)
	}

	typed["values"][0] = 99
	records := client.Payments()
	gotTyped := records[0].Order.Extra["typed"].(map[string][]int)
	if gotTyped["values"][0] != 1 {
		t.Fatalf("stored extra changed through input alias: %#v", gotTyped)
	}
	gotTyped["values"][0] = 88
	again := client.Payments()[0].Order.Extra["typed"].(map[string][]int)
	if again["values"][0] != 1 {
		t.Fatalf("stored extra changed through output alias: %#v", again)
	}
}

func TestRefundLifecycleAndCumulativeAmount(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	client, err := New(WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	order := testOrder("order-1", 100)
	if _, err := client.Pay(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	if err := client.SetPaymentStatus(order.OrderID, payment.PaymentStatusSuccess); err != nil {
		t.Fatal(err)
	}

	first := &payment.RefundRequest{OrderID: order.OrderID, RefundID: "refund-1", Amount: 60, OriginalAmount: 100}
	result, err := client.Refund(context.Background(), first)
	if err != nil || result.Status != payment.RefundStatusPending || result.TransactionID != "mock-refund-refund-1" {
		t.Fatalf("Refund() = %#v, %v", result, err)
	}
	tooMuch := &payment.RefundRequest{OrderID: order.OrderID, RefundID: "refund-2", Amount: 50, OriginalAmount: 100}
	if _, err := client.Refund(context.Background(), tooMuch); !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("cumulative Refund() error = %v", err)
	}
	if err := client.SetRefundStatus(first.RefundID, payment.RefundStatusFailed); err != nil {
		t.Fatal(err)
	}
	second := &payment.RefundRequest{OrderID: order.OrderID, RefundID: "refund-2", Amount: 50, OriginalAmount: 100}
	if _, err := client.Refund(context.Background(), second); err != nil {
		t.Fatal(err)
	}
	if err := client.SetRefundStatus(second.RefundID, payment.RefundStatusSuccess); err != nil {
		t.Fatal(err)
	}
	third := &payment.RefundRequest{OrderID: order.OrderID, RefundID: "refund-3", Amount: 50, OriginalAmount: 100}
	if _, err := client.Refund(context.Background(), third); err != nil {
		t.Fatal(err)
	}
	if err := client.SetRefundStatus(third.RefundID, payment.RefundStatusSuccess); err != nil {
		t.Fatal(err)
	}
	query, err := client.Query(context.Background(), order.OrderID)
	if err != nil || query.Status != payment.PaymentStatusRefunded {
		t.Fatalf("Query() after full refund = %#v, %v", query, err)
	}
	if _, err := client.Refund(context.Background(), third); !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("duplicate Refund() error = %v", err)
	}
	mismatch := &payment.RefundRequest{OrderID: order.OrderID, RefundID: "refund-4", Amount: 1, OriginalAmount: 101}
	if _, err := client.Refund(context.Background(), mismatch); !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("original amount mismatch error = %v", err)
	}
}

func TestRefundRejectsCumulativeAmountOverflow(t *testing.T) {
	client, err := New(WithPaymentStatus(payment.PaymentStatusSuccess))
	if err != nil {
		t.Fatal(err)
	}
	order := testOrder("max-order", int64(^uint64(0)>>1))
	if _, err := client.Pay(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	first := &payment.RefundRequest{
		OrderID:        order.OrderID,
		RefundID:       "refund-max",
		Amount:         order.Amount - 1,
		OriginalAmount: order.Amount,
	}
	if _, err := client.Refund(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	second := &payment.RefundRequest{
		OrderID:        order.OrderID,
		RefundID:       "refund-overflow",
		Amount:         2,
		OriginalAmount: order.Amount,
	}
	if _, err := client.Refund(context.Background(), second); !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("Refund(overflow) error = %v", err)
	}
}

func TestSetRefundStatusRejectsCumulativeAmountOverflow(t *testing.T) {
	client, err := New(WithPaymentStatus(payment.PaymentStatusSuccess))
	if err != nil {
		t.Fatal(err)
	}
	order := testOrder("order-status-overflow", 100)
	if _, err := client.Pay(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	first := &payment.RefundRequest{OrderID: order.OrderID, RefundID: "refund-first", Amount: 60, OriginalAmount: 100}
	if _, err := client.Refund(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	if err := client.SetRefundStatus(first.RefundID, payment.RefundStatusFailed); err != nil {
		t.Fatal(err)
	}
	second := &payment.RefundRequest{OrderID: order.OrderID, RefundID: "refund-second", Amount: 50, OriginalAmount: 100}
	if _, err := client.Refund(context.Background(), second); err != nil {
		t.Fatal(err)
	}

	for _, status := range []payment.RefundStatus{payment.RefundStatusPending, payment.RefundStatusSuccess} {
		if err := client.SetRefundStatus(first.RefundID, status); !errors.Is(err, payment.ErrInvalidRequest) {
			t.Fatalf("SetRefundStatus(%q) error = %v", status, err)
		}
		if got := client.Refunds()[0].Result.Status; got != payment.RefundStatusFailed {
			t.Fatalf("first refund status = %q after rejected transition", got)
		}
	}
}

func TestPayDeepClonesStructuredExtra(t *testing.T) {
	client, err := New()
	if err != nil {
		t.Fatal(err)
	}
	value := 7
	when := time.Date(2026, 7, 13, 12, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	payload := structuredExtra{
		Values: []int{1, 2},
		Lookup: map[string]*int{"value": &value},
		When:   when,
	}
	order := testOrder("structured-extra", 100)
	order.Extra = map[string]any{"payload": payload}
	if _, err := client.Pay(context.Background(), order); err != nil {
		t.Fatal(err)
	}

	payload.Values[0] = 99
	value = 99
	stored := client.Payments()[0].Order.Extra["payload"].(structuredExtra)
	if stored.Values[0] != 1 || *stored.Lookup["value"] != 7 || !stored.When.Equal(when) {
		t.Fatalf("stored structured extra changed through input alias: %#v", stored)
	}
	stored.Values[0] = 88
	*stored.Lookup["value"] = 88
	again := client.Payments()[0].Order.Extra["payload"].(structuredExtra)
	if again.Values[0] != 1 || *again.Lookup["value"] != 7 {
		t.Fatalf("stored structured extra changed through output alias: %#v", again)
	}
}

func TestPayRejectsUnsupportedExtra(t *testing.T) {
	cyclic := map[string]any{}
	cyclic["self"] = cyclic
	number := 1
	tests := []struct {
		name  string
		value any
	}{
		{name: "cycle", value: cyclic},
		{name: "func", value: func() {}},
		{name: "channel", value: make(chan int)},
		{name: "unsafe pointer", value: unsafe.Pointer(&number)},
		{name: "private reference field", value: privateReferenceExtra{values: []int{1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New()
			if err != nil {
				t.Fatal(err)
			}
			order := testOrder("unsupported-extra", 100)
			order.Extra = map[string]any{"value": tt.value}
			if _, err := client.Pay(context.Background(), order); !errors.Is(err, payment.ErrInvalidRequest) {
				t.Fatalf("Pay() error = %v", err)
			}
			if len(client.Payments()) != 0 || len(client.Calls()) != 0 {
				t.Fatal("invalid extra changed mock state")
			}
		})
	}
}

func TestOperationErrorDelayCallsAndReset(t *testing.T) {
	sentinel := errors.New("gateway unavailable")
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	client, err := New(
		WithClock(func() time.Time { return now }),
		WithDelay(100*time.Millisecond),
		WithOperationError(OperationPay, sentinel),
	)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if _, err := client.Pay(ctx, testOrder("timeout", 1)); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Pay(timeout) error = %v", err)
	}
	if len(client.Payments()) != 0 || len(client.Calls()) != 1 {
		t.Fatalf("state after timeout: payments=%d calls=%d", len(client.Payments()), len(client.Calls()))
	}
	if client.Calls()[0].CalledAt != now || client.Calls()[0].Operation != OperationPay {
		t.Fatalf("call = %#v", client.Calls()[0])
	}

	client, err = New(WithOperationError(OperationPay, sentinel))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Pay(context.Background(), testOrder("failed", 1)); !errors.Is(err, sentinel) {
		t.Fatalf("Pay(injected) error = %v", err)
	}
	client.Reset()
	if len(client.Calls()) != 0 || len(client.Payments()) != 0 {
		t.Fatal("Reset() did not clear records")
	}
	if _, err := client.Pay(context.Background(), testOrder("still-failed", 1)); !errors.Is(err, sentinel) {
		t.Fatalf("Reset() cleared configured error: %v", err)
	}
	if err := client.SetOperationError(OperationPay, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Pay(context.Background(), testOrder("success", 1)); err != nil {
		t.Fatal(err)
	}
	if err := client.SetOperationError("unknown", sentinel); !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("SetOperationError() error = %v", err)
	}
}

func TestConcurrentPayments(t *testing.T) {
	var clockCalls int
	client, err := New(WithClock(func() time.Time {
		clockCalls++
		return time.Unix(int64(clockCalls), 0)
	}))
	if err != nil {
		t.Fatal(err)
	}
	const count = 32
	var wg sync.WaitGroup
	for i := range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := time.Unix(int64(i), 0).UTC().Format(time.RFC3339Nano)
			if _, err := client.Pay(context.Background(), testOrder(id, 1)); err != nil {
				t.Errorf("Pay(%q) error = %v", id, err)
			}
		}()
	}
	wg.Wait()
	if got := len(client.Payments()); got != count {
		t.Fatalf("Payments() count = %d, want %d", got, count)
	}
}
