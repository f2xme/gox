package x

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
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

	got := provider.AuthCodeURL("state", oauth2.WithAuthParam("code_challenge", "challenge"))
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "x.com" || u.Path != "/i/oauth2/authorize" {
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
	if got := u.Query().Get("scope"); got != "tweet.read users.read offline.access" {
		t.Fatalf("unexpected scope: %s", got)
	}
	if got := u.Query().Get("state"); got != "state" {
		t.Fatalf("unexpected state: %s", got)
	}
	if got := u.Query().Get("code_challenge"); got != "challenge" {
		t.Fatalf("unexpected code_challenge: %s", got)
	}
	if got := u.Query().Get("code_challenge_method"); got != "S256" {
		t.Fatalf("unexpected code_challenge_method: %s", got)
	}
}

func TestAuthCodeURLWithoutPKCE(t *testing.T) {
	provider := New(WithClientID("client-id"), WithRedirectURL("https://example.com/callback"))
	u, err := url.Parse(provider.AuthCodeURL("state", oauth2.WithScopes("users.read")))
	if err != nil {
		t.Fatal(err)
	}
	if got := u.Query().Get("scope"); got != "users.read" {
		t.Fatalf("unexpected scope: %s", got)
	}
	if got := u.Query().Get("code_challenge"); got != "" {
		t.Fatalf("auth URL invented a challenge: %s", got)
	}
}

func TestBeginPKCE(t *testing.T) {
	provider := New(WithClientID("client-id"), WithRedirectURL("https://example.com/callback"))
	authURL, verifier, err := provider.Begin("state", oauth2.WithAuthParam("code_challenge", "ignored"))
	if err != nil {
		t.Fatal(err)
	}
	if len(verifier) < 43 {
		t.Fatalf("verifier too short: %d", len(verifier))
	}
	u, err := url.Parse(authURL)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(verifier))
	want := base64.RawURLEncoding.EncodeToString(sum[:])
	if got := u.Query().Get("code_challenge"); got != want {
		t.Fatalf("challenge = %s, want %s", got, want)
	}
	if got := u.Query().Get("code_challenge_method"); got != "S256" {
		t.Fatalf("unexpected method: %s", got)
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
			user, pass, ok := r.BasicAuth()
			if !ok || user != "client-id" || pass != "secret" {
				t.Fatalf("unexpected basic auth: %s %s %v", user, pass, ok)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			form, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatal(err)
			}
			if form.Get("grant_type") != "authorization_code" || form.Get("code") != "code" || form.Get("code_verifier") != "verifier" {
				t.Fatalf("unexpected form: %s", form.Encode())
			}
			if form.Get("client_id") != "client-id" || form.Get("redirect_uri") != "https://example.com/callback" {
				t.Fatalf("unexpected client or redirect: %s", form.Encode())
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"token_type":    "bearer",
				"expires_in":    7200,
				"access_token":  "access",
				"scope":         "tweet.read users.read offline.access",
				"refresh_token": "refresh",
			})
		case "/me":
			if got := r.Header.Get("Authorization"); got != "Bearer access" {
				t.Fatalf("unexpected authorization: %s", got)
			}
			if got := r.URL.Query().Get("user.fields"); got != defaultUserFields {
				t.Fatalf("unexpected fields: %s", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":                "2244994945",
					"name":              "Ada Lovelace",
					"username":          "ada",
					"profile_image_url": "https://example.com/avatar.jpg",
					"description":       "bio",
					"location":          "London",
				},
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
		WithEndpoints("", server.URL+"/token", "", server.URL+"/me"),
	)
	token, err := provider.Exchange(WithCodeVerifier(context.Background(), "verifier"), "code")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "access" || token.RefreshToken != "refresh" || token.TokenType != "bearer" {
		t.Fatalf("unexpected token: %#v", token)
	}
	if token.ExpiresIn != 7200 || token.Expiry.IsZero() || token.Scope != "tweet.read users.read offline.access" {
		t.Fatalf("unexpected expiry or scope: %#v", token)
	}

	user, err := provider.UserInfo(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if user.Provider != providerName || user.ID != "2244994945" || user.OpenID != user.ID || user.UnionID != "" {
		t.Fatalf("unexpected user id: %#v", user)
	}
	if user.Nickname != "Ada Lovelace" || user.AvatarURL != "https://example.com/avatar.jpg" {
		t.Fatalf("unexpected profile: %#v", user)
	}
	if !strings.Contains(string(user.Raw), `"username":"ada"`) && !strings.Contains(string(user.Raw), `"username": "ada"`) {
		t.Fatalf("username should stay in raw: %s", user.Raw)
	}
}

func TestExchangeCodePublicClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := r.BasicAuth(); ok {
			t.Fatal("public client should not send basic auth")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "access",
			"token_type":   "bearer",
			"expires_in":   7200,
		})
	}))
	defer server.Close()

	provider := New(WithClientID("client-id"), WithEndpoints("", server.URL, "", ""))
	token, err := provider.ExchangeCode(context.Background(), "code", "verifier")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "access" || token.RefreshToken != "" {
		t.Fatalf("unexpected token: %#v", token)
	}
}

func TestRefreshTokenRotation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		form, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatal(err)
		}
		if form.Get("grant_type") != "refresh_token" || form.Get("refresh_token") != "old-refresh" {
			t.Fatalf("unexpected form: %s", form.Encode())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access",
			"refresh_token": "new-refresh",
			"token_type":    "bearer",
			"expires_in":    7200,
		})
	}))
	defer server.Close()

	provider := New(WithClientID("client-id"), WithClientSecret("secret"), WithEndpoints("", server.URL, "", ""))
	token, err := provider.RefreshToken(context.Background(), "old-refresh")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "new-access" || token.RefreshToken != "new-refresh" {
		t.Fatalf("unexpected token: %#v", token)
	}
}

func TestRefreshTokenKeepsOriginal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "new-access",
			"token_type":   "bearer",
			"expires_in":   7200,
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", "", server.URL, ""))
	token, err := provider.RefreshToken(context.Background(), "old-refresh")
	if err != nil {
		t.Fatal(err)
	}
	if token.RefreshToken != "old-refresh" {
		t.Fatalf("unexpected refresh token: %#v", token)
	}
}

func TestUserInfoUsernameFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": "1", "username": "ada"},
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", "", "", server.URL))
	user, err := provider.UserInfo(context.Background(), &oauth2.Token{AccessToken: "access"})
	if err != nil {
		t.Fatal(err)
	}
	if user.Nickname != "ada" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

func TestExchangeProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":             "invalid_request",
			"error_description": "Value passed for the token was invalid.",
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", ""))
	_, err := provider.ExchangeCode(context.Background(), "bad", "verifier")
	if !errors.Is(err, oauth2.ErrProviderResponse) {
		t.Fatalf("expected provider response error, got %v", err)
	}
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) || pe.Code != "invalid_request" || pe.Message != "Value passed for the token was invalid." || pe.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected provider error: %#v", pe)
	}
}

func TestUserInfoProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"title":  "Unauthorized",
			"detail": "Unauthorized",
			"status": 401,
			"type":   "about:blank",
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", "", "", server.URL))
	_, err := provider.UserInfo(context.Background(), &oauth2.Token{AccessToken: "access"})
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) || pe.Code != "401" || pe.Message != "Unauthorized" || pe.StatusCode != http.StatusUnauthorized {
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
	_, err := provider.ExchangeCode(context.Background(), "code", "verifier")
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) || pe.Code != "502" || pe.Message != "upstream down" {
		t.Fatalf("unexpected provider error: %#v (%v)", pe, err)
	}
}

func TestInvalidInputs(t *testing.T) {
	provider := New()
	if _, err := provider.ExchangeCode(context.Background(), "", "verifier"); !errors.Is(err, oauth2.ErrInvalidCode) {
		t.Fatalf("empty code: %v", err)
	}
	if _, err := provider.Exchange(context.Background(), "code"); !errors.Is(err, ErrMissingCodeVerifier) {
		t.Fatalf("missing verifier: %v", err)
	}
	if _, err := provider.RefreshToken(context.Background(), ""); !errors.Is(err, oauth2.ErrMissingRefreshToken) {
		t.Fatalf("empty refresh: %v", err)
	}
	if _, err := provider.UserInfo(context.Background(), nil); !errors.Is(err, oauth2.ErrInvalidToken) {
		t.Fatalf("nil token: %v", err)
	}
}

func TestExchangeMissingAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"token_type": "bearer"})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", ""))
	_, err := provider.ExchangeCode(context.Background(), "code", "verifier")
	if !errors.Is(err, oauth2.ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}
}

func TestUserInfoMissingID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]any{{"message": "Could not authenticate you", "code": 32}},
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", "", "", server.URL))
	_, err := provider.UserInfo(context.Background(), &oauth2.Token{AccessToken: "access"})
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) || pe.Code != "32" || pe.Message != "Could not authenticate you" {
		t.Fatalf("unexpected error: %#v (%v)", pe, err)
	}
}

func TestDecodeErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", server.URL))
	if _, err := provider.ExchangeCode(context.Background(), "code", "verifier"); err == nil || !strings.Contains(err.Error(), "decode token") {
		t.Fatalf("expected token decode error, got %v", err)
	}
	if _, err := provider.UserInfo(context.Background(), &oauth2.Token{AccessToken: "access"}); err == nil || !strings.Contains(err.Error(), "decode user info") {
		t.Fatalf("expected user decode error, got %v", err)
	}
}
