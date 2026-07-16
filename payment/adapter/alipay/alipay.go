package alipay

import (
	"context"
	"fmt"

	"github.com/f2xme/gox/payment"
	"github.com/go-pay/gopay"
)

// Alipay 实现支付宝当面付、WAP、查询、退款、关单和回调验签。
// 支持密钥/证书加签，以及正式/沙箱网关环境。
type Alipay struct {
	config       Config
	gateway      gateway
	verifyNotify notifyVerifier
}

// 编译期断言：Alipay 实现核心支付与支付回调接口。
// 当面付退款为同步接口，故不实现 payment.RefundNotifier。
var (
	_ payment.Payment         = (*Alipay)(nil)
	_ payment.PaymentNotifier = (*Alipay)(nil)
)

// New 创建支付宝支付适配器。
//
// 配置须提供密钥模式（AlipayPublicKey）或证书模式
// （AppPublicCert + AlipayRootCert + AlipayPublicCert）之一。
//
// 环境通过 Environment 指定（EnvProduction / EnvSandbox）；
// 未设置时回退 Production 字段，零值默认沙箱。
func New(config Config, opts ...Option) (*Alipay, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	options := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	gw, err := newGopayGateway(config, options)
	if err != nil {
		return nil, fmt.Errorf("%w: initialize alipay client: %w", payment.ErrInvalidConfig, err)
	}
	return &Alipay{config: config, gateway: gw, verifyNotify: newNotifyVerifier(config)}, nil
}

func newWithGateway(config Config, gw gateway) *Alipay {
	return &Alipay{config: config, gateway: gw, verifyNotify: newNotifyVerifier(config)}
}

// Environment 返回当前生效的网关环境。
func (a *Alipay) Environment() Environment {
	return a.config.ResolveEnvironment()
}

// IsProduction 返回是否为正式环境。
func (a *Alipay) IsProduction() bool {
	return a.config.IsProduction()
}

// IsSandbox 返回是否为沙箱环境。
func (a *Alipay) IsSandbox() bool {
	return a.config.IsSandbox()
}

// GatewayBaseURL 返回当前使用的支付宝网关地址。
func (a *Alipay) GatewayBaseURL() string {
	return a.config.GatewayBaseURL()
}

// Pay 发起支付宝当面付预创建并返回二维码内容。
func (a *Alipay) Pay(ctx context.Context, order *payment.Order) (*payment.PaymentResult, error) {
	if err := validateCall(ctx, order); err != nil {
		return nil, err
	}
	bm := orderBody(order)
	resp, err := a.gateway.precreate(ctx, bm)
	if err != nil {
		return nil, providerError("pay", err)
	}
	if resp == nil || resp.Response == nil || resp.Response.QrCode == "" {
		return nil, providerError("pay", fmt.Errorf("empty precreate response"))
	}
	return &payment.PaymentResult{OrderID: resp.Response.OutTradeNo, PayURL: resp.Response.QrCode}, nil
}

// WAPPay 创建支付宝手机网站收银台 URL。
func (a *Alipay) WAPPay(ctx context.Context, order *payment.Order) (*payment.WAPResult, error) {
	if err := validateCall(ctx, order); err != nil {
		return nil, err
	}
	bm := orderBody(order)
	if order.ReturnURL != "" {
		bm.Set("return_url", order.ReturnURL)
	}
	url, err := a.gateway.wapPay(ctx, bm)
	if err != nil {
		return nil, providerError("wap_pay", err)
	}
	if url == "" {
		return nil, providerError("wap_pay", fmt.Errorf("empty cashier URL"))
	}
	return &payment.WAPResult{URL: url}, nil
}

// Query 查询支付宝订单。
func (a *Alipay) Query(ctx context.Context, orderID string) (*payment.QueryResult, error) {
	if err := payment.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if err := payment.ValidateOrderID(orderID); err != nil {
		return nil, err
	}
	resp, err := a.gateway.query(ctx, gopay.BodyMap{"out_trade_no": orderID})
	if err != nil {
		return nil, providerError("query", err)
	}
	if resp == nil || resp.Response == nil {
		return nil, providerError("query", fmt.Errorf("empty query response"))
	}
	status, err := mapPaymentStatus(resp.Response.TradeStatus)
	if err != nil {
		return nil, err
	}
	amount, err := yuanToCents(resp.Response.TotalAmount)
	if err != nil {
		return nil, providerError("query", err)
	}
	paidAt, err := parseAlipayTime(resp.Response.SendPayDate)
	if err != nil {
		return nil, providerError("query", err)
	}
	return &payment.QueryResult{
		OrderID:       resp.Response.OutTradeNo,
		TransactionID: resp.Response.TradeNo,
		Status:        status,
		Amount:        amount,
		PaidAt:        paidAt,
	}, nil
}

// Refund 发起支付宝退款。
func (a *Alipay) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResult, error) {
	if err := payment.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if err := payment.ValidateRefundRequest(req); err != nil {
		return nil, err
	}
	bm := gopay.BodyMap{
		"out_trade_no":   req.OrderID,
		"out_request_no": req.RefundID,
		"refund_amount":  centsToYuan(req.Amount),
	}
	if req.Reason != "" {
		bm.Set("refund_reason", req.Reason)
	}
	resp, err := a.gateway.refund(ctx, bm)
	if err != nil {
		return nil, providerError("refund", err)
	}
	if resp == nil || resp.Response == nil {
		return nil, providerError("refund", fmt.Errorf("empty refund response"))
	}
	refundAt, err := parseAlipayTime(resp.Response.GmtRefundPay)
	if err != nil {
		return nil, providerError("refund", err)
	}
	return &payment.RefundResult{
		RefundID:      req.RefundID,
		TransactionID: resp.Response.RefundSettlementId,
		Status:        payment.RefundStatusSuccess,
		RefundAt:      refundAt,
	}, nil
}

// Close 关闭支付宝订单。
func (a *Alipay) Close(ctx context.Context, orderID string) error {
	if err := payment.ValidateContext(ctx); err != nil {
		return err
	}
	if err := payment.ValidateOrderID(orderID); err != nil {
		return err
	}
	resp, err := a.gateway.close(ctx, gopay.BodyMap{"out_trade_no": orderID})
	if err != nil {
		return providerError("close", err)
	}
	if resp == nil || resp.Response == nil {
		return providerError("close", fmt.Errorf("empty close response"))
	}
	return nil
}

func orderBody(order *payment.Order) gopay.BodyMap {
	bm := gopay.BodyMap{
		"out_trade_no": order.OrderID,
		"total_amount": centsToYuan(order.Amount),
		"subject":      order.Subject,
		"notify_url":   order.NotifyURL,
	}
	if order.Description != "" {
		bm.Set("body", order.Description)
	}
	if order.ExpireAt != nil {
		bm.Set("time_expire", order.ExpireAt.In(alipayLocation).Format("2006-01-02 15:04:05"))
	}
	return bm
}

func validateCall(ctx context.Context, order *payment.Order) error {
	if err := payment.ValidateContext(ctx); err != nil {
		return err
	}
	return payment.ValidateOrder(order)
}

func providerError(operation string, err error) error {
	if err == nil {
		err = payment.ErrGateway
	}
	return &payment.ProviderError{
		Provider:  payment.ProviderAlipay,
		Operation: operation,
		Err:       fmt.Errorf("%w: %w", payment.ErrGateway, err),
	}
}
