// Package wechat 为支付接口提供微信支付适配器。
//
// 这是一个骨架实现，返回模拟数据。
// 真实的微信支付 SDK 集成将在未来的任务中实现。
package wechat

import (
	"fmt"
	"time"

	"github.com/f2xme/gox/payment"
)

// WechatPay 是实现 payment.Payment 接口的微信支付适配器
type WechatPay struct {
	appID    string
	mchID    string
	apiKey   string
	certPath string
}

// NewWechatPay 创建新的微信支付适配器
//
// 参数：
//   - appID: 微信应用 ID
//   - mchID: 微信商户 ID
//   - apiKey: 用于签名的微信 API 密钥
//   - certPath: 商户证书路径（可选，退款时必需）
func NewWechatPay(appID, mchID, apiKey, certPath string) *WechatPay {
	return &WechatPay{
		appID:    appID,
		mchID:    mchID,
		apiKey:   apiKey,
		certPath: certPath,
	}
}

// Pay initiates a payment with the given order.
//
// TODO: Integrate with real WeChat Pay API (unified order endpoint).
func (w *WechatPay) Pay(order *payment.Order) (*payment.PaymentResult, error) {
	if err := payment.ValidateOrder(order); err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	result := &payment.PaymentResult{
		OrderID:       order.OrderID,
		TransactionID: fmt.Sprintf("wx_%s_%d", order.OrderID, now),
		Extra: map[string]any{
			"prepay_id": fmt.Sprintf("prepay_%d", now),
			"appid":     w.appID,
			"mch_id":    w.mchID,
		},
	}

	return result, nil
}

// Query queries the payment status of an order.
//
// TODO: Integrate with real WeChat Pay API (order query endpoint).
func (w *WechatPay) Query(orderID string) (*payment.QueryResult, error) {
	if err := payment.ValidateOrderID(orderID); err != nil {
		return nil, err
	}

	result := &payment.QueryResult{
		OrderID:       orderID,
		TransactionID: fmt.Sprintf("wx_%s_%d", orderID, time.Now().Unix()),
		Status:        payment.PaymentStatusPending,
		Amount:        0,
		PaidAt:        nil,
	}

	return result, nil
}

// Refund initiates a refund for a paid order.
//
// TODO: Integrate with real WeChat Pay API (refund endpoint).
func (w *WechatPay) Refund(req *payment.RefundRequest) (*payment.RefundResult, error) {
	if err := payment.ValidateRefundRequest(req); err != nil {
		return nil, err
	}

	result := &payment.RefundResult{
		RefundID: req.RefundID,
		Status:   payment.RefundStatusPending,
		RefundAt: nil,
	}

	return result, nil
}

// Close closes an unpaid order.
//
// TODO: Integrate with real WeChat Pay API (close order endpoint).
func (w *WechatPay) Close(orderID string) error {
	return payment.ValidateOrderID(orderID)
}
