package alipay

import (
	"context"
	"fmt"
	"net/http"

	"github.com/f2xme/gox/payment"
	aliyun "github.com/go-pay/gopay/alipay"
)

// notifyVerifier 异步通知验签函数，测试可注入。
type notifyVerifier func(value any) (bool, error)

// notifyVerifyMode 标识异步验签所用材料类型。
type notifyVerifyMode string

const (
	notifyVerifyModeKey  notifyVerifyMode = "key"
	notifyVerifyModeCert notifyVerifyMode = "cert"
)

func verifySignKey(publicKey string) notifyVerifier {
	return func(value any) (bool, error) {
		return aliyun.VerifySign(publicKey, value)
	}
}

func verifySignCert(publicCert []byte) notifyVerifier {
	return func(value any) (bool, error) {
		return aliyun.VerifySignWithCert(publicCert, value)
	}
}

// resolveNotifyVerifyMode 返回配置对应的异步验签模式。
func resolveNotifyVerifyMode(config Config) notifyVerifyMode {
	if config.useCertMode() {
		return notifyVerifyModeCert
	}
	return notifyVerifyModeKey
}

func newNotifyVerifier(config Config) notifyVerifier {
	if resolveNotifyVerifyMode(config) == notifyVerifyModeCert {
		return verifySignCert([]byte(config.AlipayPublicCert))
	}
	return verifySignKey(config.AlipayPublicKey)
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
	ok, err := a.verifyNotify(bm)
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
