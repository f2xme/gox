package facebook

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/f2xme/gox/oauth2"
)

func TestAuthCodeURL(t *testing.T) {
	provider := New(
		WithClientID("app-id"),
		WithRedirectURL("https://example.com/callback"),
	)

	got := provider.AuthCodeURL("state", oauth2.WithAuthParam("auth_type", "rerequest"))
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "www.facebook.com" || u.Path != "/v26.0/dialog/oauth" {
		t.Fatalf("unexpected url: %s", got)
	}
	if got := u.Query().Get("client_id"); got != "app-id" {
		t.Fatalf("unexpected client_id: %s", got)
	}
	if got := u.Query().Get("redirect_uri"); got != "https://example.com/callback" {
		t.Fatalf("unexpected redirect_uri: %s", got)
	}
	if got := u.Query().Get("response_type"); got != "code" {
		t.Fatalf("unexpected response_type: %s", got)
	}
	if got := u.Query().Get("scope"); got != "public_profile,email" {
		t.Fatalf("unexpected scope: %s", got)
	}
	if got := u.Query().Get("state"); got != "state" {
		t.Fatalf("unexpected state: %s", got)
	}
	if got := u.Query().Get("auth_type"); got != "rerequest" {
		t.Fatalf("unexpected auth_type: %s", got)
	}
}

func TestAuthCodeURLOverrideScope(t *testing.T) {
	provider := New(WithClientID("app-id"), WithRedirectURL("https://example.com/callback"))
	got := provider.AuthCodeURL("state", oauth2.WithScopes("email"))
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if got := u.Query().Get("scope"); got != "email" {
		t.Fatalf("unexpected scope: %s", got)
	}
}

func TestDefaultHTTPClientHasTimeout(t *testing.T) {
	if got := defaultOptions().HTTPClient.Timeout; got != defaultHTTPTimeout {
		t.Fatalf("default HTTP timeout = %v, want %v", got, defaultHTTPTimeout)
	}
}

func TestName(t *testing.T) {
	if got := New().Name(); got != providerName {
		t.Fatalf("name = %s", got)
	}
}

func TestExchangeAndUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			query := r.URL.Query()
			if query.Get("code") != "code" || query.Get("client_id") != "app-id" || query.Get("client_secret") != "secret" {
				t.Fatalf("unexpected query: %s", query.Encode())
			}
			if query.Get("redirect_uri") != "https://example.com/callback" {
				t.Fatalf("unexpected redirect_uri: %s", query.Get("redirect_uri"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "short-lived",
				"token_type":   "bearer",
				"expires_in":   3600,
			})
		case "/me":
			if got := r.Header.Get("Authorization"); got != "Bearer short-lived" {
				t.Fatalf("unexpected authorization: %s", got)
			}
			if got := r.URL.Query().Get("access_token"); got != "" {
				t.Fatalf("access token leaked into query: %s", got)
			}
			if got := r.URL.Query().Get("fields"); got != defaultUserFields {
				t.Fatalf("unexpected fields: %s", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":    "102030",
				"name":  "Ada Lovelace",
				"email": "ada@example.com",
				"picture": map[string]any{
					"data": map[string]any{
						"url": "https://example.com/avatar.jpg",
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := New(
		WithClientID("app-id"),
		WithClientSecret("secret"),
		WithRedirectURL("https://example.com/callback"),
		WithEndpoints("", server.URL+"/token", "", server.URL+"/me"),
	)
	token, err := provider.Exchange(context.Background(), "code")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "short-lived" || token.TokenType != "bearer" || token.RefreshToken != "" {
		t.Fatalf("unexpected token: %#v", token)
	}
	if token.ExpiresIn != 3600 || token.Expiry.IsZero() {
		t.Fatalf("unexpected expiry: %#v", token)
	}

	user, err := provider.UserInfo(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if user.Provider != providerName || user.ID != "102030" || user.OpenID != user.ID || user.UnionID != "" {
		t.Fatalf("unexpected user id: %#v", user)
	}
	if user.Nickname != "Ada Lovelace" || user.AvatarURL != "https://example.com/avatar.jpg" {
		t.Fatalf("unexpected profile: %#v", user)
	}
	if !strings.Contains(string(user.Raw), "ada@example.com") {
		t.Fatalf("email should stay in raw: %s", user.Raw)
	}
}

func TestRefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		query := r.URL.Query()
		if query.Get("grant_type") != "fb_exchange_token" || query.Get("fb_exchange_token") != "short-lived" {
			t.Fatalf("unexpected query: %s", query.Encode())
		}
		if query.Get("client_id") != "app-id" || query.Get("client_secret") != "secret" {
			t.Fatalf("unexpected client: %s", query.Encode())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "long-lived",
			"token_type":   "bearer",
			"expires_in":   5183944,
		})
	}))
	defer server.Close()

	provider := New(
		WithClientID("app-id"),
		WithClientSecret("secret"),
		WithEndpoints("", server.URL, "", ""),
	)
	token, err := provider.RefreshToken(context.Background(), "short-lived")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "long-lived" || token.RefreshToken != "" || token.ExpiresIn != 5183944 {
		t.Fatalf("unexpected token: %#v", token)
	}
}

func TestExchangeProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message":       "Invalid verification code format.",
				"type":          "OAuthException",
				"code":          100,
				"error_subcode": 36009,
				"fbtrace_id":    "abc",
			},
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", ""))
	_, err := provider.Exchange(context.Background(), "bad")
	if !errors.Is(err, oauth2.ErrProviderResponse) {
		t.Fatalf("expected provider response error, got %v", err)
	}
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) || pe.Code != "100" || pe.Message != "Invalid verification code format." || pe.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected provider error: %#v", pe)
	}
}

func TestUserInfoProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "Invalid OAuth access token.",
				"type":    "OAuthException",
				"code":    190,
			},
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", "", "", server.URL))
	_, err := provider.UserInfo(context.Background(), &oauth2.Token{AccessToken: "access"})
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) || pe.Code != "190" || pe.Message != "Invalid OAuth access token." || pe.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unexpected provider error: %#v (%v)", pe, err)
	}
}

func TestProviderErrorPlainBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("upstream down"))
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", ""))
	_, err := provider.Exchange(context.Background(), "code")
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) || pe.Code != "502" || pe.Message != "upstream down" {
		t.Fatalf("unexpected provider error: %#v (%v)", pe, err)
	}
}

func TestInvalidInputs(t *testing.T) {
	provider := New()
	if _, err := provider.Exchange(context.Background(), ""); !errors.Is(err, oauth2.ErrInvalidCode) {
		t.Fatalf("exchange empty code: %v", err)
	}
	if _, err := provider.RefreshToken(context.Background(), ""); !errors.Is(err, oauth2.ErrMissingRefreshToken) {
		t.Fatalf("refresh empty token: %v", err)
	}
	if _, err := provider.UserInfo(context.Background(), nil); !errors.Is(err, oauth2.ErrInvalidToken) {
		t.Fatalf("userinfo nil token: %v", err)
	}
	if _, err := provider.UserInfo(context.Background(), &oauth2.Token{}); !errors.Is(err, oauth2.ErrInvalidToken) {
		t.Fatalf("userinfo empty token: %v", err)
	}
}

func TestExchangeMissingAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"token_type": "bearer"})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", ""))
	_, err := provider.Exchange(context.Background(), "code")
	if !errors.Is(err, oauth2.ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}
}

func TestUserInfoMissingID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"name": "Ada"})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", "", "", server.URL))
	_, err := provider.UserInfo(context.Background(), &oauth2.Token{AccessToken: "access"})
	if !errors.Is(err, oauth2.ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}
}

func TestDecodeErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", server.URL))
	if _, err := provider.Exchange(context.Background(), "code"); err == nil || !strings.Contains(err.Error(), "decode token") {
		t.Fatalf("expected token decode error, got %v", err)
	}
	if _, err := provider.UserInfo(context.Background(), &oauth2.Token{AccessToken: "access"}); err == nil || !strings.Contains(err.Error(), "decode user info") {
		t.Fatalf("expected user decode error, got %v", err)
	}
}
