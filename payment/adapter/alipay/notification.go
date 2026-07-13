package alipay

import (
	"context"
	"fmt"
	"net/http"

	"github.com/f2xme/gox/payment"
	aliyun "github.com/go-pay/gopay/alipay"
)

func verifySign(publicKey string, value any) (bool, error) {
	return aliyun.VerifySign(publicKey, value)
}

// ParsePaymentNotification 解析并验证支付宝支付通知。
func (a *Alipay) ParsePaymentNotification(ctx context.Context, req *http.Request) (*payment.PaymentNotification, error) {
	if err := payment.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if req == nil || req.Body == nil {
		return nil, fmt.Errorf("%w: empty alipay notification", payment.ErrInvalidRequest)
	}
	if err := req.ParseForm(); err != nil {
		return nil, fmt.Errorf("%w: parse alipay notification: %v", payment.ErrInvalidRequest, err)
	}
	if len(req.Form) == 0 {
		return nil, fmt.Errorf("%w: empty alipay notification", payment.ErrInvalidRequest)
	}
	bm, err := aliyun.ParseNotifyByURLValues(req.Form)
	if err != nil {
		return nil, fmt.Errorf("%w: parse alipay notification: %v", payment.ErrInvalidRequest, err)
	}
	ok, err := a.verifyNotify(a.config.AlipayPublicKey, bm)
	if err != nil || !ok {
		return nil, fmt.Errorf("%w: alipay notification", payment.ErrInvalidSignature)
	}
	if req.Form.Get("app_id") != a.config.AppID || req.Form.Get("seller_id") != a.config.SellerID {
		return nil, fmt.Errorf("%w: alipay notification merchant mismatch", payment.ErrInvalidSignature)
	}
	status, err := mapPaymentStatus(req.Form.Get("trade_status"))
	if err != nil {
		return nil, err
	}
	amount, err := yuanToCents(req.Form.Get("total_amount"))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid notification amount", payment.ErrInvalidRequest)
	}
	paidAt, err := parseAlipayTime(req.Form.Get("gmt_payment"))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid notification time", payment.ErrInvalidRequest)
	}
	extra := make(map[string]any, len(req.Form))
	for key, values := range req.Form {
		if key != "sign" && len(values) > 0 {
			extra[key] = values[0]
		}
	}
	return &payment.PaymentNotification{
		Provider:      payment.ProviderAlipay,
		OrderID:       req.Form.Get("out_trade_no"),
		TransactionID: req.Form.Get("trade_no"),
		Status:        status,
		Amount:        amount,
		PaidAt:        paidAt,
		Extra:         extra,
	}, nil
}

// SuccessResponse 返回支付宝成功回执。
func (a *Alipay) SuccessResponse() payment.NotifyResponse {
	return payment.NotifyResponse{
		StatusCode:  http.StatusOK,
		ContentType: "text/plain; charset=utf-8",
		Body:        []byte("success"),
	}
}
