package alipay

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/f2xme/gox/payment"
	"github.com/go-pay/crypto/xpem"
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
	publicKey, normalizeErr := normalizeNotifyPublicKey(publicKey)
	return func(value any) (bool, error) {
		if normalizeErr != nil {
			return false, normalizeErr
		}
		return aliyun.VerifySign(publicKey, value)
	}
}

func normalizeNotifyPublicKey(publicKey string) (string, error) {
	if block, _ := pem.Decode([]byte(publicKey)); block == nil {
		return publicKey, nil
	}
	key, err := xpem.DecodePublicKey([]byte(publicKey))
	if err != nil {
		return "", fmt.Errorf("decode alipay public key: %w", err)
	}
	if key == nil {
		return "", fmt.Errorf("decode alipay public key: empty public key")
	}
	der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", fmt.Errorf("encode alipay public key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(der), nil
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
	form := req.PostForm
	if len(form) == 0 {
		return nil, fmt.Errorf("%w: empty alipay notification", payment.ErrInvalidRequest)
	}
	bm, err := aliyun.ParseNotifyByURLValues(form)
	if err != nil {
		return nil, fmt.Errorf("%w: parse alipay notification: %v", payment.ErrInvalidRequest, err)
	}
	ok, err := a.verifyNotify(bm)
	if err != nil {
		return nil, fmt.Errorf("%w: verify alipay notification: %w", payment.ErrInvalidSignature, err)
	}
	if !ok {
		return nil, fmt.Errorf("%w: alipay notification", payment.ErrInvalidSignature)
	}
	if got := form.Get("app_id"); got != a.config.AppID {
		return nil, fmt.Errorf("%w: alipay notification app_id mismatch: got %q, want %q", payment.ErrInvalidSignature, got, a.config.AppID)
	}
	if got := form.Get("seller_id"); got != a.config.SellerID {
		return nil, fmt.Errorf("%w: alipay notification seller_id mismatch: got %q, want %q", payment.ErrInvalidSignature, got, a.config.SellerID)
	}
	status, err := mapPaymentStatus(form.Get("trade_status"))
	if err != nil {
		return nil, err
	}
	amount, err := yuanToCents(form.Get("total_amount"))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid notification amount", payment.ErrInvalidRequest)
	}
	paidAt, err := parseAlipayTime(form.Get("gmt_payment"))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid notification time", payment.ErrInvalidRequest)
	}
	extra := make(map[string]any, len(form))
	for key, values := range form {
		if key != "sign" && len(values) > 0 {
			extra[key] = values[0]
		}
	}
	return &payment.PaymentNotification{
		Provider:      payment.ProviderAlipay,
		OrderID:       form.Get("out_trade_no"),
		TransactionID: form.Get("trade_no"),
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
