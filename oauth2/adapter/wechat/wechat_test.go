package wechat

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
	provider := New(
		WithClientID("appid"),
		WithRedirectURL("https://example.com/callback"),
	)

	got := provider.AuthCodeURL("state", oauth2.WithAuthParam("login_type", "jssdk"))
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "open.weixin.qq.com" {
		t.Fatalf("unexpected host: %s", u.Host)
	}
	if got := u.Query().Get("appid"); got != "appid" {
		t.Fatalf("unexpected appid: %s", got)
	}
	if got := u.Query().Get("scope"); got != "snsapi_login" {
		t.Fatalf("unexpected scope: %s", got)
	}
	if u.Fragment != "wechat_redirect" {
		t.Fatalf("unexpected fragment: %s", u.Fragment)
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
			if r.URL.Query().Get("code") != "code" {
				t.Fatalf("unexpected code: %s", r.URL.Query().Get("code"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "access",
				"expires_in":    7200,
				"refresh_token": "refresh",
				"openid":        "openid",
				"unionid":       "unionid",
				"scope":         "snsapi_login",
			})
		case "/user":
			if r.URL.Query().Get("access_token") != "access" {
				t.Fatalf("unexpected access token: %s", r.URL.Query().Get("access_token"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"openid":     "openid",
				"unionid":    "unionid",
				"nickname":   "张三",
				"sex":        1,
				"province":   "广东",
				"city":       "深圳",
				"country":    "CN",
				"headimgurl": "https://example.com/avatar.jpg",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL+"/token", "", server.URL+"/user"))
	token, err := provider.Exchange(context.Background(), "code")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "access" || token.OpenID != "openid" || token.UnionID != "unionid" {
		t.Fatalf("unexpected token: %#v", token)
	}

	user, err := provider.UserInfo(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if user.Provider != providerName || user.ID != "openid" || user.Nickname != "张三" || user.Gender != "male" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

func TestExchangeProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 40029,
			"errmsg":  "invalid code",
		})
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", ""))
	_, err := provider.Exchange(context.Background(), "bad")
	if !errors.Is(err, oauth2.ErrProviderResponse) {
		t.Fatalf("expected provider response error, got %v", err)
	}
}
