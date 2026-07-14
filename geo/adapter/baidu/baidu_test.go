package baidu

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"

	"github.com/f2xme/gox/geo"
)

func TestLookupSuccessUTF8(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") != "114.247.50.2" {
			t.Errorf("unexpected query: %s", r.URL.Query().Get("query"))
		}
		if r.URL.Query().Get("resource_id") != "6006" {
			t.Errorf("unexpected resource_id")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "0",
			"data": []map[string]string{
				{
					"location": "广东省深圳市 电信",
					"origip":   "114.247.50.2",
				},
			},
		})
	}))
	defer server.Close()

	locator, err := New(WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	loc, err := locator.Lookup(context.Background(), "114.247.50.2")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if loc.Country != "中国" || loc.CountryCode != "CN" {
		t.Fatalf("Country/Code = %q/%q", loc.Country, loc.CountryCode)
	}
	if loc.Province != "广东省" || loc.City != "深圳市" || loc.ISP != "电信" {
		t.Fatalf("unexpected location: %+v", loc)
	}
	if loc.Extra["location"] != "广东省深圳市 电信" {
		t.Fatalf("extra location = %q", loc.Extra["location"])
	}
}

func TestLookupSuccessGBK(t *testing.T) {
	t.Parallel()

	utf8JSON := `{"status":"0","data":[{"location":"北京市 联通","origip":"1.2.3.4"}]}`
	gbkBody, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(utf8JSON))
	if err != nil {
		t.Fatalf("encode gbk: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=gbk")
		_, _ = w.Write(gbkBody)
	}))
	defer server.Close()

	locator, err := New(WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	loc, err := locator.Lookup(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if loc.Province != "北京市" || loc.City != "北京市" || loc.ISP != "联通" {
		t.Fatalf("unexpected location: %+v", loc)
	}
}

func TestLookupNumericStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// status 为数字 0
		_, _ = w.Write([]byte(`{"status":0,"data":[{"location":"浙江省杭州市 移动"}]}`))
	}))
	defer server.Close()

	locator, err := New(WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	loc, err := locator.Lookup(context.Background(), "1.1.1.1")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if loc.Province != "浙江省" || loc.City != "杭州市" {
		t.Fatalf("unexpected location: %+v", loc)
	}
}

func TestLookupOverseas(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "0",
			"data": []map[string]string{
				{"location": "美国"},
			},
		})
	}))
	defer server.Close()

	locator, err := New(WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	loc, err := locator.Lookup(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if loc.Country != "美国" {
		t.Fatalf("Country = %q, want 美国", loc.Country)
	}
	if loc.CountryCode == "CN" {
		t.Fatal("overseas IP should not set CountryCode CN")
	}
	if loc.Province != "" {
		t.Fatalf("Province should be empty for bare country, got %q", loc.Province)
	}
}

func TestLookupNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "0",
			"data":   []any{},
		})
	}))
	defer server.Close()

	locator, err := New(WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = locator.Lookup(context.Background(), "8.8.8.8")
	if !geo.IsNotFound(err) {
		t.Fatalf("Lookup() error = %v, want NotFound", err)
	}
}

