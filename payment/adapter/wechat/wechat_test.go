package wechat

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
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

	ensureCalls int
	ensureErr   error
	// ensureAdds 在 ensurePublicKey 时合并进 keys（模拟补拉成功）。
	ensureAdds map[string]*rsa.PublicKey
	mu         sync.Mutex
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
func (f *fakeGateway) publicKeys() map[string]*rsa.PublicKey {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make(map[string]*rsa.PublicKey, len(f.keys))
	for k, v := range f.keys {
		out[k] = v
	}
	return out
}
func (f *fakeGateway) ensurePublicKey(serial string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ensureCalls++
	if f.ensureErr != nil {
		return f.ensureErr
	}
	if f.ensureAdds != nil {
		if f.keys == nil {
			f.keys = make(map[string]*rsa.PublicKey)
		}
		for k, v := range f.ensureAdds {
			f.keys[k] = v
		}
	}
	_ = serial
	return nil
}

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

func boolPtr(v bool) *bool { return &v }

func baseValidConfig() Config {
	return Config{
		AppID:              "app",
		MchID:              "mch",
		MerchantSerialNo:   "serial",
		MerchantPrivateKey: "private-key",
		APIV3Key:           "12345678901234567890123456789012",
	}
}

func TestValidateConfigVerifyModes(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		want    VerifyMode
	}{
		{
			name: "infer public key mode ok",
			cfg: func() Config {
				c := baseValidConfig()
				c.WechatPayPublicKey = "pub"
				c.WechatPayPublicKeyID = "PUB_KEY_ID_1"
				return c
			}(),
			want: VerifyModePublicKey,
		},
		{
			name: "explicit public key mode ok",
			cfg: func() Config {
				c := baseValidConfig()
				c.VerifyMode = VerifyModePublicKey
				c.WechatPayPublicKey = "pub"
				c.WechatPayPublicKeyID = "PUB_KEY_ID_1"
				return c
			}(),
			want: VerifyModePublicKey,
		},
		{
			name: "infer platform cert static mode ok",
			cfg: func() Config {
				c := baseValidConfig()
				c.PlatformCert = "cert"
				c.PlatformCertSerialNo = "plat-serial"
				return c
			}(),
			want: VerifyModePlatformCertStatic,
		},
		{
			name: "explicit platform cert static mode ok",
			cfg: func() Config {
				c := baseValidConfig()
				c.VerifyMode = VerifyModePlatformCertStatic
				c.PlatformCert = "cert"
				c.PlatformCertSerialNo = "plat-serial"
				return c
			}(),
			want: VerifyModePlatformCertStatic,
		},
		{
			name: "empty materials fail-closed without explicit auto",
			cfg:  baseValidConfig(),
			wantErr: true,
		},
		{
			name: "explicit platform cert auto mode ok",
			cfg: func() Config {
				c := baseValidConfig()
				c.VerifyMode = VerifyModePlatformCertAuto
				return c
			}(),
			want: VerifyModePlatformCertAuto,
		},
		{
			name: "platform cert auto mode with refresh off ok",
			cfg: func() Config {
				c := baseValidConfig()
				c.VerifyMode = VerifyModePlatformCertAuto
				c.PlatformCertAutoRefresh = boolPtr(false)
				return c
			}(),
			want: VerifyModePlatformCertAuto,
		},
		{
			name: "partial public key rejects",
			cfg: func() Config {
				c := baseValidConfig()
				c.WechatPayPublicKey = "pub"
				return c
			}(),
			wantErr: true,
		},
		{
			name: "partial public key id rejects",
			cfg: func() Config {
				c := baseValidConfig()
				c.WechatPayPublicKeyID = "PUB_KEY_ID_1"
				return c
			}(),
			wantErr: true,
		},
		{
			name: "partial platform cert rejects",
			cfg: func() Config {
				c := baseValidConfig()
				c.PlatformCert = "cert"
				return c
			}(),
			wantErr: true,
		},
		{
			name: "partial platform cert serial rejects",
			cfg: func() Config {
				c := baseValidConfig()
				c.PlatformCertSerialNo = "plat-serial"
				return c
			}(),
			wantErr: true,
		},
		{
			name: "public key and platform cert mutually exclusive",
			cfg: func() Config {
				c := baseValidConfig()
				c.WechatPayPublicKey = "pub"
				c.WechatPayPublicKeyID = "PUB_KEY_ID_1"
				c.PlatformCert = "cert"
				c.PlatformCertSerialNo = "plat-serial"
				return c
			}(),
			wantErr: true,
		},
		{
			name: "explicit public key without materials rejects",
			cfg: func() Config {
				c := baseValidConfig()
				c.VerifyMode = VerifyModePublicKey
				return c
			}(),
			wantErr: true,
		},
		{
			name: "explicit static without materials rejects",
			cfg: func() Config {
				c := baseValidConfig()
				c.VerifyMode = VerifyModePlatformCertStatic
				return c
			}(),
			wantErr: true,
		},
		{
			name: "auto mode rejects public key materials",
			cfg: func() Config {
				c := baseValidConfig()
				c.VerifyMode = VerifyModePlatformCertAuto
				c.WechatPayPublicKey = "pub"
				c.WechatPayPublicKeyID = "PUB_KEY_ID_1"
				return c
			}(),
			wantErr: true,
		},
		{
			name: "auto mode rejects static cert materials",
			cfg: func() Config {
				c := baseValidConfig()
				c.VerifyMode = VerifyModePlatformCertAuto
				c.PlatformCert = "cert"
				c.PlatformCertSerialNo = "plat-serial"
				return c
			}(),
			wantErr: true,
		},
		{
			name: "public key mode rejects PlatformCertAutoRefresh",
			cfg: func() Config {
				c := baseValidConfig()
				c.VerifyMode = VerifyModePublicKey
				c.WechatPayPublicKey = "pub"
				c.WechatPayPublicKeyID = "PUB_KEY_ID_1"
				c.PlatformCertAutoRefresh = boolPtr(true)
				return c
			}(),
			wantErr: true,
		},
		{
			name: "static mode rejects PlatformCertAutoRefresh",
			cfg: func() Config {
				c := baseValidConfig()
				c.VerifyMode = VerifyModePlatformCertStatic
				c.PlatformCert = "cert"
				c.PlatformCertSerialNo = "plat-serial"
				c.PlatformCertAutoRefresh = boolPtr(false)
				return c
			}(),
			wantErr: true,
		},
		{
			name: "unknown verify mode rejects",
			cfg: func() Config {
				c := baseValidConfig()
				c.VerifyMode = "unknown"
				return c
			}(),
			wantErr: true,
		},
		{
			name:    "missing app id rejects",
			cfg:     Config{MchID: "mch", MerchantSerialNo: "s", MerchantPrivateKey: "k", APIV3Key: "12345678901234567890123456789012"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("validateConfig() error = nil, want error")
				}
				if !errors.Is(err, payment.ErrInvalidConfig) {
					t.Fatalf("validateConfig() error = %v, want ErrInvalidConfig", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateConfig() error = %v", err)
			}
			if got := tt.cfg.ResolveVerifyMode(); got != tt.want {
				t.Fatalf("ResolveVerifyMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveVerifyModeAndRefresh(t *testing.T) {
	if got := (Config{WechatPayPublicKey: "k", WechatPayPublicKeyID: "id"}).ResolveVerifyMode(); got != VerifyModePublicKey {
		t.Fatalf("public key resolve = %q", got)
	}
	if got := (Config{PlatformCert: "c", PlatformCertSerialNo: "s"}).ResolveVerifyMode(); got != VerifyModePlatformCertStatic {
		t.Fatalf("static resolve = %q", got)
	}
	if got := (Config{}).ResolveVerifyMode(); got != "" {
		t.Fatalf("empty materials resolve = %q, want empty", got)
	}
	auto := Config{VerifyMode: VerifyModePlatformCertAuto}
	if got := auto.ResolveVerifyMode(); got != VerifyModePlatformCertAuto {
		t.Fatalf("auto resolve = %q", got)
	}
	if !auto.platformCertShouldAutoRefresh() {
		t.Fatal("default auto refresh should be true")
	}
	auto.PlatformCertAutoRefresh = boolPtr(false)
	if auto.platformCertShouldAutoRefresh() {
		t.Fatal("PlatformCertAutoRefresh=false should disable refresh")
	}
}

func TestNormalizeAndValidatePlatformCertSerial(t *testing.T) {
	if got := normalizeCertSerial(" ab:cd:ef "); got != "ABCDEF" {
		t.Fatalf("normalizeCertSerial = %q", got)
	}
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	serialInt := big.NewInt(0x5157F09E)
	certPEM := mustSelfSignedCertPEM(t, priv, serialInt)
	want := normalizeCertSerial(serialInt.Text(16))

	fromPEM, err := platformCertSerialFromPEM(certPEM)
	if err != nil || fromPEM != want {
		t.Fatalf("platformCertSerialFromPEM = %q, %v want %q", fromPEM, err, want)
	}
	if err := validatePlatformCertSerial(certPEM, want); err != nil {
		t.Fatalf("validate match error = %v", err)
	}
	if err := validatePlatformCertSerial(certPEM, "AA:BB:CC"); !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("validate mismatch error = %v", err)
	}
	// PUBLIC KEY PEM：无法提取序列号，跳过交叉校验。
	pubPEM := mustPublicPEM(t, &priv.PublicKey)
	if err := validatePlatformCertSerial(pubPEM, "ANY"); err != nil {
		t.Fatalf("public key pem serial check error = %v", err)
	}
}

func TestClassifyPlatformCertAutoError(t *testing.T) {
	if err := classifyPlatformCertAutoError(errors.New("connection timeout")); !errors.Is(err, payment.ErrGateway) {
		t.Fatalf("timeout => %v", err)
	}
	if err := classifyPlatformCertAutoError(errors.New("aes decrypt failed")); !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("decrypt => %v", err)
	}
	if err := classifyPlatformCertAutoError(errors.New("HTTP 403 forbidden")); !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("403 => %v", err)
	}
}

func TestNewPublicKeyAndStaticModes(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privPEM := mustPrivatePEM(t, priv)
	pubPEM := mustPublicPEM(t, &priv.PublicKey)
	serialInt := big.NewInt(0xABCDEF01)
	certPEM := mustSelfSignedCertPEM(t, priv, serialInt)
	certSerial := normalizeCertSerial(serialInt.Text(16))

	t.Run("public key mode", func(t *testing.T) {
		client, err := New(Config{
			AppID:                "app",
			MchID:                "mch",
			MerchantSerialNo:     "merchant-serial",
			MerchantPrivateKey:   privPEM,
			APIV3Key:             "12345678901234567890123456789012",
			VerifyMode:           VerifyModePublicKey,
			WechatPayPublicKey:   pubPEM,
			WechatPayPublicKeyID: "PUB_KEY_ID_TEST",
		})
		if err != nil {
			t.Fatalf("New public key = %v", err)
		}
		if client == nil {
			t.Fatal("client is nil")
		}
	})

	t.Run("static cert mode", func(t *testing.T) {
		client, err := New(Config{
			AppID:                "app",
			MchID:                "mch",
			MerchantSerialNo:     "merchant-serial",
			MerchantPrivateKey:   privPEM,
			APIV3Key:             "12345678901234567890123456789012",
			VerifyMode:           VerifyModePlatformCertStatic,
			PlatformCert:         certPEM,
			PlatformCertSerialNo: certSerial,
		})
		if err != nil {
			t.Fatalf("New static cert = %v", err)
		}
		if client == nil {
			t.Fatal("client is nil")
		}
	})

	t.Run("static cert serial mismatch", func(t *testing.T) {
		_, err := New(Config{
			AppID:                "app",
			MchID:                "mch",
			MerchantSerialNo:     "merchant-serial",
			MerchantPrivateKey:   privPEM,
			APIV3Key:             "12345678901234567890123456789012",
			VerifyMode:           VerifyModePlatformCertStatic,
			PlatformCert:         certPEM,
			PlatformCertSerialNo: "DEADBEEF",
		})
		if !errors.Is(err, payment.ErrInvalidConfig) {
			t.Fatalf("serial mismatch error = %v", err)
		}
	})

	t.Run("auto mode network failure is gateway", func(t *testing.T) {
		transport := roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("connection refused")
		})
		_, err := New(Config{
			AppID:                   "app",
			MchID:                   "mch",
			MerchantSerialNo:        "merchant-serial",
			MerchantPrivateKey:      privPEM,
			APIV3Key:                "12345678901234567890123456789012",
			VerifyMode:              VerifyModePlatformCertAuto,
			PlatformCertAutoRefresh: boolPtr(false),
		}, WithHTTPTransport(transport))
		if !errors.Is(err, payment.ErrGateway) {
			t.Fatalf("auto network error = %v, want ErrGateway", err)
		}
	})

	t.Run("empty materials without auto mode", func(t *testing.T) {
		_, err := New(Config{
			AppID:              "app",
			MchID:              "mch",
			MerchantSerialNo:   "merchant-serial",
			MerchantPrivateKey: privPEM,
			APIV3Key:           "12345678901234567890123456789012",
		})
		if !errors.Is(err, payment.ErrInvalidConfig) {
			t.Fatalf("empty materials error = %v", err)
		}
	})
}

