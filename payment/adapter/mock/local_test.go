package mock

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/f2xme/gox/payment"
)

// ledger 模拟业务侧订单账本，用于本地无真实支付配置时的端到端场景。
type ledger struct {
	mu     sync.Mutex
	orders map[string]*ledgerOrder
}

type ledgerOrder struct {
	OrderID       string
	Amount        int64
	Status        payment.PaymentStatus
	TransactionID string
	PaidAt        *time.Time
	Refunded      int64
}

func newLedger() *ledger {
	return &ledger{orders: make(map[string]*ledgerOrder)}
}

func (l *ledger) create(orderID string, amount int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.orders[orderID]; ok {
		return errors.New("duplicate order")
	}
	l.orders[orderID] = &ledgerOrder{OrderID: orderID, Amount: amount, Status: payment.PaymentStatusPending}
	return nil
}

// applyPayment 幂等入账：重复成功回调不重复加钱。
func (l *ledger) applyPayment(n *payment.PaymentNotification) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	o, ok := l.orders[n.OrderID]
	if !ok {
		return errors.New("order not found")
	}
	if n.Amount != o.Amount {
		return errors.New("amount mismatch")
	}
	if o.Status == payment.PaymentStatusSuccess || o.Status == payment.PaymentStatusRefunded {
		return nil // 幂等
	}
	if n.Status != payment.PaymentStatusSuccess {
		return errors.New("unexpected status")
	}
	o.Status = payment.PaymentStatusSuccess
	o.TransactionID = n.TransactionID
	o.PaidAt = n.PaidAt
	return nil
}

func (l *ledger) applyRefund(n *payment.RefundNotification) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	o, ok := l.orders[n.OrderID]
	if !ok {
		return errors.New("order not found")
	}
	if o.Status != payment.PaymentStatusSuccess && o.Status != payment.PaymentStatusRefunded {
		return errors.New("order not paid")
	}
	// 简化：同一退款单号由调用方保证；按金额累加且不超过订单。
	if o.Refunded+n.Amount > o.Amount {
		return errors.New("refund exceeds")
	}
	if n.Status != payment.RefundStatusSuccess {
		return errors.New("unexpected refund status")
	}
	o.Refunded += n.Amount
	if o.Refunded >= o.Amount {
		o.Status = payment.PaymentStatusRefunded
	}
	return nil
}

func (l *ledger) get(orderID string) *ledgerOrder {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.orders[orderID]
}

func TestLocalPayAndDeliverFlow(t *testing.T) {
	now := time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC)
	client, err := New(WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	books := newLedger()
	order := testOrder("local-order-1", 9900)
	if err := books.create(order.OrderID, order.Amount); err != nil {
		t.Fatal(err)
	}

	result, notify, resp, err := client.PayAndDeliver(context.Background(), order)
	if err != nil {
		t.Fatal(err)
	}
	if result.PayURL == "" || result.TransactionID == "" {
		t.Fatalf("result = %#v", result)
	}
	if notify.Status != payment.PaymentStatusSuccess || notify.Amount != 9900 || notify.PaidAt == nil {
		t.Fatalf("notify = %#v", notify)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp = %#v", resp)
	}
	if err := books.applyPayment(notify); err != nil {
		t.Fatal(err)
	}
	// 幂等：重复投递
	notify2, _, err := client.DeliverPaymentNotification(context.Background(), order.OrderID)
	if err != nil {
		t.Fatal(err)
	}
	if err := books.applyPayment(notify2); err != nil {
		t.Fatal(err)
	}
	got := books.get(order.OrderID)
	if got.Status != payment.PaymentStatusSuccess || got.TransactionID != result.TransactionID {
		t.Fatalf("ledger = %#v", got)
	}
}

func TestLocalRefundFlow(t *testing.T) {
	client, err := New(
		WithPaymentStatus(payment.PaymentStatusSuccess),
		WithRefundStatus(payment.RefundStatusPending),
	)
	if err != nil {
		t.Fatal(err)
	}
	books := newLedger()
	order := testOrder("local-refund-1", 1000)
	if err := books.create(order.OrderID, order.Amount); err != nil {
		t.Fatal(err)
	}
	_, notify, _, err := client.PayAndDeliver(context.Background(), order)
	if err != nil {
		t.Fatal(err)
	}
	if err := books.applyPayment(notify); err != nil {
		t.Fatal(err)
	}

	refundReq := &payment.RefundRequest{
		OrderID: order.OrderID, RefundID: "r1", Amount: 400, OriginalAmount: 1000,
	}
	if _, err := client.Refund(context.Background(), refundReq); err != nil {
		t.Fatal(err)
	}
	rNotify, _, err := client.DeliverRefundNotification(context.Background(), "r1")
	if err != nil {
		t.Fatal(err)
	}
	if err := books.applyRefund(rNotify); err != nil {
		t.Fatal(err)
	}
	got := books.get(order.OrderID)
	if got.Refunded != 400 || got.Status != payment.PaymentStatusSuccess {
		t.Fatalf("partial refund ledger = %#v", got)
	}

	if _, err := client.Refund(context.Background(), &payment.RefundRequest{
		OrderID: order.OrderID, RefundID: "r2", Amount: 600, OriginalAmount: 1000,
	}); err != nil {
		t.Fatal(err)
	}
	rNotify2, _, err := client.DeliverRefundNotification(context.Background(), "r2")
	if err != nil {
		t.Fatal(err)
	}
	if err := books.applyRefund(rNotify2); err != nil {
		t.Fatal(err)
	}
	got = books.get(order.OrderID)
	if got.Refunded != 1000 || got.Status != payment.PaymentStatusRefunded {
		t.Fatalf("full refund ledger = %#v", got)
	}
}

