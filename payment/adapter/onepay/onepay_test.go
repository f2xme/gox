package onepay

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/f2xme/gox/payment"
)

type fakeResolver struct {
	provider     payment.Provider
	intent       string
	payerOpenID  string
	err          error
	resolveCalls int
	creates      int
	checkouts    map[payment.Provider]*Checkout
	wechatPayer  string
}

func (f *fakeResolver) ResolveOrCreate(_ context.Context, intent string, provider payment.Provider, payerOpenID string) (*Checkout, error) {
	f.intent, f.provider, f.payerOpenID = intent, provider, payerOpenID
	f.resolveCalls++
	if f.err != nil {
		return nil, f.err
	}
	if f.checkouts == nil {
		f.checkouts = make(map[payment.Provider]*Checkout)
	}
	if checkout := f.checkouts[provider]; checkout != nil {
		if provider == payment.ProviderWechat && f.wechatPayer != payerOpenID {
			return nil, errors.New("wechat payer conflict")
		}
		return checkout, nil
	}
	f.creates++
	checkout := &Checkout{Provider: provider, OrderID: string(provider) + "-1", ExpiresAt: time.Now().Add(time.Hour)}
	if provider == payment.ProviderAlipay {
		checkout.WAP = &payment.WAPResult{URL: "https://openapi.alipay.com/gateway.do?x=1"}
	} else {
		f.wechatPayer = payerOpenID
		checkout.JSAPI = &payment.JSAPIResult{AppID: "app", Timestamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s"}
	}
	f.checkouts[provider] = checkout
	return checkout, nil
}

type fakeWechat struct {
	code, openID string
	err          error
}

func (f *fakeWechat) OAuthURL(redirectURL, state string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	u, _ := url.Parse("https://open.weixin.qq.com/connect/oauth2/authorize")
	q := u.Query()
	q.Set("redirect_uri", redirectURL)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
func (f *fakeWechat) ExchangeOAuthCode(_ context.Context, code string) (string, error) {
	f.code = code
	return f.openID, f.err
}
func newTestService(t *testing.T) (*Service, *fakeResolver, *fakeWechat) {
	t.Helper()
	r := &fakeResolver{}
	w := &fakeWechat{openID: "openid"}
	s, err := New(Config{BaseURL: "https://pay.example", TokenKey: bytes.Repeat([]byte{7}, 32), Resolver: r, Wechat: w})
	if err != nil {
		t.Fatal(err)
	}
	return s, r, w
}

func TestCreateCodePNGAndToken(t *testing.T) {
	s, _, _ := newTestService(t)
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	first, err := s.CreateCode(context.Background(), "intent-1")
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.CreateCode(context.Background(), "intent-1")
	if err != nil {
		t.Fatal(err)
	}
	if first.URL == second.URL {
		t.Fatal("token nonce must be random")
	}
	img, err := png.Decode(bytes.NewReader(first.PNG))
	if err != nil || img.Bounds().Dx() != 256 || img.Bounds().Dy() != 256 {
		t.Fatalf("PNG = %v, %v", img.Bounds(), err)
	}
	token := strings.TrimPrefix(first.URL, "https://pay.example/pay/")
	payload, err := s.decryptToken(token)
	if err != nil || payload.IntentID != "intent-1" {
		t.Fatalf("decrypt = %#v, %v", payload, err)
	}
	replacement := "A"
	if token[0] == 'A' {
		replacement = "B"
	}
	bad := replacement + token[1:]
	if _, err := s.decryptToken(bad); !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("tamper error = %v", err)
	}
	s.now = func() time.Time { return first.ExpiresAt.Add(time.Second) }
	if _, err := s.decryptToken(token); !errors.Is(err, payment.ErrExpired) {
		t.Fatalf("expired error = %v", err)
	}
}

func TestOAuthStateConstraintsAndRoundTrip(t *testing.T) {
	s, _, _ := newTestService(t)
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	state, payload, err := s.createState("token", now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(state) != 79 || !regexp.MustCompile(`^[A-Z2-7]+$`).MatchString(state) {
		t.Fatalf("state length/charset = %d %q", len(state), state)
	}
	if err := s.verifyState(state, "token", payload.Nonce); err != nil {
		t.Fatalf("verifyState() error = %v", err)
	}
	replacement := "A"
	if state[0] == 'A' {
		replacement = "B"
	}
	if err := s.verifyState(replacement+state[1:], "token", payload.Nonce); !errors.Is(err, payment.ErrInvalidOAuthState) {
		t.Fatalf("tamper error = %v", err)
	}
	if err := s.verifyState(state, "other-token", payload.Nonce); !errors.Is(err, payment.ErrInvalidOAuthState) {
		t.Fatalf("wrong token error = %v", err)
	}
	if err := s.verifyState(state, "token", stateEncoding.EncodeToString(make([]byte, stateNonceSize))); !errors.Is(err, payment.ErrInvalidOAuthState) {
		t.Fatalf("wrong cookie error = %v", err)
	}
	s.now = func() time.Time { return now.Add(2 * time.Minute) }
	if err := s.verifyState(state, "token", payload.Nonce); !errors.Is(err, payment.ErrInvalidOAuthState) {
		t.Fatalf("expired state error = %v", err)
	}
}

func TestCreateCodeRejectsQRTooLargeForCanvas(t *testing.T) {
	s, _, _ := newTestService(t)
	small, err := New(s.config, WithQRSize(128))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := small.CreateCode(context.Background(), strings.Repeat("intent", 250)); err == nil {
		t.Fatal("expected QR capacity error")
	}
}

func TestCreateCodeRejectsExpirationInCurrentUnixSecond(t *testing.T) {
	s, _, _ := newTestService(t)
	now := time.Date(2026, 7, 13, 12, 0, 0, 100_000_000, time.UTC)
	s.now = func() time.Time { return now }

	if _, err := s.CreateCode(context.Background(), "intent", WithExpiresAt(now.Add(500*time.Millisecond))); !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("CreateCode() error = %v, want ErrInvalidRequest", err)
	}
}

func TestHandlerAlipayAndUnknown(t *testing.T) {
	s, resolver, _ := newTestService(t)
	code, _ := s.CreateCode(context.Background(), "intent-1")
	path := strings.TrimPrefix(code.URL, "https://pay.example")
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("User-Agent", "AlipayClient/10")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != resolver.checkouts[payment.ProviderAlipay].WAP.URL {
		t.Fatalf("response = %d %q", rec.Code, rec.Header().Get("Location"))
	}
	if resolver.provider != payment.ProviderAlipay || resolver.intent != "intent-1" {
		t.Fatalf("resolve = %q %q", resolver.intent, resolver.provider)
	}
	req = httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("User-Agent", "Safari")
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "微信或支付宝") {
		t.Fatalf("unknown response = %d %s", rec.Code, rec.Body.String())
	}
	resolver.checkouts[payment.ProviderAlipay].WAP.URL = "https://evil.example/steal"
	req = httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("User-Agent", "AlipayClient")
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("unsafe redirect status = %d", rec.Code)
	}
}

