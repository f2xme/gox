package douyin

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
	provider := New(WithClientID("client-key"), WithRedirectURL("https://example.com/callback"))
	got := provider.AuthCodeURL("state")
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "open.douyin.com" {
		t.Fatalf("unexpected host: %s", u.Host)
	}
	if got := u.Query().Get("client_key"); got != "client-key" {
		t.Fatalf("unexpected client_key: %s", got)
	}
	if got := u.Query().Get("scope"); got != "user_info" {
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
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			_ = r.ParseForm()
			if r.Form.Get("code") != "code" {
				t.Fatalf("unexpected code: %s", r.Form.Get("code"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"access_token":  "access",
					"expires_in":    7200,
					"refresh_token": "refresh",
					"open_id":       "openid",
					"union_id":      "unionid",
					"scope":         "user_info",
				},
			})
		case "/refresh":
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			_ = r.ParseForm()
			if r.Form.Get("refresh_token") != "refresh" {
				t.Fatalf("unexpected refresh token: %s", r.Form.Get("refresh_token"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"access_token":  "refreshed-access",
					"error_code":    0,
					"expires_in":    "86400",
					"refresh_token": "refresh",
					"open_id":       "openid",
					"scope":         "user_info",
				},
				"message": "success",
			})
		case "/user":
			_ = r.ParseForm()
			if r.Form.Get("open_id") != "openid" {
				t.Fatalf("unexpected open_id: %s", r.Form.Get("open_id"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"err_no":  0,
				"err_msg": "",
				"data": map[string]any{
					"open_id":       "openid",
					"union_id":      "unionid",
					"nickname":      "douyin-user",
					"avatar_larger": "https://example.com/avatar.jpg",
					"gender":        "male",
					"province":      "广东",
					"city":          "深圳",
					"country":       "CN",
					"error_code":    "0",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL+"/token", server.URL+"/refresh", server.URL+"/user"))
	token, err := provider.Exchange(context.Background(), "code")
	if err != nil {
		t.Fatal(err)
	}
	if token.OpenID != "openid" || token.UnionID != "unionid" {
		t.Fatalf("unexpected token: %#v", token)
	}
	refreshed, err := provider.RefreshToken(context.Background(), "refresh")
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.AccessToken != "refreshed-access" || refreshed.ExpiresIn != 86400 {
		t.Fatalf("unexpected refreshed token: %#v", refreshed)
	}

	user, err := provider.UserInfo(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if user.Provider != providerName || user.ID != "openid" || user.Nickname != "douyin-user" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

func TestExchangeProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"error_code":  2190008,
				"description": "invalid code",
			},
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", ""))
	_, err := provider.Exchange(context.Background(), "bad")
	if !errors.Is(err, oauth2.ErrProviderResponse) {
		t.Fatalf("expected provider response error, got %v", err)
	}
}

func TestUserInfoOfficialProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"err_no":  28001003,
			"err_msg": "access_token无效",
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", "", "", server.URL))
	_, err := provider.UserInfo(context.Background(), &oauth2.Token{AccessToken: "bad", OpenID: "openid"})
	if !errors.Is(err, oauth2.ErrProviderResponse) {
		t.Fatalf("expected provider response error, got %v", err)
	}
}