func TestNotificationsWithPlatformCertSerial(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	// 平台证书模式下 Wechatpay-Serial 为证书序列号（非 PUB_KEY_ID_ 前缀）。
	serial := "5157F09EFDC096DE15EBE81A47057A7232F1B8E1"
	w := newWithGateway(testConfig(), &fakeGateway{keys: map[string]*rsa.PublicKey{serial: &privateKey.PublicKey}})
	payPlain := `{"appid":"app","mchid":"mch","out_trade_no":"order-1","transaction_id":"tx","trade_state":"SUCCESS","success_time":"2026-07-12T12:00:00+08:00","amount":{"total":123}}`
	req := signedNotifyRequestWithSerial(t, privateKey, testConfig().APIV3Key, payPlain, serial)
	pay, err := w.ParsePaymentNotification(context.Background(), req)
	if err != nil || pay.Amount != 123 || pay.Status != payment.PaymentStatusSuccess {
		t.Fatalf("platform cert payment notification = %#v, %v", pay, err)
	}
}

func TestNotificationSerialMissRefresh(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	serial := "5157F09EFDC096DE15EBE81A47057A7232F1B8E1"
	payPlain := `{"appid":"app","mchid":"mch","out_trade_no":"order-1","transaction_id":"tx","trade_state":"SUCCESS","success_time":"2026-07-12T12:00:00+08:00","amount":{"total":123}}`

	t.Run("refresh then success", func(t *testing.T) {
		gw := &fakeGateway{
			keys:       map[string]*rsa.PublicKey{}, // 初始无 serial
			ensureAdds: map[string]*rsa.PublicKey{serial: &privateKey.PublicKey},
		}
		w := newWithGateway(testConfig(), gw)
		req := signedNotifyRequestWithSerial(t, privateKey, testConfig().APIV3Key, payPlain, serial)
		pay, err := w.ParsePaymentNotification(context.Background(), req)
		if err != nil || pay.Amount != 123 {
			t.Fatalf("after refresh = %#v, %v", pay, err)
		}
		if gw.ensureCalls != 1 {
			t.Fatalf("ensureCalls = %d, want 1", gw.ensureCalls)
		}
	})

	t.Run("known serial bad signature does not refresh", func(t *testing.T) {
		other, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}
		gw := &fakeGateway{keys: map[string]*rsa.PublicKey{serial: &other.PublicKey}}
		w := newWithGateway(testConfig(), gw)
		req := signedNotifyRequestWithSerial(t, privateKey, testConfig().APIV3Key, payPlain, serial)
		_, err = w.ParsePaymentNotification(context.Background(), req)
		if !errors.Is(err, payment.ErrInvalidSignature) {
			t.Fatalf("error = %v, want ErrInvalidSignature", err)
		}
		if gw.ensureCalls != 0 {
			t.Fatalf("ensureCalls = %d, want 0", gw.ensureCalls)
		}
	})

	t.Run("refresh failure preserves gateway error chain", func(t *testing.T) {
		gw := &fakeGateway{
			keys:      map[string]*rsa.PublicKey{},
			ensureErr: fmt.Errorf("%w: connection refused", payment.ErrGateway),
		}
		w := newWithGateway(testConfig(), gw)
		req := signedNotifyRequestWithSerial(t, privateKey, testConfig().APIV3Key, payPlain, serial)
		_, err := w.ParsePaymentNotification(context.Background(), req)
		if !errors.Is(err, payment.ErrInvalidSignature) {
			t.Fatalf("error = %v, want ErrInvalidSignature", err)
		}
		if !errors.Is(err, payment.ErrGateway) {
			t.Fatalf("error = %v, want ErrGateway in chain", err)
		}
		if gw.ensureCalls != 1 {
			t.Fatalf("ensureCalls = %d, want 1", gw.ensureCalls)
		}
	})
}

