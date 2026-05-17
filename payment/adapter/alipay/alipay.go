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
//   - privateKey: Merchant private key for signing requests
//   - publicKey: Alipay public key for verifying responses
//   - isSandbox: Whether to use sandbox environment
func NewAlipay(appID, privateKey, publicKey string, isSandbox bool) *Alipay {
	return &Alipay{
		appID:      appID,
		privateKey: privateKey,
		publicKey:  publicKey,
		isSandbox:  isSandbox,
	}
}

// Pay validates the order and returns payment.ErrNotImplemented.
func (a *Alipay) Pay(order *payment.Order) (*payment.PaymentResult, error) {
	if err := payment.ValidateOrder(order); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("alipay pay: %w", payment.ErrNotImplemented)
}

// Query validates the order ID and returns payment.ErrNotImplemented.
func (a *Alipay) Query(orderID string) (*payment.QueryResult, error) {
	if err := payment.ValidateOrderID(orderID); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("alipay query: %w", payment.ErrNotImplemented)
}

// Refund validates the request and returns payment.ErrNotImplemented.
func (a *Alipay) Refund(req *payment.RefundRequest) (*payment.RefundResult, error) {
	if err := payment.ValidateRefundRequest(req); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("alipay refund: %w", payment.ErrNotImplemented)
}

// Close validates the order ID and returns payment.ErrNotImplemented.
func (a *Alipay) Close(orderID string) error {
	if err := payment.ValidateOrderID(orderID); err != nil {
		return err
	}

	return fmt.Errorf("alipay close: %w", payment.ErrNotImplemented)
}
