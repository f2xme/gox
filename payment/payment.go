// Package payment provides a unified interface for payment operations.
//
// It abstracts common payment operations across different payment providers
// (WeChat Pay, Alipay, etc.) with a consistent API.
//
// Features:
//   - Unified payment interface for multiple providers
//   - Support for payment, query, refund, and close operations
//   - Flexible Extra fields for provider-specific parameters
//   - Amount in cents (int64) to avoid floating-point precision issues
//
// All implementations should be safe for concurrent use.
package payment

import "time"

// Payment defines the interface for payment operations.
type Payment interface {
	// Pay initiates a payment with the given order.
	// Returns payment result including transaction ID and payment URL/parameters.
	Pay(order *Order) (*PaymentResult, error)

	// Query queries the payment status of an order.
	Query(orderID string) (*QueryResult, error)

	// Refund initiates a refund for a paid order.
	Refund(req *RefundRequest) (*RefundResult, error)

	// Close closes an unpaid order.
	// Once closed, the order cannot be paid.
	Close(orderID string) error
}

// Order represents a payment order.
type Order struct {
	// OrderID is the merchant order ID (must be unique).
	OrderID string

	// Amount is the payment amount in cents (e.g., 100 = 1.00 CNY).
	Amount int64

	// Subject is the order title/subject.
	Subject string

	// Description is the order description (optional).
	Description string

	// NotifyURL is the URL for asynchronous payment notification.
	NotifyURL string

	// ReturnURL is the URL for synchronous redirect after payment (optional).
	// Used for web/H5 payments.
	ReturnURL string

	// Extra contains provider-specific parameters.
	// For example:
	//   - WeChat: {"openid": "xxx"} for JSAPI payment
	//   - Alipay: {"quit_url": "xxx"} for H5 payment
	Extra map[string]any
}

// PaymentResult represents the result of a payment initiation.
type PaymentResult struct {
	// OrderID is the merchant order ID.
	OrderID string

	// TransactionID is the payment provider's transaction ID.
	TransactionID string

	// PayURL is the payment URL for H5/PC payments.
	// Empty for APP/Mini Program payments (use Extra instead).
	PayURL string

	// Extra contains provider-specific payment parameters.
	// For example:
	//   - WeChat APP: {"appid": "xxx", "partnerid": "xxx", "prepayid": "xxx", ...}
	//   - WeChat Mini Program: {"appId": "xxx", "timeStamp": "xxx", "package": "xxx", ...}
	//   - Alipay APP: {"orderString": "xxx"}
	Extra map[string]any
}

// QueryResult represents the result of a payment query.
type QueryResult struct {
	// OrderID is the merchant order ID.
	OrderID string

	// TransactionID is the payment provider's transaction ID.
	TransactionID string

	// Status is the payment status.
	Status PaymentStatus

	// Amount is the payment amount in cents.
	Amount int64

	// PaidAt is the time when the payment was completed.
	// Nil if payment is not completed yet.
	PaidAt *time.Time
}

// RefundRequest represents a refund request.
type RefundRequest struct {
	// OrderID is the original merchant order ID.
	OrderID string

	// RefundID is the merchant refund ID (must be unique).
	RefundID string

	// Amount is the refund amount in cents.
	// Must be less than or equal to the original payment amount.
	Amount int64

	// Reason is the refund reason (optional).
	Reason string

	// NotifyURL is the URL for asynchronous refund notification (optional).
	NotifyURL string
}

// RefundResult represents the result of a refund.
type RefundResult struct {
	// RefundID is the merchant refund ID.
	RefundID string

	// Status is the refund status.
	Status RefundStatus

	// RefundAt is the time when the refund was completed.
	// Nil if refund is not completed yet.
	RefundAt *time.Time
}

// PaymentStatus represents the status of a payment.
type PaymentStatus string

const (
	// PaymentStatusPending indicates the payment is pending.
	PaymentStatusPending PaymentStatus = "pending"

	// PaymentStatusSuccess indicates the payment was successful.
	PaymentStatusSuccess PaymentStatus = "success"

	// PaymentStatusFailed indicates the payment failed.
	PaymentStatusFailed PaymentStatus = "failed"

	// PaymentStatusClosed indicates the payment was closed.
	PaymentStatusClosed PaymentStatus = "closed"
)

// RefundStatus represents the status of a refund.
type RefundStatus string

const (
	// RefundStatusPending indicates the refund is pending.
	RefundStatusPending RefundStatus = "pending"

	// RefundStatusSuccess indicates the refund was successful.
	RefundStatusSuccess RefundStatus = "success"

	// RefundStatusFailed indicates the refund failed.
	RefundStatusFailed RefundStatus = "failed"
)