func TestSkipPlatformCertRefresh(t *testing.T) {
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	serial := "ABC"
	cooldown := time.Minute

	if skipPlatformCertRefresh(now, serial, time.Time{}, nil, cooldown) {
		t.Fatal("empty state should not skip")
	}
	if !skipPlatformCertRefresh(now, serial, now.Add(-30*time.Second), nil, cooldown) {
		t.Fatal("within cooldown should skip")
	}
	if skipPlatformCertRefresh(now, serial, now.Add(-2*time.Minute), nil, cooldown) {
		t.Fatal("past cooldown should not skip")
	}
	miss := map[string]time.Time{serial: now.Add(time.Minute)}
	if !skipPlatformCertRefresh(now, serial, time.Time{}, miss, cooldown) {
		t.Fatal("active miss cache should skip")
	}
	miss[serial] = now.Add(-time.Second)
	if skipPlatformCertRefresh(now, serial, time.Time{}, miss, cooldown) {
		t.Fatal("expired miss cache should not skip")
	}
}

func TestPruneMissCache(t *testing.T) {
	now := time.Now()
	m := map[string]time.Time{
		"live":  now.Add(time.Minute),
		"stale": now.Add(-time.Second),
	}
	pruneMissCache(m, now)
	if _, ok := m["live"]; !ok {
		t.Fatal("live entry pruned")
	}
	if _, ok := m["stale"]; ok {
		t.Fatal("stale entry not pruned")
	}
}

func signedNotifyRequestWithSerial(t *testing.T, key *rsa.PrivateKey, apiKey, plain, serial string) *http.Request {
	t.Helper()
	req := signedNotifyRequest(t, key, apiKey, plain)
	req.Header.Set("Wechatpay-Serial", serial)
	return req
}

func mustPrivatePEM(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()
	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
}

func mustPublicPEM(t *testing.T, key *rsa.PublicKey) string {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

func mustSelfSignedCertPEM(t *testing.T, key *rsa.PrivateKey, serial *big.Int) string {
	t.Helper()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "wechat-platform-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}
