package qq

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/f2xme/gox/oauth2"
)

func TestAuthCodeURL(t *testing.T) {
	provider := New(WithClientID("appid"), WithRedirectURL("https://example.com/callback"))
	got := provider.AuthCodeURL("state")
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "graph.qq.com" {
		t.Fatalf("unexpected host: %s", u.Host)
	}
	if got := u.Query().Get("scope"); got != "get_user_info" {
		t.Fatalf("unexpected scope: %s", got)
	}
}

func TestDefaultHTTPClientHasTimeout(t *testing.T) {
	if got := defaultOptions().HTTPClient.Timeout; got != defaultHTTPTimeout {
		t.Fatalf("default HTTP timeout = %v, want %v", got, defaultHTTPTimeout)
	}
}

func TestExchangeAndUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "access",
				"expires_in":    7776000,
				"refresh_token": "refresh",
			})
		case "/me":
			_, _ = w.Write([]byte(`callback( {"client_id":"appid","openid":"openid","unionid":"unionid"} );`))
		case "/user":
			if r.URL.Query().Get("openid") != "openid" {
				t.Fatalf("unexpected openid: %s", r.URL.Query().Get("openid"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ret":            0,
				"nickname":       "qq-user",
				"figureurl_qq_2": "https://example.com/avatar.jpg",
				"gender":         "男",
				"province":       "广东",
				"city":           "深圳",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := New(
		WithClientID("appid"),
		WithRedirectURL("https://example.com/callback"),
		WithEndpoints("", server.URL+"/token", server.URL+"/me", server.URL+"/user"),
	)
	token, err := provider.Exchange(context.Background(), "code")
	if err != nil {
		t.Fatal(err)
	}
	if token.OpenID != "openid" || token.UnionID != "unionid" {
		t.Fatalf("unexpected token: %#v", token)
	}

	user, err := provider.UserInfo(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if user.Provider != providerName || user.ID != "openid" || user.Nickname != "qq-user" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

func TestExchangeProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":             100001,
			"error_description": "invalid code",
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", ""))
	_, err := provider.Exchange(context.Background(), "bad")
	if !errors.Is(err, oauth2.ErrProviderResponse) {
		t.Fatalf("expected provider response error, got %v", err)
	}
}
