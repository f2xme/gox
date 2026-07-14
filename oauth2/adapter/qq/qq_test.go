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

func TestExchangeTokenJSONP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "text/html;charset=utf-8")
			_, _ = w.Write([]byte(`callback( {"access_token":"tok","expires_in":"3600","refresh_token":"ref"} );`))
		case "/me":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"client_id": "appid",
				"openid":    "oid",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL+"/token", server.URL+"/me", ""))
	token, err := provider.Exchange(context.Background(), "code")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "tok" || token.OpenID != "oid" || token.ExpiresIn != 3600 {
		t.Fatalf("unexpected token: %#v", token)
	}
}

func TestExchangeTokenForm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			_, _ = w.Write([]byte(`access_token=form-tok&expires_in=120&refresh_token=form-ref`))
		case "/me":
			_ = json.NewEncoder(w).Encode(map[string]any{"openid": "oid"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL+"/token", server.URL+"/me", ""))
	token, err := provider.Exchange(context.Background(), "code")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "form-tok" || token.OpenID != "oid" {
		t.Fatalf("unexpected token: %#v", token)
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
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) || pe.Code != "100001" || pe.Message != "invalid code" {
		t.Fatalf("unexpected provider error: %#v (%v)", pe, err)
	}
}

func TestExchangeProviderErrorStringCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// error 为字符串时旧实现会解 JSON 失败，再被误判为 invalid token
		_, _ = w.Write([]byte(`{"error":"100010","error_description":"redirect uri is illegal"}`))
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", ""))
	_, err := provider.Exchange(context.Background(), "code")
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) {
		t.Fatalf("expected ProviderError, got %v", err)
	}
	if pe.Code != "100010" {
		t.Fatalf("code = %q, want 100010", pe.Code)
	}
}

func TestExchangeUnreadableBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-a-token-body`))
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL, "", ""))
	_, err := provider.Exchange(context.Background(), "code")
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) || pe.Code != "decode_token" {
		t.Fatalf("expected decode_token ProviderError, got %v", err)
	}
	if !errors.Is(err, oauth2.ErrProviderResponse) {
		t.Fatalf("expected ErrProviderResponse, got %v", err)
	}
}

func TestFillOpenIDProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "access"})
		case "/me":
			_, _ = w.Write([]byte(`callback( {"error":100016,"error_description":"access token check failed"} );`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := New(WithEndpoints("", server.URL+"/token", server.URL+"/me", ""))
	_, err := provider.Exchange(context.Background(), "code")
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) || pe.Code != "100016" {
		t.Fatalf("expected openid ProviderError, got %v", err)
	}
}

func TestParseTokenBody(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		body   string
		wantOK bool
		token  string
		code   string
		exp    int64
	}{
		{
			name:   "json success number expires",
			body:   `{"access_token":"a","expires_in":3600}`,
			wantOK: true,
			token:  "a",
			exp:    3600,
		},
		{
			name:   "json success string expires",
			body:   `{"access_token":"a","expires_in":"7200"}`,
			wantOK: true,
			token:  "a",
			exp:    7200,
		},
		{
			name:   "json error number",
			body:   `{"error":100020,"error_description":"code is reused error"}`,
			wantOK: true,
			code:   "100020",
		},
		{
			name:   "form success",
			body:   `access_token=abc&expires_in=1`,
			wantOK: true,
			token:  "abc",
			exp:    1,
		},
		{
			name:   "empty",
			body:   ``,
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			access, _, _, _, code, _, exp, ok := parseTokenBody([]byte(tt.body))
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if access != tt.token || code != tt.code || exp != tt.exp {
				t.Fatalf("access=%q code=%q exp=%d", access, code, exp)
			}
		})
	}
}

func TestTrimCallback(t *testing.T) {
	got := string(trimCallback([]byte(`callback( {"openid":"x"} );`)))
	if got != `{"openid":"x"}` {
		t.Fatalf("got %q", got)
	}
	got = string(trimCallback([]byte(`  {"openid":"x"}  `)))
	if got != `{"openid":"x"}` {
		t.Fatalf("got %q", got)
	}
}
