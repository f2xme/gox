package mock

import (
	"fmt"
	"time"

	"github.com/f2xme/gox/payment"
)

// Operation 表示可注入错误和记录调用的 mock 操作。
type Operation string

const (
	// OperationPay 表示发起支付。
	OperationPay Operation = "pay"
	// OperationQuery 表示查询支付。
	OperationQuery Operation = "query"
	// OperationRefund 表示发起退款。
	OperationRefund Operation = "refund"
	// OperationClose 表示关闭支付。
	OperationClose Operation = "close"
	// OperationParsePaymentNotification 表示解析支付回调。
	OperationParsePaymentNotification Operation = "parse_payment_notification"
	// OperationParseRefundNotification 表示解析退款回调。
	OperationParseRefundNotification Operation = "parse_refund_notification"
)

// Options 定义 mock 支付配置。
type Options struct {
	// Provider 是回调和记录使用的支付服务商。
	Provider payment.Provider
	// Clock 返回记录和状态变化使用的当前时间。
	Clock func() time.Time
	// Delay 是每次支付操作前等待的固定时长。
	Delay time.Duration
	// PaymentStatus 是新支付的默认状态。
	PaymentStatus payment.PaymentStatus
	// RefundStatus 是新退款的默认状态。
	RefundStatus payment.RefundStatus
	// OperationErrors 保存每种操作固定返回的错误。
	OperationErrors map[Operation]error
}

// Option 定义 mock 支付配置函数。
type Option func(*Options)

func defaultOptions() Options {
	return Options{
		Provider:        payment.ProviderMock,
		Clock:           time.Now,
		PaymentStatus:   payment.PaymentStatusPending,
		RefundStatus:    payment.RefundStatusPending,
		OperationErrors: make(map[Operation]error),
	}
}

// WithProvider 设置 mock 使用的支付服务商。
func WithProvider(provider payment.Provider) Option {
	return func(o *Options) { o.Provider = provider }
}

// WithClock 设置记录和状态变化使用的时钟。
func WithClock(clock func() time.Time) Option {
	return func(o *Options) { o.Clock = clock }
}

// WithDelay 设置尊重 context 取消的固定延迟。
func WithDelay(delay time.Duration) Option {
	return func(o *Options) { o.Delay = delay }
}

// WithPaymentStatus 设置新支付的默认状态。
func WithPaymentStatus(status payment.PaymentStatus) Option {
	return func(o *Options) { o.PaymentStatus = status }
}

// WithRefundStatus 设置新退款的默认状态。
func WithRefundStatus(status payment.RefundStatus) Option {
	return func(o *Options) { o.RefundStatus = status }
}

// WithOperationError 设置指定操作固定返回的错误。
func WithOperationError(operation Operation, err error) Option {
	return func(o *Options) {
		if o.OperationErrors == nil {
			o.OperationErrors = make(map[Operation]error)
		}
		o.OperationErrors[operation] = err
	}
}

func validateOptions(o Options) error {
	switch {
	case o.Provider == "":
		return fmt.Errorf("%w: mock provider cannot be empty", payment.ErrInvalidConfig)
	case o.Clock == nil:
		return fmt.Errorf("%w: mock clock cannot be nil", payment.ErrInvalidConfig)
	case o.Delay < 0:
		return fmt.Errorf("%w: mock delay cannot be negative", payment.ErrInvalidConfig)
	case !validPaymentStatus(o.PaymentStatus):
		return fmt.Errorf("%w: invalid mock payment status %q", payment.ErrInvalidConfig, o.PaymentStatus)
	case !validRefundStatus(o.RefundStatus):
		return fmt.Errorf("%w: invalid mock refund status %q", payment.ErrInvalidConfig, o.RefundStatus)
	}
	for operation := range o.OperationErrors {
		if !validOperation(operation) {
			return fmt.Errorf("%w: invalid mock operation %q", payment.ErrInvalidConfig, operation)
		}
	}
	return nil
}

func validOperation(operation Operation) bool {
	switch operation {
	case OperationPay, OperationQuery, OperationRefund, OperationClose,
		OperationParsePaymentNotification, OperationParseRefundNotification:
		return true
	default:
		return false
	}
}

func validPaymentStatus(status payment.PaymentStatus) bool {
	switch status {
	case payment.PaymentStatusPending, payment.PaymentStatusSuccess,
		payment.PaymentStatusFailed, payment.PaymentStatusClosed,
		payment.PaymentStatusRefunded:
		return true
	default:
		return false
	}
}

func validRefundStatus(status payment.RefundStatus) bool {
	switch status {
	case payment.RefundStatusPending, payment.RefundStatusSuccess,
		payment.RefundStatusFailed, payment.RefundStatusClosed:
		return true
	default:
		return false
	}
}
