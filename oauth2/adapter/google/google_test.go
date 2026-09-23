package google

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/f2xme/gox/oauth2"
)

func TestAuthCodeURL(t *testing.T) {
	provider := New(
		WithClientID("client-id"),
		WithRedirectURL("https://example.com/callback"),
	)

	got := provider.AuthCodeURL("state", oauth2.WithAuthParam("hd", "example.com"))
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "accounts.google.com" || u.Path != "/o/oauth2/v2/auth" {
		t.Fatalf("unexpected url: %s", got)
	}
	if got := u.Query().Get("client_id"); got != "client-id" {
		t.Fatalf("unexpected client_id: %s", got)
	}
	if got := u.Query().Get("redirect_uri"); got != "https://example.com/callback" {
		t.Fatalf("unexpected redirect_uri: %s", got)
	}
	if got := u.Query().Get("response_type"); got != "code" {
		t.Fatalf("unexpected response_type: %s", got)
	}
	if got := u.Query().Get("scope"); got != "openid email profile" {
		t.Fatalf("unexpected scope: %s", got)
	}
	if got := u.Query().Get("state"); got != "state" {
		t.Fatalf("unexpected state: %s", got)
	}
	if got := u.Query().Get("access_type"); got != "offline" {
		t.Fatalf("unexpected access_type: %s", got)
	}
	if got := u.Query().Get("hd"); got != "example.com" {
		t.Fatalf("unexpected hd: %s", got)
	}
}

func TestAuthCodeURLOverride(t *testing.T) {
	provider := New(WithClientID("client-id"), WithRedirectURL("https://example.com/callback"))
	got := provider.AuthCodeURL("state",
		oauth2.WithScopes("openid", "email"),
		oauth2.WithAuthParam("access_type", "online"),
		oauth2.WithAuthParam("prompt", "consent"),
	)
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if got := u.Query().Get("scope"); got != "openid email" {
		t.Fatalf("unexpected scope: %s", got)
	}
	if got := u.Query().Get("access_type"); got != "online" {
		t.Fatalf("unexpected access_type: %s", got)
	}
	if got := u.Query().Get("prompt"); got != "consent" {
		t.Fatalf("unexpected prompt: %s", got)
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
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			form, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatal(err)
			}
			if form.Get("grant_type") != "authorization_code" || form.Get("code") != "code" {
				t.Fatalf("unexpected form: %s", form.Encode())
			}
			if form.Get("client_id") != "client-id" || form.Get("client_secret") != "secret" {
				t.Fatalf("unexpected client: %s", form.Encode())
			}
			if form.Get("redirect_uri") != "https://example.com/callback" {
				t.Fatalf("unexpected redirect_uri: %s", form.Get("redirect_uri"))
			}
			if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/x-www-form-urlencoded") {
				t.Fatalf("unexpected content type: %s", ct)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "access",
				"expires_in":    3599,
				"refresh_token": "refresh",
				"scope":         "openid email profile",
				"token_type":    "Bearer",
				"id_token":      "header.payload.sig",
			})
		case "/userinfo":
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer access" {
				t.Fatalf("unexpected authorization: %s", got)
			}
			if r.URL.RawQuery != "" {
				t.Fatalf("access token leaked into query: %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"sub":            "110169484474386276334",
				"name":           "Ada Lovelace",
				"picture":        "https://example.com/avatar.jpg",
				"email":          "ada@example.com",
				"email_verified": true,
				"locale":         "en",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := New(
		WithClientID("client-id"),
		WithClientSecret("secret"),
		WithRedirectURL("https://example.com/callback"),
		WithEndpoints("", server.URL+"/token", "", server.URL+"/userinfo"),
	)
	token, err := provider.Exchange(context.Background(), "code")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "access" || token.RefreshToken != "refresh" || token.TokenType != "Bearer" {
		t.Fatalf("unexpected token: %#v", token)
	}
	if token.ExpiresIn != 3599 || token.Expiry.IsZero() || token.Scope != "openid email profile" {
		t.Fatalf("unexpected token expiry or scope: %#v", token)
	}
	if !strings.Contains(string(token.Raw), "id_token") {
		t.Fatalf("id_token should stay in raw: %s", token.Raw)
	}

	user, err := provider.UserInfo(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if user.Provider != providerName || user.ID != "110169484474386276334" || user.OpenID != user.ID {
		t.Fatalf("unexpected user id: %#v", user)
	}
	if user.UnionID != "" || user.Nickname != "Ada Lovelace" || user.AvatarURL != "https://example.com/avatar.jpg" {
		t.Fatalf("unexpected user profile: %#v", user)
	}
	if !strings.Contains(string(user.Raw), "ada@example.com") {
		t.Fatalf("email should stay in raw: %s", user.Raw)
	}
}

func TestUserInfoLegacyID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "legacy-id",
			"given_name":  "Ada",
			"family_name": "Lovelace",
			"picture":     "https://example.com/avatar.jpg",
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", "", "", server.URL))
	user, err := provider.UserInfo(context.Background(), &oauth2.Token{AccessToken: "access"})
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != "legacy-id" || user.Nickname != "Ada Lovelace" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

func TestRefreshTokenKeepsOriginal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		form, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatal(err)
		}
		if form.Get("grant_type") != "refresh_token" || form.Get("refresh_token") != "refresh" {
			t.Fatalf("unexpected form: %s", form.Encode())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "new-access",
			"expires_in":   3600,
			"token_type":   "Bearer",
			"scope":        "openid email",
		})
	}))
	defer server.Close()

	provider := New(
		WithClientID("client-id"),
		WithClientSecret("secret"),
		WithEndpoints("", server.URL, "", ""),
	)
	token, err := provider.RefreshToken(context.Background(), "refresh")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "new-access" || token.RefreshToken != "refresh" {
		t.Fatalf("unexpected token: %#v", token)
	}
}

func TestRefreshTokenReplacesRefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access",
			"expires_in":    3600,
			"refresh_token": "rotated",
			"token_type":    "Bearer",
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", "", server.URL, ""))
	token, err := provider.RefreshToken(context.Background(), "refresh")
	if err != nil {
		t.Fatal(err)
	}
	if token.RefreshToken != "rotated" {
		t.Fatalf("unexpected refresh token: %#v", token)
	}
}

func TestExchangeProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":             "invalid_grant",
			"error_description": "Bad Request",
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", ""))
	_, err := provider.Exchange(context.Background(), "bad")
	if !errors.Is(err, oauth2.ErrProviderResponse) {
		t.Fatalf("expected provider response error, got %v", err)
	}
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) || pe.Code != "invalid_grant" || pe.Message != "Bad Request" || pe.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected provider error: %#v", pe)
	}
}

func TestUserInfoProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    401,
				"message": "Invalid Credentials",
				"status":  "UNAUTHENTICATED",
			},
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", "", "", server.URL))
	_, err := provider.UserInfo(context.Background(), &oauth2.Token{AccessToken: "access"})
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) || pe.Code != "UNAUTHENTICATED" || pe.Message != "Invalid Credentials" || pe.StatusCode != http.StatusUnauthorized {
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
		_ = json.NewEncoder(w).Encode(map[string]any{"token_type": "Bearer"})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", ""))
	_, err := provider.Exchange(context.Background(), "code")
	if !errors.Is(err, oauth2.ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}
}

func TestUserInfoMissingSubject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"email": "ada@example.com"})
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
