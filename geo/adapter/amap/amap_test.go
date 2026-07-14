package amap

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/f2xme/gox/geo"
)

func TestNewRequiresKey(t *testing.T) {
	t.Parallel()

	_, err := New()
	if !geo.IsCode(err, geo.ErrCodeInvalidArgument) {
		t.Fatalf("New() without key error = %v", err)
	}
}

func TestLookupSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "test-key" {
			t.Errorf("unexpected key: %s", r.URL.Query().Get("key"))
		}
		if r.URL.Query().Get("ip") != "114.247.50.2" {
			t.Errorf("unexpected ip: %s", r.URL.Query().Get("ip"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":    "1",
			"info":      "OK",
			"infocode":  "10000",
			"province":  "北京市",
			"city":      "北京市",
			"adcode":    "110000",
			"rectangle": "116.0119343,39.66127144;116.7829835,40.2164962",
		})
	}))
	defer server.Close()

	locator, err := New(WithKey("test-key"), WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	loc, err := locator.Lookup(context.Background(), "114.247.50.2")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if loc.Province != "北京市" || loc.City != "北京市" {
		t.Fatalf("unexpected location: %+v", loc)
	}
	if loc.Country != "中国" || loc.CountryCode != "CN" {
		t.Fatalf("Country/Code = %q/%q", loc.Country, loc.CountryCode)
	}
	if loc.Extra["adcode"] != "110000" {
		t.Fatalf("adcode = %q", loc.Extra["adcode"])
	}
}

func TestLookupEmptyProvinceCity(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 内网 IP 场景：status=1 但 province/city 为空数组
		_, _ = w.Write([]byte(`{"status":"1","info":"OK","infocode":"10000","province":[],"city":[],"adcode":[],"rectangle":""}`))
	}))
	defer server.Close()

	locator, err := New(WithKey("test-key"), WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = locator.Lookup(context.Background(), "127.0.0.1")
	if !geo.IsNotFound(err) {
		t.Fatalf("Lookup() error = %v, want NotFound", err)
	}
}

func TestLookupAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "0",
			"info":     "INVALID_USER_KEY",
			"infocode": "10001",
		})
	}))
	defer server.Close()

	locator, err := New(WithKey("bad-key"), WithEndpoint(server.URL), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = locator.Lookup(context.Background(), "1.2.3.4")
	if !geo.IsUnavailable(err) {
		t.Fatalf("Lookup() error = %v, want Unavailable", err)
	}
}

func TestLookupHTTPErrorAndCanceledContext(t *testing.T) {
	t.Parallel()

	t.Run("http 502", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		locator, err := New(WithKey("test-key"), WithEndpoint(server.URL))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		_, err = locator.Lookup(context.Background(), "1.2.3.4")
		if !geo.IsUnavailable(err) {
			t.Fatalf("Lookup() error = %v", err)
		}
	})

	t.Run("canceled context", func(t *testing.T) {
		t.Parallel()
		locator, err := New(WithKey("test-key"), WithEndpoint("http://example.com"))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = locator.Lookup(ctx, "1.2.3.4")
		if err == nil {
			t.Fatal("expected error for canceled context")
		}
	})

	t.Run("quota exhausted", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status":   "0",
				"info":     "CUQPS_HAS_EXCEEDED_THE_LIMIT",
				"infocode": "10044",
			})
		}))
		defer server.Close()

		locator, err := New(WithKey("test-key"), WithEndpoint(server.URL))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		_, err = locator.Lookup(context.Background(), "1.2.3.4")
		if !geo.IsUnavailable(err) {
			t.Fatalf("Lookup() error = %v", err)
		}
	})
}

func TestLookupInvalidIP(t *testing.T) {
	t.Parallel()

	locator, err := New(WithKey("test-key"), WithEndpoint("http://example.com"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = locator.Lookup(context.Background(), "not-an-ip")
	if !geo.IsInvalidIP(err) {
		t.Fatalf("Lookup() error = %v", err)
	}
}

func TestDecodeAmapString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "string", raw: `"广东省"`, want: "广东省"},
		{name: "empty array", raw: `[]`, want: ""},
		{name: "null", raw: `null`, want: ""},
		{name: "empty string", raw: `""`, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := decodeAmapString(json.RawMessage(tt.raw)); got != tt.want {
				t.Fatalf("decodeAmapString(%s) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