func TestCompletePaymentRejectsClosed(t *testing.T) {
	client, err := New()
	if err != nil {
		t.Fatal(err)
	}
	order := testOrder("closed-1", 10)
	if _, err := client.Pay(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	if err := client.Close(context.Background(), order.OrderID); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CompletePayment(order.OrderID); !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("CompletePayment(closed) = %v", err)
	}
}

func TestNewForProvider(t *testing.T) {
	client, err := NewForProvider(payment.ProviderMock)
	if err != nil || client == nil {
		t.Fatalf("NewForProvider(mock) = %v, %v", client, err)
	}
	if _, err := NewForProvider(payment.ProviderWechat); !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("NewForProvider(wechat) = %v, want ErrInvalidConfig", err)
	}

	// opts 中的 WithProvider 不得覆盖 mock 身份。
	client, err = NewForProvider(payment.ProviderMock, WithProvider(payment.ProviderWechat))
	if err != nil {
		t.Fatal(err)
	}
	order := testOrder("force-provider", 10)
	if _, err := client.Pay(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	req, err := client.CompletePayment(order.OrderID)
	if err != nil {
		t.Fatal(err)
	}
	notify, err := client.ParsePaymentNotification(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if notify.Provider != payment.ProviderMock {
		t.Fatalf("notify.Provider = %q, want mock (forced)", notify.Provider)
	}
}

func TestCompleteRefundAtomicSnapshot(t *testing.T) {
	client, err := New(WithPaymentStatus(payment.PaymentStatusSuccess))
	if err != nil {
		t.Fatal(err)
	}
	order := testOrder("refund-atomic", 500)
	if _, err := client.Pay(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Refund(context.Background(), &payment.RefundRequest{
		OrderID: order.OrderID, RefundID: "ra-1", Amount: 200, OriginalAmount: 500,
	}); err != nil {
		t.Fatal(err)
	}
	req, err := client.CompleteRefund("ra-1")
	if err != nil {
		t.Fatal(err)
	}
	notify, err := client.ParseRefundNotification(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if notify.Status != payment.RefundStatusSuccess || notify.Amount != 200 {
		t.Fatalf("refund notify = %#v", notify)
	}
	// 重复 Complete 保持 success 并仍可生成回调。
	req2, err := client.CompleteRefund("ra-1")
	if err != nil {
		t.Fatal(err)
	}
	notify2, err := client.ParseRefundNotification(context.Background(), req2)
	if err != nil || notify2.Status != payment.RefundStatusSuccess {
		t.Fatalf("second CompleteRefund = %#v, %v", notify2, err)
	}
}

func TestCompletePaymentKeepsRefundedStatus(t *testing.T) {
	client, err := New(WithPaymentStatus(payment.PaymentStatusSuccess))
	if err != nil {
		t.Fatal(err)
	}
	order := testOrder("refunded-keep", 100)
	if _, err := client.Pay(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Refund(context.Background(), &payment.RefundRequest{
		OrderID: order.OrderID, RefundID: "rf-1", Amount: 100, OriginalAmount: 100,
	}); err != nil {
		t.Fatal(err)
	}
	if err := client.SetRefundStatus("rf-1", payment.RefundStatusSuccess); err != nil {
		t.Fatal(err)
	}
	req, err := client.CompletePayment(order.OrderID)
	if err != nil {
		t.Fatal(err)
	}
	notify, err := client.ParsePaymentNotification(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if notify.Status != payment.PaymentStatusRefunded {
		t.Fatalf("notify.Status = %q, want refunded", notify.Status)
	}
}

func TestInterfaceWiringWithMock(t *testing.T) {
	// 业务只依赖接口，本地注入 mock。
	var (
		pay    payment.Payment
		notify payment.PaymentNotifier
	)
	client, err := NewForProvider(payment.ProviderMock)
	if err != nil {
		t.Fatal(err)
	}
	pay = client
	notify = client

	order := testOrder("iface-1", 50)
	if _, err := pay.Pay(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	req, err := client.CompletePayment(order.OrderID)
	if err != nil {
		t.Fatal(err)
	}
	n, err := notify.ParsePaymentNotification(context.Background(), req)
	if err != nil || n.Status != payment.PaymentStatusSuccess {
		t.Fatalf("notify = %#v, %v", n, err)
	}
	_ = notify.SuccessResponse()
}
