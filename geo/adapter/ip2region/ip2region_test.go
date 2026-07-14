package ip2region

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/f2xme/gox/geo"
)

type fakeSearcher struct {
	region string
	err    error
	closed bool
}

func (f *fakeSearcher) Search(ip any) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.region, nil
}

func (f *fakeSearcher) Close() {
	f.closed = true
}

func TestParseRegion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want geo.Location
	}{
		{
			name: "new format",
			in:   "中国|广东省|深圳市|电信|CN",
			want: geo.Location{Country: "中国", Province: "广东省", City: "深圳市", ISP: "电信", CountryCode: "CN"},
		},
		{
			name: "old format",
			in:   "中国|0|广东省|深圳市|电信",
			want: geo.Location{Country: "中国", Province: "广东省", City: "深圳市", ISP: "电信"},
		},
		{
			name: "english new format",
			in:   "Australia|Queensland|Brisbane|0|AU",
			want: geo.Location{Country: "Australia", Province: "Queensland", City: "Brisbane", CountryCode: "AU"},
		},
		{
			name: "four fields",
			in:   "中国|广东省|深圳市|电信",
			want: geo.Location{Country: "中国", Province: "广东省", City: "深圳市", ISP: "电信"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseRegion(tt.in)
			if got.Country != tt.want.Country ||
				got.Province != tt.want.Province ||
				got.City != tt.want.City ||
				got.ISP != tt.want.ISP ||
				got.CountryCode != tt.want.CountryCode {
				t.Fatalf("parseRegion(%q) = %+v, want %+v", tt.in, got, tt.want)
			}
		})
	}
}

func TestLookupWithFakeSearcher(t *testing.T) {
	t.Parallel()

	locator := &Locator{
		service: &fakeSearcher{region: "中国|0|浙江省|杭州市|移动"},
	}

	loc, err := locator.Lookup(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if loc.Province != "浙江省" || loc.City != "杭州市" || loc.ISP != "移动" {
		t.Fatalf("unexpected location: %+v", loc)
	}

	locator.service = &fakeSearcher{region: ""}
	_, err = locator.Lookup(context.Background(), "1.2.3.4")
	if !geo.IsNotFound(err) {
		t.Fatalf("empty region error = %v", err)
	}

	locator.service = &fakeSearcher{region: "内网IP|0|内网IP|内网IP|内网IP"}
	_, err = locator.Lookup(context.Background(), "127.0.0.1")
	if !geo.IsNotFound(err) {
		t.Fatalf("private region error = %v, want NotFound", err)
	}

	locator.service = &fakeSearcher{err: errors.New("boom")}
	_, err = locator.Lookup(context.Background(), "1.2.3.4")
	if !geo.IsCode(err, geo.ErrCodeInternal) {
		t.Fatalf("search error = %v", err)
	}

	_, err = locator.Lookup(context.Background(), "bad")
	if !geo.IsInvalidIP(err) {
		t.Fatalf("invalid ip error = %v", err)
	}
}

func TestClose(t *testing.T) {
	t.Parallel()

	fake := &fakeSearcher{}
	locator := &Locator{service: fake}
	if err := locator.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !fake.closed {
		t.Fatal("expected searcher closed")
	}
	if err := (*Locator)(nil).Close(); err != nil {
		t.Fatalf("nil Close() error = %v", err)
	}
}

func TestValidateOptions(t *testing.T) {
	t.Parallel()

	_, err := New()
	if !geo.IsCode(err, geo.ErrCodeInvalidArgument) {
		t.Fatalf("New() without db path error = %v", err)
	}

	_, err = New(WithV4DBPath("x.xdb"), WithPoolSize(0))
	if !geo.IsCode(err, geo.ErrCodeInvalidArgument) {
		t.Fatalf("New() invalid pool error = %v", err)
	}

	_, err = New(WithV4DBPath("x.xdb"), WithCachePolicy("unknown"))
	if !geo.IsCode(err, geo.ErrCodeInvalidArgument) {
		t.Fatalf("New() invalid policy error = %v", err)
	}
}

func TestIntegrationLookup(t *testing.T) {
	dbPath := os.Getenv("GEO_IP2REGION_V4_DB")
	if dbPath == "" {
		t.Skip("set GEO_IP2REGION_V4_DB to run integration test")
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Skipf("db path not accessible: %v", err)
	}

	locator, err := New(WithV4DBPath(dbPath), WithCachePolicy(CachePolicyBuffer))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer locator.Close()

	loc, err := locator.Lookup(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if loc.Empty() {
		t.Fatalf("expected non-empty location, got %+v", loc)
	}
	t.Logf("8.8.8.8 => %s", loc.String())
}
