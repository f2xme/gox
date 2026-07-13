package payment

import (
	"context"
	"time"
)

// Provider 表示支付服务提供商。
type Provider string

const (
	// ProviderAlipay 表示支付宝。
	ProviderAlipay Provider = "alipay"
	// ProviderWechat 表示微信支付。
	ProviderWechat Provider = "wechat"
)

// Payment 定义统一的支付操作接口。
type Payment interface {
	// Pay 使用给定订单发起支付。
	Pay(ctx context.Context, order *Order) (*PaymentResult, error)
	// Query 查询订单支付状态。
	Query(ctx context.Context, orderID string) (*QueryResult, error)
	// Refund 为已支付订单发起退款。
	Refund(ctx context.Context, req *RefundRequest) (*RefundResult, error)
	// Close 关闭未支付订单。
	Close(ctx context.Context, orderID string) error
}

// Order 表示支付订单。
type Order struct {
	// OrderID 是商户订单号，必须唯一。
	OrderID string
	// Amount 是支付金额，单位为分。
	Amount int64
	// Subject 是订单标题。
	Subject string
	// Description 是订单描述。
	Description string
	// NotifyURL 是异步支付通知地址。
	NotifyURL string
	// ReturnURL 是支付完成后的同步跳转地址。
	ReturnURL string
	// ExpireAt 是订单支付截止时间。
	ExpireAt *time.Time
	// Extra 保存服务商专有参数。
	Extra map[string]any
}

// PaymentResult 表示发起支付后的结果。
type PaymentResult struct {
	// OrderID 是商户订单号。
	OrderID string
	// TransactionID 是支付服务商交易流水号。
	TransactionID string
	// PayURL 是二维码内容或收银台 URL。
	PayURL string
	// Extra 保存服务商专有支付参数。
	Extra map[string]any
}

// QueryResult 表示支付查询结果。
type QueryResult struct {
	// OrderID 是商户订单号。
	OrderID string
	// TransactionID 是支付服务商交易流水号。
	TransactionID string
	// Status 是支付状态。
	Status PaymentStatus
	// Amount 是支付金额，单位为分。
	Amount int64
	// PaidAt 是支付完成时间。
	PaidAt *time.Time
}

// RefundRequest 表示退款请求。
type RefundRequest struct {
	// OrderID 是原商户订单号。
	OrderID string
	// RefundID 是商户退款单号，必须唯一。
	RefundID string
	// Amount 是退款金额，单位为分。
	Amount int64
	// OriginalAmount 是原订单总金额，单位为分。
	OriginalAmount int64
	// Reason 是退款原因。
	Reason string
	// NotifyURL 是异步退款通知地址。
	NotifyURL string
}

// RefundResult 表示退款结果。
type RefundResult struct {
	// RefundID 是商户退款单号。
	RefundID string
	// TransactionID 是支付服务商退款流水号。
	TransactionID string
	// Status 是退款状态。
	Status RefundStatus
	// RefundAt 是退款完成时间。
	RefundAt *time.Time
}

// WAPResult 表示移动网页收银台结果。
type WAPResult struct {
	// URL 是服务商生成的完整收银台 URL。
	URL string
}

// JSAPIResult 表示微信 JSAPI 调起支付参数。
type JSAPIResult struct {
	// AppID 是微信应用 ID。
	AppID string `json:"appId"`
	// Timestamp 是支付签名时间戳。
	Timestamp string `json:"timeStamp"`
	// NonceStr 是支付签名随机串。
	NonceStr string `json:"nonceStr"`
	// Package 是微信预支付包。
	Package string `json:"package"`
	// SignType 是签名算法。
	SignType string `json:"signType"`
	// PaySign 是支付签名。
	PaySign string `json:"paySign"`
}

// PaymentStatus 表示支付状态。
type PaymentStatus string

const (
	// PaymentStatusPending 表示支付待处理。
	PaymentStatusPending PaymentStatus = "pending"
	// PaymentStatusSuccess 表示支付成功。
	PaymentStatusSuccess PaymentStatus = "success"
	// PaymentStatusFailed 表示支付失败。
	PaymentStatusFailed PaymentStatus = "failed"
	// PaymentStatusClosed 表示支付已关闭。
	PaymentStatusClosed PaymentStatus = "closed"
	// PaymentStatusRefunded 表示支付已转入退款。
	PaymentStatusRefunded PaymentStatus = "refunded"
)

// RefundStatus 表示退款状态。
type RefundStatus string

const (
	// RefundStatusPending 表示退款待处理。
	RefundStatusPending RefundStatus = "pending"
	// RefundStatusSuccess 表示退款成功。
	RefundStatusSuccess RefundStatus = "success"
	// RefundStatusFailed 表示退款失败。
	RefundStatusFailed RefundStatus = "failed"
	// RefundStatusClosed 表示退款已关闭。
	RefundStatusClosed RefundStatus = "closed"
)