func TestHandlerWechatOAuthAndJSAPI(t *testing.T) {
	s, resolver, wechat := newTestService(t)
	code, _ := s.CreateCode(context.Background(), "intent-wx")
	path := strings.TrimPrefix(code.URL, "https://pay.example")
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("User-Agent", "MicroMessenger")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("OAuth status = %d", rec.Code)
	}
	location, _ := url.Parse(rec.Header().Get("Location"))
	state := location.Query().Get("state")
	cookies := rec.Result().Cookies()
	if state == "" || len(cookies) != 1 || !cookies[0].Secure || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("state/cookie = %q %#v", state, cookies)
	}
	callback := path + "?code=oauth-code&state=" + url.QueryEscape(state)
	req = httptest.NewRequest(http.MethodGet, callback, nil)
	req.Header.Set("User-Agent", "MicroMessenger")
	req.AddCookie(cookies[0])
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "WeixinJSBridge.invoke") || !strings.Contains(body, "prepay_id=p") {
		t.Fatalf("JSAPI response = %d %s", rec.Code, body)
	}
	if !strings.Contains(body, DefaultWechatLoadingText) {
		t.Fatalf("default loading text missing: %s", body)
	}
	if !strings.Contains(rec.Header().Get("Content-Security-Policy"), "default-src 'none'") || resolver.provider != payment.ProviderWechat || wechat.code != "oauth-code" || resolver.payerOpenID != "openid" {
		t.Fatalf("callback state = %q %q", rec.Header().Get("Content-Security-Policy"), wechat.code)
	}
}

