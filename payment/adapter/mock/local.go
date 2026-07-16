package mock

import (
	"context"
	"fmt"
	"net/http"

	"github.com/f2xme/gox/payment"
)

// CompletePayment 将订单置为支付成功（或保持已付/已退款），并生成可 Parse 的回调请求。
//
// 用于本地/测试环境一键模拟「用户已付款」：无需真实微信/支付宝配置与出网。
// 订单须已存在。状态规则：
//
//   - pending / failed → success，并设置 PaidAt
//   - success / refunded → 保持当前状态（refunded 不会改回 success），通知 Status 为当前订单状态
//   - closed 等其它状态 → ErrInvalidRequest
//
// 状态写入与回调快照在同一把锁内完成，避免中间被并发改写。
// 返回的 HTTP 回执请用 SuccessResponse；与通知里的业务 Status 无关。
func (c *Client) CompletePayment(orderID string) (*http.Request, error) {
	if err := payment.ValidateOrderID(orderID); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	record, exists := c.payments[orderID]
	if !exists {
		return nil, fmt.Errorf("%w: mock order %q not found", payment.ErrInvalidRequest, orderID)
	}
	switch record.Status {
	case payment.PaymentStatusSuccess, payment.PaymentStatusRefunded:
		// 已终态成功类：仅补齐 PaidAt，便于重复投递；refunded 保持不变。
		if record.PaidAt == nil {
			record.PaidAt = cloneTimeValue(c.now())
		}
	case payment.PaymentStatusPending, payment.PaymentStatusFailed:
		record.Status = payment.PaymentStatusSuccess
		if record.PaidAt == nil {
			record.PaidAt = cloneTimeValue(c.now())
		}
	default:
		return nil, fmt.Errorf("%w: mock order %q cannot complete payment from %q", payment.ErrInvalidRequest, orderID, record.Status)
	}
	return c.paymentNotificationRequestLocked(orderID)
}

// CompleteRefund 将退款置为成功，并生成可 ParseRefundNotification 的回调请求。
//
// 退款须已存在；若已是 success 则保持状态并仍可生成回调。
// 状态写入与回调快照在同一把锁内完成，与 CompletePayment 一致，避免并发改写窗口。
func (c *Client) CompleteRefund(refundID string) (*http.Request, error) {
	if refundID == "" {
		return nil, fmt.Errorf("%w: refund ID cannot be empty", payment.ErrInvalidRequest)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.setRefundStatusLocked(refundID, payment.RefundStatusSuccess); err != nil {
		return nil, err
	}
	return c.refundNotificationRequestLocked(refundID)
}

// DeliverPaymentNotification 将订单推进到可回调状态并解析支付回调，返回通知与成功回执。
//
// 适合业务回调 handler 的本地/单元测试：不经过 HTTP 服务，直接走 Parse 路径。
// 若订单已是 refunded，通知 Status 为 refunded（不是 success）；HTTP 回执仍为 SuccessResponse。
func (c *Client) DeliverPaymentNotification(ctx context.Context, orderID string) (*payment.PaymentNotification, payment.NotifyResponse, error) {
	req, err := c.CompletePayment(orderID)
	if err != nil {
		return nil, payment.NotifyResponse{}, err
	}
	notify, err := c.ParsePaymentNotification(ctx, req)
	if err != nil {
		return nil, payment.NotifyResponse{}, err
	}
	return notify, c.SuccessResponse(), nil
}

// DeliverRefundNotification 完成退款成功并解析回调，返回通知与成功回执。
func (c *Client) DeliverRefundNotification(ctx context.Context, refundID string) (*payment.RefundNotification, payment.NotifyResponse, error) {
	req, err := c.CompleteRefund(refundID)
	if err != nil {
		return nil, payment.NotifyResponse{}, err
	}
	notify, err := c.ParseRefundNotification(ctx, req)
	if err != nil {
		return nil, payment.NotifyResponse{}, err
	}
	return notify, c.SuccessResponse(), nil
}

// PayAndDeliver 发起支付、模拟付款成功并解析回调。
//
// 默认新订单为 pending，Complete 后再投递 success 通知，贴近真实异步回调时序。
// 若构造时使用了 WithPaymentStatus(success)，则 Pay 后状态已是成功，Complete 会幂等复用。
func (c *Client) PayAndDeliver(ctx context.Context, order *payment.Order) (*payment.PaymentResult, *payment.PaymentNotification, payment.NotifyResponse, error) {
	result, err := c.Pay(ctx, order)
	if err != nil {
		return nil, nil, payment.NotifyResponse{}, err
	}
	notify, resp, err := c.DeliverPaymentNotification(ctx, order.OrderID)
	if err != nil {
		return result, nil, payment.NotifyResponse{}, err
	}
	return result, notify, resp, nil
}

// NewForProvider 按渠道装配：仅 mock 在此包内创建；真实渠道须由业务注入 wechat/alipay adapter。
//
// 推荐业务侧：
//
//	switch p, _ := payment.ParseProvider(os.Getenv("PAYMENT_PROVIDER")); p {
//	case payment.ProviderMock:
//	    return mock.NewForProvider(p, opts...)
//	case payment.ProviderWechat:
//	    return wechat.New(...)
//	}
//
// 当 provider 不是 mock 时返回 ErrInvalidConfig，避免误把空配置当成可收款实现。
// 无论 opts 是否包含 WithProvider，通知与记录中的 Provider 均强制为 ProviderMock。
func NewForProvider(provider payment.Provider, opts ...Option) (*Client, error) {
	if provider != payment.ProviderMock {
		return nil, fmt.Errorf("%w: mock.NewForProvider only supports %q, got %q (wire wechat/alipay adapters in application code)",
			payment.ErrInvalidConfig, payment.ProviderMock, provider)
	}
	merged := make([]Option, 0, len(opts)+1)
	merged = append(merged, opts...)
	// 强制 mock 身份，避免 opts 中的 WithProvider 覆盖通知 Provider 字段。
	merged = append(merged, WithProvider(payment.ProviderMock))
	return New(merged...)
}
