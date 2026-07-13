package alipay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/f2xme/gox/payment"
	"github.com/go-pay/gopay"
	aliyun "github.com/go-pay/gopay/alipay"
)

var (
	_ payment.Payment         = (*Alipay)(nil)
	_ payment.PaymentNotifier = (*Alipay)(nil)
)

type fakeGateway struct {
	precreateResp *aliyun.TradePrecreateResponse
	wapURL        string
	queryResp     *aliyun.TradeQueryResponse
	refundResp    *aliyun.TradeRefundResponse
	closeResp     *aliyun.TradeCloseResponse
	err           error
	lastBody      gopay.BodyMap
}

func (f *fakeGateway) precreate(_ context.Context, bm gopay.BodyMap) (*aliyun.TradePrecreateResponse, error) {
	f.lastBody = bm
	return f.precreateResp, f.err
}
func (f *fakeGateway) wapPay(_ context.Context, bm gopay.BodyMap) (string, error) {
	f.lastBody = bm
	return f.wapURL, f.err
}
func (f *fakeGateway) query(_ context.Context, bm gopay.BodyMap) (*aliyun.TradeQueryResponse, error) {
	f.lastBody = bm
	return f.queryResp, f.err
}
func (f *fakeGateway) refund(_ context.Context, bm gopay.BodyMap) (*aliyun.TradeRefundResponse, error) {
	f.lastBody = bm
	return f.refundResp, f.err
}
func (f *fakeGateway) close(_ context.Context, bm gopay.BodyMap) (*aliyun.TradeCloseResponse, error) {
	f.lastBody = bm
	return f.closeResp, f.err
}

func TestPayAndWAP(t *testing.T) {
	gw := &fakeGateway{
		precreateResp: &aliyun.TradePrecreateResponse{Response: &aliyun.TradePrecreate{OutTradeNo: "o1", QrCode: "https://qr.example/o1"}},
		wapURL:        "https://openapi.alipay.com/gateway.do?sign=x",
	}
	client := newWithGateway(Config{}, gw)
	expiresAt := time.Date(2030, 7, 12, 4, 0, 0, 0, time.UTC)
	order := &payment.Order{OrderID: "o1", Amount: 123, Subject: "商品", NotifyURL: "https://example.com/notify", ReturnURL: "https://example.com/return", ExpireAt: &expiresAt}

	result, err := client.Pay(context.Background(), order)
	if err != nil {
		t.Fatal(err)
	}
	if result.PayURL != "https://qr.example/o1" || gw.lastBody.GetString("total_amount") != "1.23" || gw.lastBody.GetString("time_expire") != "2030-07-12 12:00:00" {
		t.Fatalf("unexpected pay result=%#v body=%v", result, gw.lastBody)
	}

	wap, err := client.WAPPay(context.Background(), order)
	if err != nil {
		t.Fatal(err)
	}
	if wap.URL != gw.wapURL || gw.lastBody.GetString("return_url") != order.ReturnURL {
		t.Fatalf("unexpected wap result=%#v body=%v", wap, gw.lastBody)
	}
}

func TestQueryRefundClose(t *testing.T) {
	gw := &fakeGateway{
		queryResp: &aliyun.TradeQueryResponse{Response: &aliyun.TradeQuery{
			OutTradeNo: "o1", TradeNo: "trade1", TradeStatus: "TRADE_SUCCESS", TotalAmount: "12.34", SendPayDate: "2026-07-12 12:00:00",
		}},
		refundResp: &aliyun.TradeRefundResponse{Response: &aliyun.TradeRefund{TradeNo: "trade1", OutTradeNo: "o1", RefundSettlementId: "refund1", FundChange: "Y", GmtRefundPay: "2026-07-12 12:01:00"}},
		closeResp:  &aliyun.TradeCloseResponse{Response: &aliyun.TradeClose{OutTradeNo: "o1"}},
	}
	client := newWithGateway(Config{}, gw)

	query, err := client.Query(context.Background(), "o1")
	if err != nil {
		t.Fatal(err)
	}
	if query.Status != payment.PaymentStatusSuccess || query.Amount != 1234 || query.PaidAt == nil {
		t.Fatalf("unexpected query: %#v", query)
	}

	refund, err := client.Refund(context.Background(), &payment.RefundRequest{OrderID: "o1", RefundID: "r1", Amount: 100, OriginalAmount: 1234})
	if err != nil {
		t.Fatal(err)
	}
	if refund.Status != payment.RefundStatusSuccess || refund.TransactionID != "refund1" || gw.lastBody.GetString("refund_amount") != "1.00" {
		t.Fatalf("unexpected refund=%#v body=%v", refund, gw.lastBody)
	}
	if err := client.Close(context.Background(), "o1"); err != nil {
		t.Fatal(err)
	}
}

func TestGatewayErrorPreservesCause(t *testing.T) {
	cause := context.DeadlineExceeded
	client := newWithGateway(Config{}, &fakeGateway{err: cause})
	_, err := client.Pay(context.Background(), &payment.Order{OrderID: "o1", Amount: 1, Subject: "商品", NotifyURL: "https://example.com/notify"})
	if !errors.Is(err, payment.ErrGateway) || !errors.Is(err, cause) {
		t.Fatalf("expected gateway and cause, got %v", err)
	}
}

func TestParsePaymentNotification(t *testing.T) {
	form := url.Values{
		"app_id":       {"app1"},
		"seller_id":    {"seller1"},
		"out_trade_no": {"o1"},
		"trade_no":     {"trade1"},
		"trade_status": {"TRADE_SUCCESS"},
		"total_amount": {"1.23"},
		"gmt_payment":  {"2026-07-12 12:00:00"},
		"sign":         {"signature"},
		"sign_type":    {"RSA2"},
	}
	req := httptest.NewRequest("POST", "https://example.com/notify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := newWithGateway(Config{AppID: "app1", SellerID: "seller1"}, &fakeGateway{})
	client.verifyNotify = func(string, any) (bool, error) { return true, nil }

	notify, err := client.ParsePaymentNotification(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if notify.OrderID != "o1" || notify.Amount != 123 || notify.Status != payment.PaymentStatusSuccess || notify.PaidAt == nil {
		t.Fatalf("unexpected notification: %#v", notify)
	}
	if _, ok := notify.Extra["sign"]; ok {
		t.Fatal("signature must not be copied to Extra")
	}
}

func TestParsePaymentNotificationRejectsSignature(t *testing.T) {
	req := httptest.NewRequest("POST", "https://example.com/notify", strings.NewReader("app_id=app1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := newWithGateway(Config{}, &fakeGateway{})
	client.verifyNotify = func(string, any) (bool, error) { return false, nil }
	_, err := client.ParsePaymentNotification(context.Background(), req)
	if !errors.Is(err, payment.ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestAmountConversion(t *testing.T) {
	tests := map[string]int64{"0": 0, "1": 100, "1.2": 120, "1.23": 123}
	for input, want := range tests {
		got, err := yuanToCents(input)
		if err != nil || got != want {
			t.Fatalf("yuanToCents(%q) = %d, %v; want %d", input, got, err, want)
		}
	}
	if _, err := yuanToCents("1.234"); err == nil {
		t.Fatal("expected precision error")
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	if _, err := New(Config{}); !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestNewRejectsInvalidPublicKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
	_, err = New(Config{
		AppID:           "app",
		SellerID:        "seller",
		PrivateKey:      string(privatePEM),
		AlipayPublicKey: "not-a-public-key",
	})
	if !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("New() error = %v, want ErrInvalidConfig", err)
	}
}
