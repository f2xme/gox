package wechat

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"

	"github.com/f2xme/gox/payment"
	wx "github.com/go-pay/gopay/wechat/v3"
)

func (w *WechatPay) setNotifyDefaults() {
	w.parseNotify = wx.V3ParseNotify
	w.verifyNotify = func(req *wx.V3NotifyReq, keys map[string]*rsa.PublicKey) error { return req.VerifySignByPKMap(keys) }
	w.decryptPay = func(req *wx.V3NotifyReq, key string) (*wx.V3DecryptPayResult, error) {
		return req.DecryptPayCipherText(key)
	}
	w.decryptRefund = func(req *wx.V3NotifyReq, key string) (*wx.V3DecryptRefundResult, error) {
		return req.DecryptRefundCipherText(key)
	}
}

// ParsePaymentNotification 验签、解密并解析微信支付通知。
func (w *WechatPay) ParsePaymentNotification(ctx context.Context, req *http.Request) (*payment.PaymentNotification, error) {
	if err := validateNotification(ctx, req); err != nil {
		return nil, err
	}
	notify, err := w.parseNotify(req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse wechat notification", payment.ErrInvalidRequest)
	}
	if err := w.verifyNotify(notify, w.gateway.publicKeys()); err != nil {
		return nil, fmt.Errorf("%w: wechat notification", payment.ErrInvalidSignature)
	}
	result, err := w.decryptPay(notify, w.config.APIV3Key)
	if err != nil || result == nil {
		return nil, fmt.Errorf("%w: decrypt wechat notification", payment.ErrInvalidRequest)
	}
	if result.Appid != w.config.AppID || result.Mchid != w.config.MchID {
		return nil, fmt.Errorf("%w: wechat notification merchant mismatch", payment.ErrInvalidSignature)
	}
	if result.Amount == nil {
		return nil, fmt.Errorf("%w: missing notification amount", payment.ErrInvalidRequest)
	}
	status, err := mapPaymentStatus(result.TradeState)
	if err != nil {
		return nil, err
	}
	paidAt, err := parseWechatTime(result.SuccessTime)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid notification time", payment.ErrInvalidRequest)
	}
	return &payment.PaymentNotification{Provider: payment.ProviderWechat, OrderID: result.OutTradeNo, TransactionID: result.TransactionId, Status: status, Amount: int64(result.Amount.Total), PaidAt: paidAt, Extra: map[string]any{"trade_type": result.TradeType, "trade_state_desc": result.TradeStateDesc}}, nil
}

// ParseRefundNotification 验签、解密并解析微信退款通知。
func (w *WechatPay) ParseRefundNotification(ctx context.Context, req *http.Request) (*payment.RefundNotification, error) {
	if err := validateNotification(ctx, req); err != nil {
		return nil, err
	}
	notify, err := w.parseNotify(req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse wechat refund notification", payment.ErrInvalidRequest)
	}
	if err := w.verifyNotify(notify, w.gateway.publicKeys()); err != nil {
		return nil, fmt.Errorf("%w: wechat refund notification", payment.ErrInvalidSignature)
	}
	result, err := w.decryptRefund(notify, w.config.APIV3Key)
	if err != nil || result == nil {
		return nil, fmt.Errorf("%w: decrypt wechat refund notification", payment.ErrInvalidRequest)
	}
	if result.Mchid != w.config.MchID {
		return nil, fmt.Errorf("%w: wechat refund notification merchant mismatch", payment.ErrInvalidSignature)
	}
	if result.Amount == nil {
		return nil, fmt.Errorf("%w: missing refund notification amount", payment.ErrInvalidRequest)
	}
	status, err := mapRefundStatus(result.RefundStatus)
	if err != nil {
		return nil, err
	}
	refundAt, err := parseWechatTime(result.SuccessTime)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid refund notification time", payment.ErrInvalidRequest)
	}
	return &payment.RefundNotification{Provider: payment.ProviderWechat, OrderID: result.OutTradeNo, TransactionID: result.TransactionId, RefundID: result.OutRefundNo, ProviderRefundID: result.RefundId, Status: status, Amount: int64(result.Amount.Refund), RefundAt: refundAt}, nil
}

// SuccessResponse 返回微信要求的成功回执。
func (w *WechatPay) SuccessResponse() payment.NotifyResponse {
	return payment.NotifyResponse{StatusCode: http.StatusOK, ContentType: "application/json; charset=utf-8", Body: []byte(`{"code":"SUCCESS","message":"成功"}`)}
}

func validateNotification(ctx context.Context, req *http.Request) error {
	if err := payment.ValidateContext(ctx); err != nil {
		return err
	}
	if req == nil || req.Body == nil {
		return fmt.Errorf("%w: empty wechat notification", payment.ErrInvalidRequest)
	}
	return nil
}
