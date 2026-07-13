package payment

import (
	"context"
	"net/http"
	"time"
)

// PaymentNotifier 定义支付回调解析能力。
type PaymentNotifier interface {
	// ParsePaymentNotification 解析并验证支付回调。
	ParsePaymentNotification(ctx context.Context, req *http.Request) (*PaymentNotification, error)
	// SuccessResponse 返回服务商要求的成功回执。
	SuccessResponse() NotifyResponse
}

// RefundNotifier 定义退款回调解析能力。
type RefundNotifier interface {
	// ParseRefundNotification 解析并验证退款回调。
	ParseRefundNotification(ctx context.Context, req *http.Request) (*RefundNotification, error)
	// SuccessResponse 返回服务商要求的成功回执。
	SuccessResponse() NotifyResponse
}

// PaymentNotification 表示验签后的支付通知。
type PaymentNotification struct {
	// Provider 是支付服务商。
	Provider Provider
	// OrderID 是商户订单号。
	OrderID string
	// TransactionID 是服务商交易流水号。
	TransactionID string
	// Status 是支付状态。
	Status PaymentStatus
	// Amount 是支付金额，单位为分。
	Amount int64
	// PaidAt 是支付完成时间。
	PaidAt *time.Time
	// Extra 保存复制后的服务商扩展字段。
	Extra map[string]any
}

// RefundNotification 表示验签后的退款通知。
type RefundNotification struct {
	// Provider 是支付服务商。
	Provider Provider
	// OrderID 是原商户订单号。
	OrderID string
	// TransactionID 是服务商交易流水号。
	TransactionID string
	// RefundID 是商户退款单号。
	RefundID string
	// ProviderRefundID 是服务商退款流水号。
	ProviderRefundID string
	// Status 是退款状态。
	Status RefundStatus
	// Amount 是退款金额，单位为分。
	Amount int64
	// RefundAt 是退款完成时间。
	RefundAt *time.Time
	// Extra 保存复制后的服务商扩展字段。
	Extra map[string]any
}

// NotifyResponse 表示支付服务商要求的 HTTP 回执。
type NotifyResponse struct {
	// StatusCode 是 HTTP 状态码。
	StatusCode int
	// ContentType 是响应 Content-Type。
	ContentType string
	// Body 是响应正文。
	Body []byte
}

// WriteTo 把回执写入 HTTP 响应。
func (r NotifyResponse) WriteTo(w http.ResponseWriter) error {
	if r.ContentType != "" {
		w.Header().Set("Content-Type", r.ContentType)
	}
	status := r.StatusCode
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, err := w.Write(r.Body)
	return err
}
