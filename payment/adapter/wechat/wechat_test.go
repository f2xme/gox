package wechat

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/f2xme/gox/payment"
	"github.com/go-pay/gopay"
	wx "github.com/go-pay/gopay/wechat/v3"
)

type fakeGateway struct {
	body    gopay.BodyMap
	nativeR *wx.NativeRsp
	jsapiR  *wx.PrepayRsp
	queryR  *wx.QueryOrderRsp
	refundR *wx.RefundRsp
	closeR  *wx.EmptyRsp
	err     error
	keys    map[string]*rsa.PublicKey
}

func (f *fakeGateway) native(_ context.Context, bm gopay.BodyMap) (*wx.NativeRsp, error) {
	f.body = bm
	return f.nativeR, f.err
}
func (f *fakeGateway) jsapi(_ context.Context, bm gopay.BodyMap) (*wx.PrepayRsp, error) {
	f.body = bm
	return f.jsapiR, f.err
}
func (f *fakeGateway) paySign(_, _ string) (*wx.JSAPIPayParams, error) {
	return &wx.JSAPIPayParams{AppId: "app", TimeStamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s"}, f.err
}
func (f *fakeGateway) query(_ context.Context, _ string) (*wx.QueryOrderRsp, error) {
	return f.queryR, f.err
}
func (f *fakeGateway) refund(_ context.Context, bm gopay.BodyMap) (*wx.RefundRsp, error) {
	f.body = bm
	return f.refundR, f.err
}
func (f *fakeGateway) close(_ context.Context, _ string) (*wx.EmptyRsp, error) {
	return f.closeR, f.err
}
func (f *fakeGateway) publicKeys() map[string]*rsa.PublicKey { return f.keys }

func testConfig() Config {
	return Config{AppID: "app", OAuthAppSecret: "secret", MchID: "mch", APIV3Key: "12345678901234567890123456789012"}
}
func testOrder() *payment.Order {
	return &payment.Order{OrderID: "order-1", Amount: 123, Subject: "subject", NotifyURL: "https://example.com/notify"}
}

func TestWechatPayImplementsInterfaces(t *testing.T) {
	var _ payment.Payment = (*WechatPay)(nil)
	var _ payment.PaymentNotifier = (*WechatPay)(nil)
	var _ payment.RefundNotifier = (*WechatPay)(nil)
}

func TestPayAndJSAPI(t *testing.T) {
	gw := &fakeGateway{nativeR: &wx.NativeRsp{Code: wx.Success, Response: &wx.Native{CodeUrl: "weixin://pay"}}, jsapiR: &wx.PrepayRsp{Code: wx.Success, Response: &wx.Prepay{PrepayId: "p"}}}
	w := newWithGateway(testConfig(), gw)
	got, err := w.Pay(context.Background(), testOrder())
	if err != nil || got.PayURL != "weixin://pay" {
		t.Fatalf("Pay() = %#v, %v", got, err)
	}
	if gw.body.GetString("appid") != "app" || gw.body.GetString("mchid") != "mch" {
		t.Fatalf("body = %#v", gw.body)
	}
	w.config.OAuthAppSecret = ""
	js, err := w.JSAPIPay(context.Background(), testOrder(), "openid")
	if err != nil || js.Package != "prepay_id=p" || js.PaySign != "s" {
		t.Fatalf("JSAPIPay() = %#v, %v", js, err)
	}
}

func TestQueryRefundClose(t *testing.T) {
	gw := &fakeGateway{
		queryR:  &wx.QueryOrderRsp{Code: wx.Success, Response: &wx.QueryOrder{OutTradeNo: "order-1", TransactionId: "tx", TradeState: "SUCCESS", SuccessTime: "2026-07-12T12:00:00+08:00", Amount: &wx.Amount{Total: 123}}},
		refundR: &wx.RefundRsp{Code: wx.Success, Response: &wx.RefundOrderResponse{OutRefundNo: "refund-1", RefundId: "wx-refund", Status: "PROCESSING"}},
		closeR:  &wx.EmptyRsp{Code: wx.Success},
	}
	w := newWithGateway(testConfig(), gw)
	query, err := w.Query(context.Background(), "order-1")
	if err != nil || query.Status != payment.PaymentStatusSuccess || query.Amount != 123 || query.PaidAt == nil {
		t.Fatalf("Query() = %#v, %v", query, err)
	}
	refund, err := w.Refund(context.Background(), &payment.RefundRequest{OrderID: "order-1", RefundID: "refund-1", Amount: 23, OriginalAmount: 123})
	if err != nil || refund.Status != payment.RefundStatusPending || refund.TransactionID != "wx-refund" {
		t.Fatalf("Refund() = %#v, %v", refund, err)
	}
	if err := w.Close(context.Background(), "order-1"); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestGatewayAndStatusErrors(t *testing.T) {
	w := newWithGateway(testConfig(), &fakeGateway{nativeR: &wx.NativeRsp{Code: http.StatusBadRequest, ErrResponse: wx.ErrResponse{Code: "INVALID_REQUEST", Message: "bad"}}})
	_, err := w.Pay(context.Background(), testOrder())
	if !errors.Is(err, payment.ErrGateway) {
		t.Fatalf("Pay() error = %v", err)
	}
	var providerErr *payment.ProviderError
	if !errors.As(err, &providerErr) || providerErr.Code != "INVALID_REQUEST" {
		t.Fatalf("provider error = %#v", providerErr)
	}
	if _, err := mapPaymentStatus("NEW_STATE"); !errors.Is(err, payment.ErrUnknownStatus) {
		t.Fatalf("status error = %v", err)
	}
	w = newWithGateway(testConfig(), &fakeGateway{queryR: &wx.QueryOrderRsp{Code: wx.Success, Response: &wx.QueryOrder{OutTradeNo: "order-1", TradeState: "SUCCESS"}}})
	if _, err := w.Query(context.Background(), "order-1"); !errors.Is(err, payment.ErrGateway) {
		t.Fatalf("missing amount error = %v, want ErrGateway", err)
	}
}

func TestOAuth(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("code") != "code" {
			t.Errorf("code = %q", r.URL.Query().Get("code"))
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"openid":"openid-1"}`)), Header: make(http.Header)}, nil
	})
	w := newWithGateway(testConfig(), &fakeGateway{})
	w.oauthClient = &http.Client{Transport: transport}
	w.oauthAuthURL = defaultOAuthAuthURL
	w.oauthTokenURL = "https://api.example/token"
	authURL, err := w.OAuthURL("https://merchant.example/callback", "state")
	if err != nil || authURL == "" {
		t.Fatalf("OAuthURL() = %q, %v", authURL, err)
	}
	openID, err := w.ExchangeOAuthCode(context.Background(), "code")
	if err != nil || openID != "openid-1" {
		t.Fatalf("ExchangeOAuthCode() = %q, %v", openID, err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestNotifications(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	w := newWithGateway(testConfig(), &fakeGateway{keys: map[string]*rsa.PublicKey{"PUB_KEY_ID_TEST": &privateKey.PublicKey}})
	payPlain := `{"appid":"app","mchid":"mch","out_trade_no":"order-1","transaction_id":"tx","trade_state":"SUCCESS","success_time":"2026-07-12T12:00:00+08:00","amount":{"total":123}}`
	req := signedNotifyRequest(t, privateKey, testConfig().APIV3Key, payPlain)
	pay, err := w.ParsePaymentNotification(context.Background(), req)
	if err != nil || pay.Amount != 123 || pay.Status != payment.PaymentStatusSuccess {
		t.Fatalf("payment notification = %#v, %v", pay, err)
	}
	refundPlain := `{"mchid":"mch","out_trade_no":"order-1","transaction_id":"tx","out_refund_no":"refund-1","refund_id":"wx-refund","refund_status":"SUCCESS","success_time":"2026-07-12T12:00:00+08:00","amount":{"refund":23}}`
	req = signedNotifyRequest(t, privateKey, testConfig().APIV3Key, refundPlain)
	refund, err := w.ParseRefundNotification(context.Background(), req)
	if err != nil || refund.Amount != 23 || refund.Status != payment.RefundStatusSuccess {
		t.Fatalf("refund notification = %#v, %v", refund, err)
	}
	if string(w.SuccessResponse().Body) != `{"code":"SUCCESS","message":"成功"}` {
		t.Fatal("unexpected success response")
	}
}

func signedNotifyRequest(t *testing.T, key *rsa.PrivateKey, apiKey, plain string) *http.Request {
	t.Helper()
	block, err := aes.NewCipher([]byte(apiKey))
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	nonce, additional := "0123456789ab", "resource"
	ciphertext := gcm.Seal(nil, []byte(nonce), []byte(plain), []byte(additional))
	body, err := json.Marshal(map[string]any{"id": "event", "resource": map[string]string{"algorithm": "AEAD_AES_256_GCM", "ciphertext": base64.StdEncoding.EncodeToString(ciphertext), "associated_data": additional, "nonce": nonce}})
	if err != nil {
		t.Fatal(err)
	}
	timestamp, headerNonce := "1783838400", "notify-nonce"
	digest := sha256.Sum256([]byte(timestamp + "\n" + headerNonce + "\n" + string(body) + "\n"))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	req.Header.Set("Wechatpay-Timestamp", timestamp)
	req.Header.Set("Wechatpay-Nonce", headerNonce)
	req.Header.Set("Wechatpay-Serial", "PUB_KEY_ID_TEST")
	req.Header.Set("Wechatpay-Signature", base64.StdEncoding.EncodeToString(signature))
	return req
}

func TestParseWechatTime(t *testing.T) {
	got, err := parseWechatTime("2026-07-12T12:00:00+08:00")
	if err != nil || got.Equal(time.Time{}) {
		t.Fatalf("parseWechatTime() = %v, %v", got, err)
	}
}
