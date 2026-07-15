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

// verifyNotification 用已加载公钥验签；自动模式下若序列号未登记则补拉一次后再验。
//
// 与 gopay 同步应答验签（serial miss → AutoVerifySign(false)）对齐。
// 若 key 已存在仍验签失败（签名本身错误），不会触发补拉。
//
// 返回错误始终可 errors.Is(..., ErrInvalidSignature)；补拉网络/配置失败时
// 额外保留内层 ErrGateway / ErrInvalidConfig，便于运维分流。
func (w *WechatPay) verifyNotification(notify *wx.V3NotifyReq) error {
	keys := w.gateway.publicKeys()
	err := w.verifyNotify(notify, keys)
	if err == nil {
		return nil
	}
	serial := notifySerial(notify)
	if serial == "" {
		return fmt.Errorf("%w: %w", payment.ErrInvalidSignature, err)
	}
	if _, ok := keys[serial]; ok {
		// 序列号已登记仍失败：签名错误，不补拉。
		return fmt.Errorf("%w: %w", payment.ErrInvalidSignature, err)
	}
	if refreshErr := w.gateway.ensurePublicKey(serial); refreshErr != nil {
		// 同时包裹签名失败与补拉根因（%w 保留 classify 结果）。
		return fmt.Errorf("%w: refresh platform cert for serial %s: %w", payment.ErrInvalidSignature, serial, refreshErr)
	}
	if retryErr := w.verifyNotify(notify, w.gateway.publicKeys()); retryErr != nil {
		return fmt.Errorf("%w: %w", payment.ErrInvalidSignature, retryErr)
	}
	return nil
}

func notifySerial(notify *wx.V3NotifyReq) string {
	if notify == nil || notify.SignInfo == nil {
		return ""
	}
	return notify.SignInfo.HeaderSerial
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
	if err := w.verifyNotification(notify); err != nil {
		return nil, fmt.Errorf("%w: wechat notification: %w", payment.ErrInvalidSignature, err)
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
	if err := w.verifyNotification(notify); err != nil {
		return nil, fmt.Errorf("%w: wechat refund notification: %w", payment.ErrInvalidSignature, err)
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
