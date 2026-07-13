package alipay

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/f2xme/gox/oauth2"
)

func TestAuthCodeURL(t *testing.T) {
	provider := New(WithClientID("app-id"), WithRedirectURL("https://example.com/callback"))
	got := provider.AuthCodeURL("state")
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "openauth.alipay.com" {
		t.Fatalf("unexpected host: %s", u.Host)
	}
	if got := u.Query().Get("app_id"); got != "app-id" {
		t.Fatalf("unexpected app_id: %s", got)
	}
	if got := u.Query().Get("scope"); got != "auth_user" {
		t.Fatalf("unexpected scope: %s", got)
	}
}

func TestDefaultHTTPClientHasTimeout(t *testing.T) {
	if got := defaultOptions().HTTPClient.Timeout; got != defaultHTTPTimeout {
		t.Fatalf("default HTTP timeout = %v, want %v", got, defaultHTTPTimeout)
	}
}

func TestExchangeAndUserInfo(t *testing.T) {
	keys := testKeyPair(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("app_id") != "app-id" || r.Form.Get("sign") == "" {
			t.Fatalf("unexpected signed request: %v", r.Form)
		}

		switch r.Form.Get("method") {
		case "alipay.system.oauth.token":
			if r.Form.Get("code") != "code" {
				t.Fatalf("unexpected code: %s", r.Form.Get("code"))
			}
			writeSignedResponse(t, w, keys.key, "alipay_system_oauth_token_response", map[string]any{
				"code":          "10000",
				"msg":           "Success",
				"access_token":  "access",
				"expires_in":    "7200",
				"refresh_token": "refresh",
				"user_id":       "user-id",
				"open_id":       "open-id",
			})
		case "alipay.user.info.share":
			if r.Form.Get("auth_token") != "access" {
				t.Fatalf("unexpected auth token: %s", r.Form.Get("auth_token"))
			}
			writeSignedResponse(t, w, keys.key, "alipay_user_info_share_response", map[string]any{
				"code":      "10000",
				"msg":       "Success",
				"user_id":   "user-id",
				"open_id":   "open-id",
				"nick_name": "支付宝用户",
				"avatar":    "https://example.com/avatar.jpg",
				"province":  "浙江",
				"city":      "杭州",
				"gender":    "m",
			})
		default:
			t.Fatalf("unexpected method: %s", r.Form.Get("method"))
		}
	}))
	defer server.Close()

	provider := New(
		WithClientID("app-id"),
		WithPrivateKey(keys.privatePEM),
		WithAlipayPublicKey(keys.publicPEM),
		WithRedirectURL("https://example.com/callback"),
		WithEndpoints("", server.URL),
	)
	token, err := provider.Exchange(context.Background(), "code")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "access" || token.OpenID != "open-id" || token.UnionID != "user-id" {
		t.Fatalf("unexpected token: %#v", token)
	}

	user, err := provider.UserInfo(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if user.Provider != providerName || user.ID != "open-id" || user.Nickname != "支付宝用户" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

func TestExchangeProviderError(t *testing.T) {
	keys := testKeyPair(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeSignedResponse(t, w, keys.key, "alipay_system_oauth_token_response", map[string]any{
			"code":     "40002",
			"msg":      "Invalid Arguments",
			"sub_code": "isv.code-invalid",
			"sub_msg":  "invalid code",
		})
	}))
	defer server.Close()

	provider := New(WithClientID("app-id"), WithPrivateKey(keys.privatePEM), WithAlipayPublicKey(keys.publicPEM), WithEndpoints("", server.URL))
	_, err := provider.Exchange(context.Background(), "bad")
	if !errors.Is(err, oauth2.ErrProviderResponse) {
		t.Fatalf("expected provider response error, got %v", err)
	}
}

func TestExchangeRejectsInvalidResponseSignature(t *testing.T) {
	keys := testKeyPair(t)
	otherKeys := testKeyPair(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeSignedResponse(t, w, otherKeys.key, "alipay_system_oauth_token_response", map[string]any{
			"code":         "10000",
			"access_token": "forged",
		})
	}))
	defer server.Close()

	provider := New(WithClientID("app-id"), WithPrivateKey(keys.privatePEM), WithAlipayPublicKey(keys.publicPEM), WithEndpoints("", server.URL))
	if _, err := provider.Exchange(context.Background(), "code"); err == nil {
		t.Fatal("Exchange() error = nil, want signature error")
	}
}

func TestSignValues(t *testing.T) {
	keys := testKeyPair(t)
	values := url.Values{}
	values.Set("app_id", "app-id")
	values.Set("method", "alipay.system.oauth.token")
	values.Set("sign_type", "RSA2")

	sign, err := signValues(values, keys.privatePEM, "RSA2")
	if err != nil {
		t.Fatal(err)
	}
	if sign == "" {
		t.Fatal("expected non-empty sign")
	}
}

type testKeys struct {
	key        *rsa.PrivateKey
	privatePEM string
	publicPEM  string
}

func testKeyPair(t *testing.T) testKeys {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	publicDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return testKeys{
		key:        key,
		privatePEM: string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})),
		publicPEM:  string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})),
	}
}

func writeSignedResponse(t *testing.T, w http.ResponseWriter, key *rsa.PrivateKey, responseKey string, payload any) {
	t.Helper()
	content, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(content)
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(w).Encode(map[string]any{
		responseKey: json.RawMessage(content),
		"sign":      base64.StdEncoding.EncodeToString(signature),
	}); err != nil {
		t.Fatal(err)
	}
}
