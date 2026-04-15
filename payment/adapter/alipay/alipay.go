// Package alipay 为支付接口提供支付宝适配器。
//
// 这是一个骨架实现，返回模拟数据。
// 真实的支付宝 SDK 集成将在未来的任务中实现。
package alipay

import (
	"fmt"
	"time"

	"github.com/f2xme/gox/payment"
)

const (
	gatewayProd    = "https://openapi.alipay.com/gateway.do"
	gatewaySandbox = "https://openapi.alipaydev.com/gateway.do"
)

// Alipay 是实现 payment.Payment 接口的支付宝适配器
type Alipay struct {
	appID      string
	privateKey string
	publicKey  string
	isSandbox  bool
}

// NewAlipay 创建新的支付宝适配器
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

// Pay initiates a payment with the given order.
//
// TODO: Integrate with real Alipay API (alipay.trade.page.pay or alipay.trade.wap.pay).
func (a *Alipay) Pay(order *payment.Order) (*payment.PaymentResult, error) {
	if err := payment.ValidateOrder(order); err != nil {
		return nil, err
	}

	gateway := gatewayProd
	if a.isSandbox {
		gateway = gatewaySandbox
	}

	now := time.Now().Unix()
	result := &payment.PaymentResult{
		OrderID:       order.OrderID,
		TransactionID: fmt.Sprintf("alipay_%s_%d", order.OrderID, now),
		PayURL:        fmt.Sprintf("%s?out_trade_no=%s&total_amount=%.2f", gateway, order.OrderID, float64(order.Amount)/100),
		Extra: map[string]any{
			"app_id": a.appID,
			"method": "alipay.trade.page.pay",
		},
	}

	return result, nil
}

// Query queries the payment status of an order.
//
// TODO: Integrate with real Alipay API (alipay.trade.query).
func (a *Alipay) Query(orderID string) (*payment.QueryResult, error) {
	if err := payment.ValidateOrderID(orderID); err != nil {
		return nil, err
	}

	result := &payment.QueryResult{
		OrderID:       orderID,
		TransactionID: fmt.Sprintf("alipay_%s_%d", orderID, time.Now().Unix()),
		Status:        payment.PaymentStatusPending,
		Amount:        0,
		PaidAt:        nil,
	}

	return result, nil
}

// Refund initiates a refund for a paid order.
//
// TODO: Integrate with real Alipay API (alipay.trade.refund).
func (a *Alipay) Refund(req *payment.RefundRequest) (*payment.RefundResult, error) {
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
// TODO: Integrate with real Alipay API (alipay.trade.close).
func (a *Alipay) Close(orderID string) error {
	return payment.ValidateOrderID(orderID)
}
