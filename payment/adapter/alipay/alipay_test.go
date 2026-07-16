package alipay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
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

// 测试密钥/证书只生成一次，避免重复 2048-bit RSA 拖慢套件。
var (
	testMaterialOnce sync.Once
	testPrivatePEM   string
	testPublicPEM    string
	testCertPEM      string
)

func initTestMaterial(t *testing.T) {
	t.Helper()
	testMaterialOnce.Do(func() {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}
		privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			panic(err)
		}
		testPrivatePEM = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER}))

		publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			panic(err)
		}
		testPublicPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}))

		tmpl := &x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject:      pkix.Name{CommonName: "test-alipay-cert"},
			NotBefore:    time.Now().Add(-time.Hour),
			NotAfter:     time.Now().Add(24 * time.Hour),
			KeyUsage:     x509.KeyUsageDigitalSignature,
		}
		certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &privateKey.PublicKey, privateKey)
		if err != nil {
			panic(err)
		}
		testCertPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}))
	})
}

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
	client.verifyNotify = func(any) (bool, error) { return true, nil }

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
	client.verifyNotify = func(any) (bool, error) { return false, nil }
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

func TestValidateConfigModes(t *testing.T) {
	privatePEM := mustPrivatePEM(t)
	validCert := mustSelfSignedCertPEM(t)
	publicPEM := mustPublicPEM(t)

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "key mode ok",
			config: Config{
				AppID: "app", SellerID: "seller", PrivateKey: privatePEM,
				AlipayPublicKey: publicPEM,
			},
		},
		{
			name: "cert mode ok",
			config: Config{
				AppID: "app", SellerID: "seller", PrivateKey: privatePEM,
				AppPublicCert: validCert, AlipayRootCert: validCert, AlipayPublicCert: validCert,
			},
		},
		{
			name: "missing both modes",
			config: Config{
				AppID: "app", SellerID: "seller", PrivateKey: privatePEM,
			},
			wantErr: true,
		},
		{
			name: "partial cert rejects",
			config: Config{
				AppID: "app", SellerID: "seller", PrivateKey: privatePEM,
				AppPublicCert: validCert, AlipayPublicKey: publicPEM,
			},
			wantErr: true,
		},
		{
			name: "partial cert without public key rejects",
			config: Config{
				AppID: "app", SellerID: "seller", PrivateKey: privatePEM,
				AppPublicCert: validCert, AlipayRootCert: validCert,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.wantErr {
				if !errors.Is(err, payment.ErrInvalidConfig) {
					t.Fatalf("validateConfig() error = %v, want ErrInvalidConfig", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateConfig() unexpected error: %v", err)
			}
		})
	}
}

func TestNewRejectsInvalidPublicKey(t *testing.T) {
	_, err := New(Config{
		AppID:           "app",
		SellerID:        "seller",
		PrivateKey:      mustPrivatePEM(t),
		AlipayPublicKey: "not-a-public-key",
	})
	if !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("New() error = %v, want ErrInvalidConfig", err)
	}
}

func TestNewRejectsUnsupportedPEMType(t *testing.T) {
	// DecodePublicKey 对不受支持的 PEM type 返回 (nil, nil)，须被 New 拒绝。
	unsupported := string(pem.EncodeToMemory(&pem.Block{Type: "FOO", Bytes: []byte("not-a-key")}))
	_, err := New(Config{
		AppID:           "app",
		SellerID:        "seller",
		PrivateKey:      mustPrivatePEM(t),
		AlipayPublicKey: unsupported,
	})
	if !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("New() error = %v, want ErrInvalidConfig", err)
	}
	if !strings.Contains(err.Error(), "empty public key") {
		t.Fatalf("New() error = %v, want empty public key cause", err)
	}
}

func TestNewRejectsInvalidCert(t *testing.T) {
	_, err := New(Config{
		AppID:            "app",
		SellerID:         "seller",
		PrivateKey:       mustPrivatePEM(t),
		AppPublicCert:    "not-a-cert",
		AlipayRootCert:   "not-a-cert",
		AlipayPublicCert: "not-a-cert",
	})
	if !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("New() error = %v, want ErrInvalidConfig", err)
	}
}

func TestNewAcceptsCertMode(t *testing.T) {
	certPEM := mustSelfSignedCertPEM(t)
	client, err := New(Config{
		AppID:            "app",
		SellerID:         "seller",
		PrivateKey:       mustPrivatePEM(t),
		AppPublicCert:    certPEM,
		AlipayRootCert:   certPEM,
		AlipayPublicCert: certPEM,
	})
	if err != nil {
		t.Fatalf("New() cert mode error = %v", err)
	}
	if client == nil {
		t.Fatal("New() returned nil client")
	}
	if !client.config.useCertMode() {
		t.Fatal("expected cert mode")
	}
}

