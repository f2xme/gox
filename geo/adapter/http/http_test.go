package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/f2xme/gox/geo"
)

func TestLookupSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/8.8.8.8" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":      "success",
			"country":     "United States",
			"countryCode": "US",
			"regionName":  "California",
			"city":        "Mountain View",
			"isp":         "Google LLC",
			"lat":         37.386,
			"lon":         -122.0838,
		})
	}))
	defer server.Close()

	locator, err := New(WithEndpoint(server.URL + "/json/"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	loc, err := locator.Lookup(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if loc.CountryCode != "US" || loc.City != "Mountain View" {
		t.Fatalf("unexpected location: %+v", loc)
	}
	if loc.Latitude == 0 || loc.Longitude == 0 {
		t.Fatalf("expected coordinates, got lat=%v lon=%v", loc.Latitude, loc.Longitude)
	}
}

func TestLookupTemplateURLAndCustomParser(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/geo" || r.URL.Query().Get("ip") != "1.1.1.1" {
			t.Errorf("unexpected request: %s", r.URL.String())
		}
		if r.Header.Get("X-Token") != "secret" {
			t.Errorf("missing header")
		}
		_, _ = w.Write([]byte(`{"c":"澳大利亚","p":"昆士兰","city":"布里斯班"}`))
	}))
	defer server.Close()

	locator, err := New(
		WithEndpoint(server.URL+"/geo?ip=%s"),
		WithHeader("X-Token", "secret"),
		WithResponseParser(func(body []byte, statusCode int, ip string) (*geo.Location, error) {
			var raw map[string]string
			if err := json.Unmarshal(body, &raw); err != nil {
				return nil, err
			}
			return &geo.Location{
				Country:  raw["c"],
				Province: raw["p"],
				City:     raw["city"],
			}, nil
		}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	loc, err := locator.Lookup(context.Background(), "1.1.1.1")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if loc.Country != "澳大利亚" || loc.City != "布里斯班" {
		t.Fatalf("unexpected location: %+v", loc)
	}
}

func TestLookupErrors(t *testing.T) {
	t.Parallel()

	t.Run("invalid endpoint", func(t *testing.T) {
		t.Parallel()
		_, err := New()
		if !geo.IsCode(err, geo.ErrCodeInvalidArgument) {
			t.Fatalf("New() error = %v", err)
		}
		_, err = New(WithEndpoint("ftp://example.com/"))
		if !geo.IsCode(err, geo.ErrCodeInvalidArgument) {
			t.Fatalf("New() error = %v", err)
		}
	})

	t.Run("invalid ip", func(t *testing.T) {
		t.Parallel()
		locator, err := New(WithEndpoint("http://example.com/json/"))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		_, err = locator.Lookup(context.Background(), "bad-ip")
		if !geo.IsInvalidIP(err) {
			t.Fatalf("Lookup() error = %v", err)
		}
	})

	t.Run("not found status", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		locator, err := New(WithEndpoint(server.URL + "/"))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		_, err = locator.Lookup(context.Background(), "1.2.3.4")
		if !geo.IsNotFound(err) {
			t.Fatalf("Lookup() error = %v", err)
		}
	})

	t.Run("fail reserved range", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status":  "fail",
				"message": "reserved range",
			})
		}))
		defer server.Close()

		locator, err := New(WithEndpoint(server.URL + "/"))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		_, err = locator.Lookup(context.Background(), "127.0.0.1")
		if !geo.IsNotFound(err) {
			t.Fatalf("Lookup() error = %v", err)
		}
	})

	t.Run("fail invalid key is unavailable", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status":  "fail",
				"message": "invalid key",
			})
		}))
		defer server.Close()

		locator, err := New(WithEndpoint(server.URL + "/"))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		_, err = locator.Lookup(context.Background(), "1.2.3.4")
		if !geo.IsUnavailable(err) {
			t.Fatalf("Lookup() error = %v, want Unavailable not NotFound", err)
		}
	})

	t.Run("fail invalid ip is not found", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status":  "fail",
				"message": "invalid ip",
			})
		}))
		defer server.Close()

		locator, err := New(WithEndpoint(server.URL + "/"))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		_, err = locator.Lookup(context.Background(), "1.2.3.4")
		if !geo.IsNotFound(err) {
			t.Fatalf("Lookup() error = %v, want NotFound", err)
		}
	})

	t.Run("server error", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		locator, err := New(WithEndpoint(server.URL+"/"), WithTimeout(time.Second))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		_, err = locator.Lookup(context.Background(), "1.2.3.4")
		if !geo.IsUnavailable(err) {
			t.Fatalf("Lookup() error = %v", err)
		}
	})
}

func TestIsNotFoundFailMessage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		msg  string
		want bool
	}{
		{"reserved range", true},
		{"private range", true},
		{"invalid ip", true},
		{"invalid query", true},
		{"invalid key", false},
		{"invalid api key", false},
		{"invalid token", false},
		{"rate limited", false},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			t.Parallel()
			if got := isNotFoundFailMessage(tt.msg); got != tt.want {
				t.Fatalf("isNotFoundFailMessage(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

func TestNestedDataParser(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"country": "中国",
				"city":    "杭州",
			},
		})
	}))
	defer server.Close()

	locator, err := New(WithEndpoint(server.URL + "/"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	loc, err := locator.Lookup(context.Background(), "114.114.114.114")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if loc.Country != "中国" || loc.City != "杭州" {
		t.Fatalf("unexpected location: %+v", loc)
	}
}

func TestBuildRequestURLPreservesPercentEncoding(t *testing.T) {
	t.Parallel()

	// endpoint 自身含 %2F 编码，且带 %s 占位符
	got, err := buildRequestURL("https://example.com/path%2Fv1?ip=%s", "1.2.3.4")
	if err != nil {
		t.Fatalf("buildRequestURL error = %v", err)
	}
	want := "https://example.com/path%2Fv1?ip=1.2.3.4"
	if got != want {
		t.Fatalf("buildRequestURL = %q, want %q", got, want)
	}

	// 若误用 fmt.Sprintf，%2F 会被破坏；这里确保 %2F 仍在
	if !strings.Contains(got, "path%2Fv1") || !strings.Contains(got, "1.2.3.4") {
		t.Fatalf("unexpected url: %s", got)
	}
}

func TestNewZeroTimeoutFallsBackToDefault(t *testing.T) {
	t.Parallel()

	locator, err := New(WithEndpoint("http://example.com/json/"), WithTimeout(0))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if locator.client == nil || locator.client.Timeout != 5*time.Second {
		t.Fatalf("client timeout = %v, want 5s", locator.client.Timeout)
	}
}

func TestLookupCanceledContext(t *testing.T) {
	t.Parallel()

	locator, err := New(WithEndpoint("http://example.com/json/"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = locator.Lookup(ctx, "1.2.3.4")
	if err == nil {
		t.Fatal("expected canceled context error")
	}
}