func TestHandlerWechatCustomPageTexts(t *testing.T) {
	r := &fakeResolver{}
	w := &fakeWechat{openID: "openid"}
	s, err := New(Config{
		BaseURL:  "https://pay.example",
		TokenKey: bytes.Repeat([]byte{7}, 32),
		Resolver: r,
		Wechat:   w,
		WechatPage: WechatPage{
			Title:       "  收银台  ",
			LoadingText: "  正在支付…  ",
			SuccessText: "已付成功",
			FailText:    "付失败了",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	code, _ := s.CreateCode(context.Background(), "intent-wx-custom")
	path := strings.TrimPrefix(code.URL, "https://pay.example")
	rec := completeWechatFlow(t, s, path)
	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, body)
	}
	for _, want := range []string{"收银台", "正在支付…", `"已付成功"`, `"付失败了"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in %s", want, body)
		}
	}
	if strings.Contains(body, DefaultWechatLoadingText) || strings.Contains(body, DefaultWechatTitle) {
		t.Fatalf("defaults should be overridden: %s", body)
	}
	if csp := rec.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'none'") || !strings.Contains(csp, "nonce-") {
		t.Fatalf("csp = %q", csp)
	}
}

func TestHandlerWechatCustomTemplate(t *testing.T) {
	tpl := template.Must(template.New("custom").Parse(`<!doctype html><body data-custom="1"><p>{{.LoadingText}}</p><script nonce="{{.Nonce}}">const pay={{.Params}};const ok={{.SuccessText}};</script></body>`))
	r := &fakeResolver{}
	w := &fakeWechat{openID: "openid"}
	s, err := New(Config{
		BaseURL:    "https://pay.example",
		TokenKey:   bytes.Repeat([]byte{7}, 32),
		Resolver:   r,
		Wechat:     w,
		WechatPage: WechatPage{LoadingText: "自定义加载", Template: tpl},
	})
	if err != nil {
		t.Fatal(err)
	}
	code, _ := s.CreateCode(context.Background(), "intent-wx-tpl")
	path := strings.TrimPrefix(code.URL, "https://pay.example")
	rec := completeWechatFlow(t, s, path)
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, `data-custom="1"`) || !strings.Contains(body, "自定义加载") || !strings.Contains(body, "prepay_id=p") {
		t.Fatalf("custom template response = %d %s", rec.Code, body)
	}
}

func TestHandlerWechatBadTemplateReturns502(t *testing.T) {
	// missing template name → Execute error; must not return partial 200
	tpl := template.Must(template.New("bad").Parse(`{{template "missing"}}`))
	r := &fakeResolver{}
	w := &fakeWechat{openID: "openid"}
	s, err := New(Config{
		BaseURL:    "https://pay.example",
		TokenKey:   bytes.Repeat([]byte{7}, 32),
		Resolver:   r,
		Wechat:     w,
		WechatPage: WechatPage{Template: tpl},
	})
	if err != nil {
		t.Fatal(err)
	}
	code, _ := s.CreateCode(context.Background(), "intent-wx-bad-tpl")
	path := strings.TrimPrefix(code.URL, "https://pay.example")
	rec := completeWechatFlow(t, s, path)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "WeixinJSBridge") {
		t.Fatalf("should not leak bridge body on template error: %s", rec.Body.String())
	}
}

func TestHandlerWechatPageEscapesXSS(t *testing.T) {
	r := &fakeResolver{}
	w := &fakeWechat{openID: "openid"}
	s, err := New(Config{
		BaseURL:  "https://pay.example",
		TokenKey: bytes.Repeat([]byte{7}, 32),
		Resolver: r,
		Wechat:   w,
		WechatPage: WechatPage{
			Title:       `<img src=x onerror=alert(1)>`,
			LoadingText: `<script>alert(1)</script>`,
			SuccessText: `</script><script>alert(2)</script>`,
			FailText:    `";alert(3)//`,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	code, _ := s.CreateCode(context.Background(), "intent-wx-xss")
	path := strings.TrimPrefix(code.URL, "https://pay.example")
	rec := completeWechatFlow(t, s, path)
	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, body)
	}
	// HTML context: raw tags must not appear
	for _, raw := range []string{
		`<script>alert(1)</script>`,
		`<img src=x onerror=alert(1)>`,
		`</script><script>alert(2)</script>`,
	} {
		if strings.Contains(body, raw) {
			t.Fatalf("raw XSS payload in body: %q\n%s", raw, body)
		}
	}
	// HTML-escaped loading/title
	if !strings.Contains(body, "&lt;script&gt;") && !strings.Contains(body, "&lt;img") {
		t.Fatalf("expected HTML entity escape: %s", body)
	}
	// Success/Fail as JSON string literals (encoding/json escapes < and ")
	if !strings.Contains(body, `\u003c`) && !strings.Contains(body, `\u003C`) {
		t.Fatalf("expected JSON-escaped < in script strings: %s", body)
	}
	if !strings.Contains(body, `\u0022`) && !strings.Contains(body, `\"`) {
		t.Fatalf("expected JSON-escaped quote in fail text: %s", body)
	}
}