func TestValidateAESKey(t *testing.T) {
	// gopay 使用字符串原样字节作 AES key：长度须 16/24/32；空/纯空白表示不启用。
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{name: "empty", key: ""},
		{name: "whitespace only", key: "   "},
		{name: "16 bytes", key: "0123456789abcdef"},
		{name: "24 bytes open-platform style", key: "KvKUTqSVZX2fUgmxnFyMaQ=="},
		{name: "32 bytes", key: "0123456789abcdef0123456789abcdef"},
		{name: "too short", key: "short", wantErr: true},
		{name: "15 bytes", key: "0123456789abcde", wantErr: true},
		{name: "25 bytes", key: "0123456789abcdef012345678", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAESKey(tt.key)
			if tt.wantErr {
				if !errors.Is(err, payment.ErrInvalidConfig) {
					t.Fatalf("validateAESKey() = %v, want ErrInvalidConfig", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateAESKey() = %v", err)
			}
		})
	}
}

func TestNewAcceptsAESKey(t *testing.T) {
	// 合法长度 AESKey：New 成功并保留配置（加密路径由 gopay 在请求期执行）。
	client, err := New(Config{
		AppID:           "app",
		SellerID:        "seller",
		PrivateKey:      mustPrivatePEM(t),
		AlipayPublicKey: mustPublicPEM(t),
		AESKey:          "KvKUTqSVZX2fUgmxnFyMaQ==", // 24 chars → AES-192 key material in gopay
		Environment:     EnvSandbox,
	})
	if err != nil {
		t.Fatalf("New() with AESKey error = %v", err)
	}
	if client == nil || client.config.AESKey == "" {
		t.Fatalf("expected AESKey retained on client config, got %#v", client)
	}
}

func TestNewRejectsInvalidAESKey(t *testing.T) {
	_, err := New(Config{
		AppID:           "app",
		SellerID:        "seller",
		PrivateKey:      mustPrivatePEM(t),
		AlipayPublicKey: mustPublicPEM(t),
		AESKey:          "not-valid-aes-key-len",
		Environment:     EnvSandbox,
	})
	if !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("New() invalid AESKey error = %v, want ErrInvalidConfig", err)
	}
}

func TestNewAcceptsEmptyAESKey(t *testing.T) {
	client, err := New(Config{
		AppID:           "app",
		SellerID:        "seller",
		PrivateKey:      mustPrivatePEM(t),
		AlipayPublicKey: mustPublicPEM(t),
		AESKey:          "",
		Environment:     EnvSandbox,
	})
	if err != nil {
		t.Fatalf("New() empty AESKey error = %v", err)
	}
	if client.config.AESKey != "" {
		t.Fatalf("expected empty AESKey, got %q", client.config.AESKey)
	}
}

func TestUseCertMode(t *testing.T) {
	if (Config{}).useCertMode() {
		t.Fatal("empty config should not use cert mode")
	}
	if (Config{AppPublicCert: "a"}).useCertMode() {
		t.Fatal("partial cert should not use cert mode")
	}
	if !(Config{AppPublicCert: "a", AlipayRootCert: "b", AlipayPublicCert: "c"}).useCertMode() {
		t.Fatal("full cert set should use cert mode")
	}
}

