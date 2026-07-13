package mock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/f2xme/gox/payment"
)

const (
	maxNotificationBody = 64 << 10
	notificationVersion = 1
	matchPaymentKind    = "payment"
	matchRefundKind     = "refund"
)

type notificationEnvelope struct {
	Version          int                   `json:"version"`
	Kind             string                `json:"kind"`
	Provider         payment.Provider      `json:"provider"`
	OrderID          string                `json:"order_id"`
	TransactionID    string                `json:"transaction_id"`
	RefundID         string                `json:"refund_id,omitempty"`
	ProviderRefundID string                `json:"provider_refund_id,omitempty"`
	PaymentStatus    payment.PaymentStatus `json:"payment_status,omitempty"`
	RefundStatus     payment.RefundStatus  `json:"refund_status,omitempty"`
	Amount           int64                 `json:"amount"`
	OccurredAt       *time.Time            `json:"occurred_at,omitempty"`
	Extra            map[string]any        `json:"extra,omitempty"`
}

// PaymentNotificationRequest 根据当前支付状态创建 mock 回调请求。
func (c *Client) PaymentNotificationRequest(orderID string) (*http.Request, error) {
	if err := payment.ValidateOrderID(orderID); err != nil {
		return nil, err
	}
	c.mu.RLock()
	record, exists := c.payments[orderID]
	if !exists {
		c.mu.RUnlock()
		return nil, fmt.Errorf("%w: mock order %q not found", payment.ErrInvalidRequest, orderID)
	}
	envelope := notificationEnvelope{
		Version:       notificationVersion,
		Kind:          matchPaymentKind,
		Provider:      c.options.Provider,
		OrderID:       record.Order.OrderID,
		TransactionID: record.Result.TransactionID,
		PaymentStatus: record.Status,
		Amount:        record.Order.Amount,
		OccurredAt:    cloneTime(record.PaidAt),
		Extra:         cloneMap(record.Order.Extra),
	}
	c.mu.RUnlock()
	return newNotificationRequest("https://mock.invalid/payment/notify", envelope)
}

// RefundNotificationRequest 根据当前退款状态创建 mock 回调请求。
func (c *Client) RefundNotificationRequest(refundID string) (*http.Request, error) {
	if refundID == "" {
		return nil, fmt.Errorf("%w: refund ID cannot be empty", payment.ErrInvalidRequest)
	}
	c.mu.RLock()
	refund, exists := c.refunds[refundID]
	if !exists {
		c.mu.RUnlock()
		return nil, fmt.Errorf("%w: mock refund %q not found", payment.ErrInvalidRequest, refundID)
	}
	order := c.payments[refund.Request.OrderID]
	envelope := notificationEnvelope{
		Version:          notificationVersion,
		Kind:             matchRefundKind,
		Provider:         c.options.Provider,
		OrderID:          refund.Request.OrderID,
		TransactionID:    order.Result.TransactionID,
		RefundID:         refund.Request.RefundID,
		ProviderRefundID: refund.Result.TransactionID,
		RefundStatus:     refund.Result.Status,
		Amount:           refund.Request.Amount,
		OccurredAt:       cloneTime(refund.Result.RefundAt),
	}
	c.mu.RUnlock()
	return newNotificationRequest("https://mock.invalid/refund/notify", envelope)
}

// ParsePaymentNotification 解析 mock JSON 支付回调。
func (c *Client) ParsePaymentNotification(ctx context.Context, req *http.Request) (*payment.PaymentNotification, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	envelope, err := decodeNotification(req)
	if err != nil {
		return nil, err
	}
	if envelope.Kind != matchPaymentKind || envelope.Provider != c.options.Provider || envelope.OrderID == "" || envelope.TransactionID == "" || envelope.Amount <= 0 || !validPaymentStatus(envelope.PaymentStatus) {
		return nil, fmt.Errorf("%w: invalid mock payment notification", payment.ErrInvalidRequest)
	}
	if err := c.begin(ctx, OperationParsePaymentNotification, envelope.OrderID, ""); err != nil {
		return nil, err
	}
	return &payment.PaymentNotification{
		Provider:      envelope.Provider,
		OrderID:       envelope.OrderID,
		TransactionID: envelope.TransactionID,
		Status:        envelope.PaymentStatus,
		Amount:        envelope.Amount,
		PaidAt:        cloneTime(envelope.OccurredAt),
		Extra:         cloneMap(envelope.Extra),
	}, nil
}

// ParseRefundNotification 解析 mock JSON 退款回调。
func (c *Client) ParseRefundNotification(ctx context.Context, req *http.Request) (*payment.RefundNotification, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	envelope, err := decodeNotification(req)
	if err != nil {
		return nil, err
	}
	if envelope.Kind != matchRefundKind || envelope.Provider != c.options.Provider || envelope.OrderID == "" || envelope.TransactionID == "" || envelope.RefundID == "" || envelope.ProviderRefundID == "" || envelope.Amount <= 0 || !validRefundStatus(envelope.RefundStatus) {
		return nil, fmt.Errorf("%w: invalid mock refund notification", payment.ErrInvalidRequest)
	}
	if err := c.begin(ctx, OperationParseRefundNotification, envelope.OrderID, envelope.RefundID); err != nil {
		return nil, err
	}
	return &payment.RefundNotification{
		Provider:         envelope.Provider,
		OrderID:          envelope.OrderID,
		TransactionID:    envelope.TransactionID,
		RefundID:         envelope.RefundID,
		ProviderRefundID: envelope.ProviderRefundID,
		Status:           envelope.RefundStatus,
		Amount:           envelope.Amount,
		RefundAt:         cloneTime(envelope.OccurredAt),
		Extra:            cloneMap(envelope.Extra),
	}, nil
}

// SuccessResponse 返回 mock 回调成功回执。
func (c *Client) SuccessResponse() payment.NotifyResponse {
	return payment.NotifyResponse{
		StatusCode:  http.StatusOK,
		ContentType: "application/json; charset=utf-8",
		Body:        []byte(`{"code":"SUCCESS"}`),
	}
}

func newNotificationRequest(target string, envelope notificationEnvelope) (*http.Request, error) {
	body, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("encode mock notification: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create mock notification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func decodeNotification(req *http.Request) (notificationEnvelope, error) {
	if req == nil || req.Body == nil {
		return notificationEnvelope{}, fmt.Errorf("%w: mock notification request cannot be nil", payment.ErrInvalidRequest)
	}
	body, err := io.ReadAll(io.LimitReader(req.Body, maxNotificationBody+1))
	if err != nil {
		return notificationEnvelope{}, fmt.Errorf("%w: read mock notification: %v", payment.ErrInvalidRequest, err)
	}
	if len(body) > maxNotificationBody {
		return notificationEnvelope{}, fmt.Errorf("%w: mock notification exceeds 64 KiB", payment.ErrInvalidRequest)
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var envelope notificationEnvelope
	if err := decoder.Decode(&envelope); err != nil {
		return notificationEnvelope{}, fmt.Errorf("%w: decode mock notification: %v", payment.ErrInvalidRequest, err)
	}
	if envelope.Version != notificationVersion {
		return notificationEnvelope{}, fmt.Errorf("%w: unsupported mock notification version", payment.ErrInvalidRequest)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return notificationEnvelope{}, fmt.Errorf("%w: trailing mock notification JSON", payment.ErrInvalidRequest)
	}
	return envelope, nil
}
