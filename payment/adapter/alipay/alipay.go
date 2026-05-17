// Package alipay 为 payment.Payment 接口提供支付宝适配器占位实现。
//
// 当前版本不会连接支付宝网关，所有支付操作都会返回
// payment.ErrNotImplemented。
package alipay

import (
	"fmt"

	"github.com/f2xme/gox/payment"
)

// Alipay 是实现 payment.Payment 接口的支付宝适配器占位实现。
type Alipay struct {
	appID      string
	privateKey string
	publicKey  string
	isSandbox  bool
}

// NewAlipay 创建新的支付宝适配器占位实现。
//
// 参数：
//   - appID: 支付宝应用 ID
//   - privateKey: 商户私钥，用于请求签名
//   - publicKey: 支付宝公钥，用于响应验签
//   - isSandbox: 是否使用沙箱环境
func NewAlipay(appID, privateKey, publicKey string, isSandbox bool) *Alipay {
	return &Alipay{
		appID:      appID,
		privateKey: privateKey,
		publicKey:  publicKey,
		isSandbox:  isSandbox,
	}
}

// Pay 校验订单并返回 payment.ErrNotImplemented。
func (a *Alipay) Pay(order *payment.Order) (*payment.PaymentResult, error) {
	if err := payment.ValidateOrder(order); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("alipay pay: %w", payment.ErrNotImplemented)
}

// Query 校验订单号并返回 payment.ErrNotImplemented。
func (a *Alipay) Query(orderID string) (*payment.QueryResult, error) {
	if err := payment.ValidateOrderID(orderID); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("alipay query: %w", payment.ErrNotImplemented)
}

// Refund 校验退款请求并返回 payment.ErrNotImplemented。
func (a *Alipay) Refund(req *payment.RefundRequest) (*payment.RefundResult, error) {
	if err := payment.ValidateRefundRequest(req); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("alipay refund: %w", payment.ErrNotImplemented)
}

// Close 校验订单号并返回 payment.ErrNotImplemented。
func (a *Alipay) Close(orderID string) error {
	if err := payment.ValidateOrderID(orderID); err != nil {
		return err
	}

	return fmt.Errorf("alipay close: %w", payment.ErrNotImplemented)
}