func TestResolveNotifyVerifyMode(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want notifyVerifyMode
	}{
		{
			name: "key mode",
			cfg:  Config{AlipayPublicKey: "pk"},
			want: notifyVerifyModeKey,
		},
		{
			name: "cert mode takes precedence when full set present",
			cfg: Config{
				AlipayPublicKey:  "pk",
				AppPublicCert:    "a",
				AlipayRootCert:   "b",
				AlipayPublicCert: "c",
			},
			want: notifyVerifyModeCert,
		},
		{
			name: "partial cert falls back to key mode",
			cfg: Config{
				AlipayPublicKey: "pk",
				AppPublicCert:   "a",
			},
			want: notifyVerifyModeKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveNotifyVerifyMode(tt.cfg); got != tt.want {
				t.Fatalf("resolveNotifyVerifyMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewNotifyVerifierRejectsBadSignature(t *testing.T) {
	bm := gopay.BodyMap{
		"out_trade_no": "o1",
		"sign":         "not-a-real-signature",
		"sign_type":    "RSA2",
	}

	keyVerifier := newNotifyVerifier(Config{AlipayPublicKey: mustPublicPEM(t)})
	ok, err := keyVerifier(bm)
	if err == nil && ok {
		t.Fatal("key verifier accepted invalid signature")
	}

	certVerifier := newNotifyVerifier(Config{
		AppPublicCert: mustSelfSignedCertPEM(t), AlipayRootCert: mustSelfSignedCertPEM(t),
		AlipayPublicCert: mustSelfSignedCertPEM(t),
	})
	ok, err = certVerifier(bm)
	if err == nil && ok {
		t.Fatal("cert verifier accepted invalid signature")
	}
}

func TestEnvironmentResolution(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want Environment
		url  string
	}{
		{
			name: "zero value defaults to sandbox",
			cfg:  Config{},
			want: EnvSandbox,
			url:  gatewayURLSandbox,
		},
		{
			name: "Production true maps to production",
			cfg:  Config{Production: true},
			want: EnvProduction,
			url:  gatewayURLProduction,
		},
		{
			name: "Environment sandbox overrides Production true",
			cfg:  Config{Environment: EnvSandbox, Production: true},
			want: EnvSandbox,
			url:  gatewayURLSandbox,
		},
		{
			name: "Environment production overrides Production false",
			cfg:  Config{Environment: EnvProduction, Production: false},
			want: EnvProduction,
			url:  gatewayURLProduction,
		},
		{
			name: "invalid Environment falls back to Production false -> sandbox",
			cfg:  Config{Environment: "staging", Production: false},
			want: EnvSandbox,
			url:  gatewayURLSandbox,
		},
		{
			name: "invalid Environment falls back to Production true -> production",
			cfg:  Config{Environment: "staging", Production: true},
			want: EnvProduction,
			url:  gatewayURLProduction,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.ResolveEnvironment(); got != tt.want {
				t.Fatalf("ResolveEnvironment() = %q, want %q", got, tt.want)
			}
			if got := tt.cfg.IsSandbox(); got != (tt.want == EnvSandbox) {
				t.Fatalf("IsSandbox() = %v, want %v", got, tt.want == EnvSandbox)
			}
			if got := tt.cfg.IsProduction(); got != (tt.want == EnvProduction) {
				t.Fatalf("IsProduction() = %v, want %v", got, tt.want == EnvProduction)
			}
			if got := tt.cfg.GatewayBaseURL(); got != tt.url {
				t.Fatalf("GatewayBaseURL() = %q, want %q", got, tt.url)
			}
		})
	}
}

func TestNewRejectsInvalidEnvironment(t *testing.T) {
	err := validateConfig(Config{
		AppID: "app", SellerID: "seller", PrivateKey: mustPrivatePEM(t),
		AlipayPublicKey: mustPublicPEM(t),
		Environment:     "staging",
	})
	if !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("validateConfig() error = %v, want ErrInvalidConfig", err)
	}
}

func TestNewSandboxAndProduction(t *testing.T) {
	base := Config{
		AppID: "app", SellerID: "seller",
		PrivateKey: mustPrivatePEM(t), AlipayPublicKey: mustPublicPEM(t),
	}

	sandboxCfg := base
	sandboxCfg.Environment = EnvSandbox
	sandbox, err := New(sandboxCfg)
	if err != nil {
		t.Fatalf("New sandbox: %v", err)
	}
	if !sandbox.IsSandbox() || sandbox.IsProduction() || sandbox.Environment() != EnvSandbox {
		t.Fatalf("expected sandbox client, got env=%q production=%v", sandbox.Environment(), sandbox.IsProduction())
	}
	if sandbox.GatewayBaseURL() != gatewayURLSandbox {
		t.Fatalf("sandbox gateway = %q", sandbox.GatewayBaseURL())
	}

	prodCfg := base
	prodCfg.Environment = EnvProduction
	prod, err := New(prodCfg)
	if err != nil {
		t.Fatalf("New production: %v", err)
	}
	if prod.IsSandbox() || !prod.IsProduction() || prod.Environment() != EnvProduction {
		t.Fatalf("expected production client, got env=%q production=%v", prod.Environment(), prod.IsProduction())
	}
	if prod.GatewayBaseURL() != gatewayURLProduction {
		t.Fatalf("production gateway = %q", prod.GatewayBaseURL())
	}
}

func mustPrivatePEM(t *testing.T) string {
	t.Helper()
	initTestMaterial(t)
	return testPrivatePEM
}

func mustPublicPEM(t *testing.T) string {
	t.Helper()
	initTestMaterial(t)
	return testPublicPEM
}

func mustSelfSignedCertPEM(t *testing.T) string {
	t.Helper()
	initTestMaterial(t)
	return testCertPEM
}
