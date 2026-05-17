// Package wechat 为 payment.Payment 接口提供微信支付适配器占位实现。
//
// 当前版本不会连接微信支付网关，所有支付操作都会返回
// payment.ErrNotImplemented。
package wechat

import (
	"fmt"

	"github.com/f2xme/gox/payment"
)

// WechatPay 是实现 payment.Payment 接口的微信支付适配器占位实现。
type WechatPay struct {
	appID    string
	mchID    string
	apiKey   string
	certPath string
}

// NewWechatPay 创建新的微信支付适配器占位实现。
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

// Pay validates the order and returns payment.ErrNotImplemented.
func (w *WechatPay) Pay(order *payment.Order) (*payment.PaymentResult, error) {
	if err := payment.ValidateOrder(order); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("wechat pay: %w", payment.ErrNotImplemented)
}

// Query validates the order ID and returns payment.ErrNotImplemented.
func (w *WechatPay) Query(orderID string) (*payment.QueryResult, error) {
	if err := payment.ValidateOrderID(orderID); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("wechat query: %w", payment.ErrNotImplemented)
}

// Refund validates the request and returns payment.ErrNotImplemented.
func (w *WechatPay) Refund(req *payment.RefundRequest) (*payment.RefundResult, error) {
	if err := payment.ValidateRefundRequest(req); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("wechat refund: %w", payment.ErrNotImplemented)
}

// Close validates the order ID and returns payment.ErrNotImplemented.
func (w *WechatPay) Close(orderID string) error {
	if err := payment.ValidateOrderID(orderID); err != nil {
		return err
	}

	return fmt.Errorf("wechat close: %w", payment.ErrNotImplemented)
}
