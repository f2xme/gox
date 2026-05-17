package payment

import (
	"errors"
	"time"
)

// ErrNotImplemented 表示适配器尚未接入真实支付服务商。
var ErrNotImplemented = errors.New("payment adapter is not implemented")

// Payment 定义统一的支付操作接口。
type Payment interface {
	// Pay 使用给定订单发起支付。
	// 返回结果包含交易流水号、支付链接或服务商专有支付参数。
	Pay(order *Order) (*PaymentResult, error)

	// Query 查询订单的支付状态。
	Query(orderID string) (*QueryResult, error)

	// Refund 为已支付订单发起退款。
	Refund(req *RefundRequest) (*RefundResult, error)

	// Close 关闭未支付订单。
	// 订单关闭后不能继续支付。
	Close(orderID string) error
}

// Order 表示支付订单。
type Order struct {
	// OrderID 是商户订单号，必须唯一。
	OrderID string

	// Amount 是支付金额，单位为分，例如 100 表示 1.00 元。
	Amount int64

	// Subject 是订单标题。
	Subject string

	// Description 是订单描述，可选。
	Description string

	// NotifyURL 是异步支付通知地址。
	NotifyURL string

	// ReturnURL 是支付完成后的同步跳转地址，可选。
	// 常用于网页或 H5 支付。
	ReturnURL string

	// Extra 保存服务商专有参数。
	// 例如：
	//   - 微信 JSAPI 支付可传入 {"openid": "xxx"}
	//   - 支付宝 H5 支付可传入 {"quit_url": "xxx"}
	Extra map[string]any
}

// PaymentResult 表示发起支付后的结果。
type PaymentResult struct {
	// OrderID 是商户订单号。
	OrderID string

	// TransactionID 是支付服务商交易流水号。
	TransactionID string

	// PayURL 是 H5 或 PC 支付链接。
	// APP 或小程序支付通常为空，应使用 Extra 中的专有参数。
	PayURL string

	// Extra 保存服务商专有支付参数。
	// 例如：
	//   - 微信 APP 支付：{"appid": "xxx", "partnerid": "xxx", "prepayid": "xxx"}
	//   - 微信小程序支付：{"appId": "xxx", "timeStamp": "xxx", "package": "xxx"}
	//   - 支付宝 APP 支付：{"orderString": "xxx"}
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
	// 支付尚未完成时为 nil。
	PaidAt *time.Time
}

// RefundRequest 表示退款请求。
type RefundRequest struct {
	// OrderID 是原商户订单号。
	OrderID string

	// RefundID 是商户退款单号，必须唯一。
	RefundID string

	// Amount 是退款金额，单位为分。
	// 必须小于或等于原支付金额。
	Amount int64

	// Reason 是退款原因，可选。
	Reason string

	// NotifyURL 是异步退款通知地址，可选。
	NotifyURL string
}

// RefundResult 表示退款结果。
type RefundResult struct {
	// RefundID 是商户退款单号。
	RefundID string

	// Status 是退款状态。
	Status RefundStatus

	// RefundAt 是退款完成时间。
	// 退款尚未完成时为 nil。
	RefundAt *time.Time
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
)
