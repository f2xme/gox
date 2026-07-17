package baidu

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/f2xme/gox/idverify"
)

func TestBaiduVerifySuccessAndTokenCache(t *testing.T) {
	var tokenCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "token") || r.URL.Path == "/token":
			atomic.AddInt32(&tokenCalls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok-1", "expires_in": 7200})
		case strings.Contains(r.URL.Path, "idmatch") || r.URL.Path == "/idmatch":
			if r.URL.Query().Get("access_token") != "tok-1" {
				t.Fatalf("token=%s", r.URL.Query().Get("access_token"))
			}
			body, _ := io.ReadAll(r.Body)
			var req map[string]string
			_ = json.Unmarshal(body, &req)
			if req["name"] != "张三" || req["id_card_number"] != "110101199001011234" {
				t.Fatalf("body=%s", body)
			}
			_, _ = w.Write([]byte(`{"error_code":0,"error_msg":"SUCCESS"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	v, err := New(
		WithAPIKey("ak"),
		WithSecretKey("sk"),
		WithHTTPClient(srv.Client()),
		WithTokenURL(srv.URL+"/token"),
		WithMatchURL(srv.URL+"/idmatch"),
	)
	if err != nil {
		t.Fatal(err)
	}

	res, err := v.Verify(context.Background(), idverify.Request{Name: "张三", IDNumber: "110101199001011234"})
	if err != nil || !res.Matched {
		t.Fatalf("%+v %v", res, err)
	}
	_, err = v.Verify(context.Background(), idverify.Request{Name: "张三", IDNumber: "110101199001011234"})
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&tokenCalls) != 1 {
		t.Fatalf("tokenCalls=%d", tokenCalls)
	}
}

func TestBaiduBizErrors(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"222351", idverify.CodeNameMismatch},
		{"222022", idverify.CodeIDInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			srv := baiduStub(t, tt.code, "biz")
			v, err := New(WithAPIKey("ak"), WithSecretKey("sk"), WithHTTPClient(srv.Client()),
				WithTokenURL(srv.URL+"/token"), WithMatchURL(srv.URL+"/idmatch"))
			if err != nil {
				t.Fatal(err)
			}
			res, err := v.Verify(context.Background(), idverify.Request{Name: "李四", IDNumber: "110101199001011234"})
			if err != nil || res.Matched || res.ErrorCode != tt.want {
				t.Fatalf("%+v %v", res, err)
			}
		})
	}
}

func TestBaiduUnknownCodeSystemError(t *testing.T) {
	srv := baiduStub(t, "999999", "unknown")
	v, err := New(WithAPIKey("ak"), WithSecretKey("sk"), WithHTTPClient(srv.Client()),
		WithTokenURL(srv.URL+"/token"), WithMatchURL(srv.URL+"/idmatch"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = v.Verify(context.Background(), idverify.Request{Name: "王五", IDNumber: "110101199001011234"})
	if err == nil || !errors.Is(err, idverify.ErrUnavailable) {
		t.Fatalf("err=%v", err)
	}
}

func TestBaiduNotConfigured(t *testing.T) {
	_, err := New(WithAPIKey(""), WithSecretKey(""))
	if !errors.Is(err, idverify.ErrNotConfigured) {
		t.Fatalf("err=%v", err)
	}
}

func baiduStub(t *testing.T, errorCode, errorMsg string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "token") {
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"error_code": errorCode, "error_msg": errorMsg})
	}))
	t.Cleanup(srv.Close)
	return srv
}
