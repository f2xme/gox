// Package wechat 为 payment 包提供微信支付 V3 实现。
package wechat

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"time"

	"github.com/f2xme/gox/payment"
	"github.com/go-pay/gopay"
	wx "github.com/go-pay/gopay/wechat/v3"
)

// WechatPay 实现微信 Native、JSAPI、查询、退款、关单和回调处理。
type WechatPay struct {
	config        Config
	gateway       gateway
	oauthClient   oauthHTTPClient
	oauthAuthURL  string
	oauthTokenURL string
	parseNotify   func(*http.Request) (*wx.V3NotifyReq, error)
	verifyNotify  func(*wx.V3NotifyReq, map[string]*rsa.PublicKey) error
	decryptPay    func(*wx.V3NotifyReq, string) (*wx.V3DecryptPayResult, error)
	decryptRefund func(*wx.V3NotifyReq, string) (*wx.V3DecryptRefundResult, error)
}

// New 创建微信支付适配器。
func New(config Config, opts ...Option) (*WechatPay, error) {
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
		return nil, fmt.Errorf("%w: initialize wechat client: %v", payment.ErrInvalidConfig, err)
	}
	w := newWithGateway(config, gw)
	w.oauthClient = options.oauthClient()
	w.oauthAuthURL = options.oauthAuthURL
	w.oauthTokenURL = options.oauthTokenURL
	return w, nil
}

func newWithGateway(config Config, gw gateway) *WechatPay {
	w := &WechatPay{config: config, gateway: gw}
	w.setNotifyDefaults()
	return w
}

// Pay 发起微信 Native 支付并返回二维码内容。
func (w *WechatPay) Pay(ctx context.Context, order *payment.Order) (*payment.PaymentResult, error) {
	if err := validateCall(ctx, order); err != nil {
		return nil, err
	}
	resp, err := w.gateway.native(ctx, w.orderBody(order))
	if err != nil {
		return nil, providerError("native_pay", err)
	}
	if err := checkResponse(respCode(resp), nativeError(resp), resp != nil && resp.Response != nil); err != nil {
		return nil, providerError("native_pay", err)
	}
	if resp.Response.CodeUrl == "" {
		return nil, providerError("native_pay", fmt.Errorf("empty code URL"))
	}
	return &payment.PaymentResult{OrderID: order.OrderID, PayURL: resp.Response.CodeUrl}, nil
}

// JSAPIPay 创建 JSAPI 预支付单及前端调起参数。
func (w *WechatPay) JSAPIPay(ctx context.Context, order *payment.Order, openID string) (*payment.JSAPIResult, error) {
	if err := validateCall(ctx, order); err != nil {
		return nil, err
	}
	if openID == "" {
		return nil, fmt.Errorf("%w: open ID cannot be empty", payment.ErrInvalidRequest)
	}
	bm := w.orderBody(order)
	bm.Set("payer", gopay.BodyMap{"openid": openID})
	resp, err := w.gateway.jsapi(ctx, bm)
	if err != nil {
		return nil, providerError("jsapi_pay", err)
	}
	if err := checkResponse(respCode(resp), prepayError(resp), resp != nil && resp.Response != nil); err != nil {
		return nil, providerError("jsapi_pay", err)
	}
	if resp.Response.PrepayId == "" {
		return nil, providerError("jsapi_pay", fmt.Errorf("empty prepay ID"))
	}
	params, err := w.gateway.paySign(w.config.AppID, resp.Response.PrepayId)
	if err != nil || params == nil {
		return nil, providerError("jsapi_sign", err)
	}
	return &payment.JSAPIResult{
		AppID: params.AppId, Timestamp: params.TimeStamp, NonceStr: params.NonceStr,
		Package: params.Package, SignType: params.SignType, PaySign: params.PaySign,
	}, nil
}

// Query 查询微信支付订单。
func (w *WechatPay) Query(ctx context.Context, orderID string) (*payment.QueryResult, error) {
	if err := payment.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if err := payment.ValidateOrderID(orderID); err != nil {
		return nil, err
	}
	resp, err := w.gateway.query(ctx, orderID)
	if err != nil {
		return nil, providerError("query", err)
	}
	if err := checkResponse(respCode(resp), queryError(resp), resp != nil && resp.Response != nil); err != nil {
		return nil, providerError("query", err)
	}
	result := resp.Response
	if result.Amount == nil {
		return nil, providerError("query", fmt.Errorf("missing amount"))
	}
	status, err := mapPaymentStatus(result.TradeState)
	if err != nil {
		return nil, err
	}
	paidAt, err := parseWechatTime(result.SuccessTime)
	if err != nil {
		return nil, providerError("query", err)
	}
	return &payment.QueryResult{OrderID: result.OutTradeNo, TransactionID: result.TransactionId, Status: status, Amount: int64(result.Amount.Total), PaidAt: paidAt}, nil
}

// Refund 发起微信退款。
func (w *WechatPay) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResult, error) {
	if err := payment.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if err := payment.ValidateRefundRequest(req); err != nil {
		return nil, err
	}
	bm := gopay.BodyMap{
		"out_trade_no": req.OrderID, "out_refund_no": req.RefundID,
		"amount": gopay.BodyMap{"refund": req.Amount, "total": req.OriginalAmount, "currency": "CNY"},
	}
	if req.Reason != "" {
		bm.Set("reason", req.Reason)
	}
	if req.NotifyURL != "" {
		bm.Set("notify_url", req.NotifyURL)
	}
	resp, err := w.gateway.refund(ctx, bm)
	if err != nil {
		return nil, providerError("refund", err)
	}
	if err := checkResponse(respCode(resp), refundError(resp), resp != nil && resp.Response != nil); err != nil {
		return nil, providerError("refund", err)
	}
	status, err := mapRefundStatus(resp.Response.Status)
	if err != nil {
		return nil, err
	}
	refundAt, err := parseWechatTime(resp.Response.SuccessTime)
	if err != nil {
		return nil, providerError("refund", err)
	}
	return &payment.RefundResult{RefundID: resp.Response.OutRefundNo, TransactionID: resp.Response.RefundId, Status: status, RefundAt: refundAt}, nil
}

// Close 关闭微信支付订单。
func (w *WechatPay) Close(ctx context.Context, orderID string) error {
	if err := payment.ValidateContext(ctx); err != nil {
		return err
	}
	if err := payment.ValidateOrderID(orderID); err != nil {
		return err
	}
	resp, err := w.gateway.close(ctx, orderID)
	if err != nil {
		return providerError("close", err)
	}
	if err := checkResponse(respCode(resp), emptyError(resp), resp != nil); err != nil {
		return providerError("close", err)
	}
	return nil
}

func (w *WechatPay) orderBody(order *payment.Order) gopay.BodyMap {
	bm := gopay.BodyMap{
		"appid": w.config.AppID, "mchid": w.config.MchID, "description": order.Subject,
		"out_trade_no": order.OrderID, "notify_url": order.NotifyURL,
		"amount": gopay.BodyMap{"total": order.Amount, "currency": "CNY"},
	}
	if order.ExpireAt != nil {
		bm.Set("time_expire", order.ExpireAt.Format(time.RFC3339))
	}
	return bm
}