func TestResolveWechatPageDefaults(t *testing.T) {
	got := resolveWechatPage(WechatPage{})
	if got.Title != DefaultWechatTitle || got.LoadingText != DefaultWechatLoadingText || got.SuccessText != DefaultWechatSuccessText || got.FailText != DefaultWechatFailText {
		t.Fatalf("defaults = %+v", got)
	}
	got = resolveWechatPage(WechatPage{Title: "  t  ", LoadingText: "  x  ", SuccessText: " y ", FailText: "z "})
	if got.Title != "t" || got.LoadingText != "x" || got.SuccessText != "y" || got.FailText != "z" {
		t.Fatalf("trim = %+v", got)
	}
	got = resolveWechatPage(WechatPage{LoadingText: "   "})
	if got.LoadingText != DefaultWechatLoadingText {
		t.Fatalf("whitespace-only should default: %+v", got)
	}
}

func TestWechatBridgeCSP(t *testing.T) {
	csp := WechatBridgeCSP("abc")
	if !strings.Contains(csp, "nonce-abc") || !strings.Contains(csp, "default-src 'none'") {
		t.Fatalf("csp = %q", csp)
	}
}

func TestHandlerWechatRepeatedScanReusesCheckout(t *testing.T) {
	s, resolver, _ := newTestService(t)
	code, _ := s.CreateCode(context.Background(), "intent-wx")
	path := strings.TrimPrefix(code.URL, "https://pay.example")
	for range 2 {
		completeWechatFlow(t, s, path)
	}
	if resolver.resolveCalls != 2 || resolver.creates != 1 {
		t.Fatalf("resolve calls=%d creates=%d", resolver.resolveCalls, resolver.creates)
	}
}

func TestHandlerWechatRejectsDifferentPayer(t *testing.T) {
	s, _, wechat := newTestService(t)
	code, _ := s.CreateCode(context.Background(), "intent-wx")
	path := strings.TrimPrefix(code.URL, "https://pay.example")
	completeWechatFlow(t, s, path)
	wechat.openID = "another-openid"
	rec := completeWechatFlow(t, s, path)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("different payer status = %d", rec.Code)
	}
}

func completeWechatFlow(t *testing.T, s *Service, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("User-Agent", "MicroMessenger")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	location, err := url.Parse(rec.Header().Get("Location"))
	if err != nil || rec.Code != http.StatusSeeOther {
		t.Fatalf("OAuth response = %d %v", rec.Code, err)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %#v", cookies)
	}
	callback := path + "?code=oauth-code&state=" + url.QueryEscape(location.Query().Get("state"))
	req = httptest.NewRequest(http.MethodGet, callback, nil)
	req.Header.Set("User-Agent", "MicroMessenger")
	req.AddCookie(cookies[0])
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	return rec
}

func TestHandlerRejectsInvalidState(t *testing.T) {
	s, _, _ := newTestService(t)
	code, _ := s.CreateCode(context.Background(), "intent-wx")
	path := strings.TrimPrefix(code.URL, "https://pay.example")
	req := httptest.NewRequest(http.MethodGet, path+"?code=x&state=bad", nil)
	req.Header.Set("User-Agent", "MicroMessenger")
	req.AddCookie(&http.Cookie{Name: cookieName(strings.TrimPrefix(path, "/pay/")), Value: "bad"})
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "授权状态无效") {
		t.Fatalf("response = %d %s", rec.Code, rec.Body.String())
	}
}

func TestConfigValidation(t *testing.T) {
	r, w := &fakeResolver{}, &fakeWechat{}
	tests := []Config{
		{BaseURL: "http://pay.example", TokenKey: make([]byte, 32), Resolver: r, Wechat: w},
		{BaseURL: "https://pay.example/path", TokenKey: make([]byte, 32), Resolver: r, Wechat: w},
		{BaseURL: "https://pay.example", Path: "pay", TokenKey: make([]byte, 32), Resolver: r, Wechat: w},
		{BaseURL: "https://pay.example", TokenKey: make([]byte, 31), Resolver: r, Wechat: w},
	}
	for _, config := range tests {
		if _, err := New(config); !errors.Is(err, payment.ErrInvalidConfig) {
			t.Fatalf("New(%#v) error = %v", config, err)
		}
	}
}