func TestLookupErrors(t *testing.T) {
	t.Parallel()

	t.Run("invalid ip", func(t *testing.T) {
		t.Parallel()
		locator, err := New()
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		_, err = locator.Lookup(context.Background(), "bad")
		if !geo.IsInvalidIP(err) {
			t.Fatalf("Lookup() error = %v", err)
		}
	})

	t.Run("non-zero status", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "1", "data": []any{}})
		}))
		defer server.Close()

		locator, err := New(WithEndpoint(server.URL))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		_, err = locator.Lookup(context.Background(), "1.2.3.4")
		if !geo.IsUnavailable(err) {
			t.Fatalf("Lookup() error = %v", err)
		}
	})

	t.Run("http 502", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		locator, err := New(WithEndpoint(server.URL))
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
		locator, err := New()
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = locator.Lookup(ctx, "1.2.3.4")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestParseBaiduLocation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in       string
		country  string
		code     string
		province string
		city     string
		district string
		isp      string
	}{
		{in: "广东省深圳市 电信", country: "中国", code: "CN", province: "广东省", city: "深圳市", isp: "电信"},
		{in: "北京市 联通", country: "中国", code: "CN", province: "北京市", city: "北京市", isp: "联通"},
		{in: "浙江省杭州市", country: "中国", code: "CN", province: "浙江省", city: "杭州市"},
		{in: "内蒙古自治区呼和浩特市 移动", country: "中国", code: "CN", province: "内蒙古自治区", city: "呼和浩特市", isp: "移动"},
		{in: "香港特别行政区", country: "中国", code: "CN", province: "香港特别行政区", city: "香港特别行政区"},
		{in: "重庆市渝中区", country: "中国", code: "CN", province: "重庆市", city: "重庆市", district: "渝中区"},
		{in: "美国", country: "美国"},
		{in: "日本 东京都", country: "日本", province: "东京都"},
		// 单独「xx市」不得默认判为中国
		{in: "大阪市", country: "大阪市"},
		{in: "首尔市", country: "首尔市"},
		// 海外 ISP 识别
		{in: "美国 AT&T", country: "美国", isp: "AT&T"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got := parseBaiduLocation(tt.in)
			if got.Country != tt.country ||
				got.CountryCode != tt.code ||
				got.Province != tt.province ||
				got.City != tt.city ||
				got.District != tt.district ||
				got.ISP != tt.isp {
				t.Fatalf("parseBaiduLocation(%q) = %+v", tt.in, got)
			}
		})
	}
}

func TestLookupPrivateLocation(t *testing.T) {
	t.Parallel()

	cases := []string{"局域网", "内网IP", "本机地址", "private network"}
	for _, location := range cases {
		location := location
		t.Run(location, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"status": "0",
					"data":   []map[string]string{{"location": location}},
				})
			}))
			defer server.Close()

			locator, err := New(WithEndpoint(server.URL))
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			_, err = locator.Lookup(context.Background(), "127.0.0.1")
			if !geo.IsNotFound(err) {
				t.Fatalf("Lookup(%q) error = %v, want NotFound", location, err)
			}
		})
	}
}

func TestIsChineseAdminRegion(t *testing.T) {
	t.Parallel()
	tests := map[string]bool{
		"广东省深圳市":  true,
		"北京市":     true,
		"重庆市渝中区":  true,
		"香港特别行政区": true,
		"大阪市":     false,
		"首尔市":     false,
		"美国":      false,
		"纽约市":     false,
	}
	for in, want := range tests {
		if got := isChineseAdminRegion(in); got != want {
			t.Fatalf("isChineseAdminRegion(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestNewZeroTimeoutFallsBack(t *testing.T) {
	t.Parallel()
	locator, err := New(WithEndpoint("http://example.com"), WithTimeout(0))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if locator.client == nil || locator.client.Timeout != 5*time.Second {
		t.Fatalf("timeout = %v, want 5s", locator.client.Timeout)
	}
}

func TestDecodeResponseBody(t *testing.T) {
	t.Parallel()

	utf8Text := `{"ok":true}`
	got, err := decodeResponseBody([]byte(utf8Text), "application/json; charset=utf-8")
	if err != nil || got != utf8Text {
		t.Fatalf("utf8 decode = %q, err=%v", got, err)
	}

	src := "北京市"
	gbk, err := simplifiedchinese.GBK.NewEncoder().String(src)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err = decodeResponseBody([]byte(gbk), "text/html; charset=gbk")
	if err != nil || got != src {
		t.Fatalf("gbk decode = %q, err=%v", got, err)
	}

	// 无 charset 且非法 UTF-8 时回退 GBK
	got, err = decodeResponseBody([]byte(gbk), "text/html")
	if err != nil || got != src {
		t.Fatalf("fallback gbk decode = %q, err=%v", got, err)
	}
}

func TestParseCharset(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"":                                   "",
		"application/json":                   "",
		"text/html; charset=gbk":             "gbk",
		`text/html; charset="GB2312"`:        "gb2312",
		"application/json;charset=utf-8;q=1": "utf-8",
	}
	for in, want := range tests {
		if got := parseCharset(in); got != want {
			t.Fatalf("parseCharset(%q) = %q, want %q", in, got, want)
		}
	}
}
