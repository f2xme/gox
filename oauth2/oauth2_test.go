package oauth2

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestApplyAuthCodeOptions(t *testing.T) {
	options := ApplyAuthCodeOptions(
		WithScopes("profile", "email"),
		WithAuthParam("display", "mobile"),
	)

	if len(options.Scopes) != 2 || options.Scopes[0] != "profile" || options.Scopes[1] != "email" {
		t.Fatalf("unexpected scopes: %#v", options.Scopes)
	}
	if got := options.Extra.Get("display"); got != "mobile" {
		t.Fatalf("unexpected extra param: %q", got)
	}

	options.Extra.Set("x", "1")
	next := ApplyAuthCodeOptions()
	if next.Extra == nil {
		t.Fatal("expected non-nil Extra")
	}
	if got := next.Extra.Encode(); got != "" {
		t.Fatalf("unexpected shared values: %s", got)
	}
}

func TestDefaultHTTPClientHasTimeout(t *testing.T) {
	if got := defaultOptions().HTTPClient.Timeout; got != defaultHTTPTimeout {
		t.Fatalf("default HTTP timeout = %v, want %v", got, defaultHTTPTimeout)
	}
}

func TestTokenExpired(t *testing.T) {
	if (&Token{}).Expired() {
		t.Fatal("zero expiry token should not be expired")
	}
	if !(&Token{Expiry: time.Now().Add(-time.Second)}).Expired() {
		t.Fatal("past expiry token should be expired")
	}
	if (&Token{Expiry: time.Now().Add(time.Hour)}).Expired() {
		t.Fatal("future expiry token should not be expired")
	}
}

func TestProviderErrorUnwrap(t *testing.T) {
	err := &ProviderError{Provider: "wechat", Code: "40029", Message: "invalid code", Raw: []byte(`{}`)}
	if !errors.Is(err, ErrProviderResponse) {
		t.Fatal("ProviderError should unwrap to ErrProviderResponse")
	}
	if err.Error() == "" {
		t.Fatal("expected non-empty message")
	}
}

func TestWithAuthParamInitializesExtra(t *testing.T) {
	options := AuthCodeOptions{}
	WithAuthParam("k", "v")(&options)
	if options.Extra == nil || options.Extra.Get("k") != "v" {
		t.Fatalf("unexpected extra values: %v", url.Values(options.Extra))
	}
}

func TestBuildAuthCodeURL(t *testing.T) {
	got := BuildAuthCodeURL(AuthCodeURLConfig{
		AuthURL:        "https://provider.example/auth",
		ClientID:       "client",
		ClientIDParam:  "appid",
		RedirectURL:    "https://example.com/callback",
		ScopeSeparator: ",",
		DefaultScopes:  []string{"profile", "email"},
		Fragment:       "redirect",
	}, "state", WithAuthParam("display", "mobile"))

	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if got := u.Query().Get("appid"); got != "client" {
		t.Fatalf("unexpected appid: %s", got)
	}
	if got := u.Query().Get("scope"); got != "profile,email" {
		t.Fatalf("unexpected scope: %s", got)
	}
	if got := u.Query().Get("display"); got != "mobile" {
		t.Fatalf("unexpected display: %s", got)
	}
	if u.Fragment != "redirect" {
		t.Fatalf("unexpected fragment: %s", u.Fragment)
	}
}

func TestBuildAuthCodeURLPreservesEndpointQueryAndFragment(t *testing.T) {
	got := BuildAuthCodeURL(AuthCodeURLConfig{
		AuthURL:     "https://provider.example/auth?tenant=fixed#provider-fragment",
		ClientID:    "client",
		RedirectURL: "https://example.com/callback",
	}, "state")

	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if got := u.Query().Get("tenant"); got != "fixed" {
		t.Fatalf("tenant = %q, want fixed", got)
	}
	if u.Fragment != "provider-fragment" {
		t.Fatalf("fragment = %q, want provider-fragment", u.Fragment)
	}
}

func TestNewTokenDoesNotExpireWhenExpiresInIsZero(t *testing.T) {
	token := NewToken(TokenInfo{AccessToken: "access"})
	if !token.Expiry.IsZero() {
		t.Fatalf("expected zero expiry, got %s", token.Expiry)
	}
	if token.Expired() {
		t.Fatal("token without expiry should not be expired")
	}
}

func TestAuthCodeClientExchangeAndRefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		switch r.Form.Get("grant_type") {
		case "authorization_code":
			if r.Form.Get("code") != "code" {
				t.Fatalf("unexpected code: %s", r.Form.Get("code"))
			}
		case "refresh_token":
			if r.Form.Get("refresh_token") != "refresh" {
				t.Fatalf("unexpected refresh token: %s", r.Form.Get("refresh_token"))
			}
		default:
			t.Fatalf("unexpected grant type: %s", r.Form.Get("grant_type"))
		}
		if r.Form.Get("client_id") != "client" || r.Form.Get("client_secret") != "secret" {
			t.Fatalf("unexpected client credentials: %v", r.Form)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "refresh-next",
			"scope":         "profile",
		})
	}))
	defer server.Close()

	client := NewClient(
		WithClientID("client"),
		WithClientSecret("secret"),
		WithRedirectURL("https://example.com/callback"),
		WithEndpoint(Endpoint{TokenURL: server.URL}),
	)

	token, err := client.Exchange(context.Background(), "code")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "access" || token.TokenType != "Bearer" || token.RefreshToken != "refresh-next" {
		t.Fatalf("unexpected token: %#v", token)
	}
	if token.Expiry.IsZero() {
		t.Fatal("expected expiry")
	}

	token, err = client.RefreshToken(context.Background(), "refresh")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "access" {
		t.Fatalf("unexpected refreshed token: %#v", token)
	}
}

func TestDecodeTokenProviderError(t *testing.T) {
	_, err := DecodeToken([]byte(`{"error":"invalid_grant","error_description":"bad code"}`))
	if !errors.Is(err, ErrProviderResponse) {
		t.Fatalf("expected provider response error, got %v", err)
	}
}

func TestDoGetHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := DoGet(context.Background(), server.Client(), server.URL, nil, "test")
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %v", err)
	}
	if providerErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("unexpected status code: %d", providerErr.StatusCode)
	}
}

func TestDoGetRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", int(maxResponseBodyBytes)+1)))
	}))
	defer server.Close()

	if _, err := DoGet(context.Background(), server.Client(), server.URL, nil, "test"); err == nil {
		t.Fatal("DoGet() error = nil, want oversized response error")
	}
}
