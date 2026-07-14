package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/f2xme/gox/geo"
)

func TestLookup(t *testing.T) {
	t.Parallel()

	locator, err := New(
		WithLocation("1.2.3.4", &geo.Location{
			Country:  "中国",
			Province: "广东省",
			City:     "深圳市",
			ISP:      "电信",
		}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	loc, err := locator.Lookup(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if loc.Country != "中国" || loc.City != "深圳市" {
		t.Fatalf("unexpected location: %+v", loc)
	}
	if loc.IP != "1.2.3.4" {
		t.Fatalf("IP = %q, want 1.2.3.4", loc.IP)
	}

	// 修改返回值不应影响内部存储
	loc.City = "广州市"
	again, err := locator.Lookup(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if again.City != "深圳市" {
		t.Fatalf("internal location mutated: %q", again.City)
	}
}

func TestLookupNotFoundAndInvalid(t *testing.T) {
	t.Parallel()

	locator, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = locator.Lookup(context.Background(), "9.9.9.9")
	if !geo.IsNotFound(err) {
		t.Fatalf("Lookup missing ip error = %v, want NotFound", err)
	}

	_, err = locator.Lookup(context.Background(), "bad")
	if !geo.IsInvalidIP(err) {
		t.Fatalf("Lookup invalid ip error = %v, want InvalidIP", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = locator.Lookup(ctx, "1.1.1.1")
	if err == nil {
		t.Fatal("Lookup with canceled context should fail")
	}
}

func TestSetDeleteResetAndLookupError(t *testing.T) {
	t.Parallel()

	locator, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := locator.Set("8.8.8.8", &geo.Location{Country: "美国"}); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if locator.Count() != 1 {
		t.Fatalf("Count() = %d, want 1", locator.Count())
	}

	loc, err := locator.Lookup(context.Background(), "8.8.8.8")
	if err != nil || loc.Country != "美国" {
		t.Fatalf("Lookup after Set = %+v, err=%v", loc, err)
	}

	if err := locator.Delete("8.8.8.8"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if locator.Count() != 0 {
		t.Fatalf("Count after Delete = %d", locator.Count())
	}

	_ = locator.Set("1.1.1.1", &geo.Location{Country: "澳大利亚"})
	locator.Reset()
	if locator.Count() != 0 {
		t.Fatalf("Count after Reset = %d", locator.Count())
	}

	locator.SetLookupError(errors.New("boom"))
	_, err = locator.Lookup(context.Background(), "1.1.1.1")
	if err == nil || err.Error() != "boom" {
		t.Fatalf("LookupError = %v, want boom", err)
	}
}

func TestWithLocationsAndMustNew(t *testing.T) {
	t.Parallel()

	locator, err := New(WithLocations(map[string]*geo.Location{
		"  1.1.1.1  ": {Country: "澳大利亚", City: "悉尼"},
	}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	loc, err := locator.Lookup(context.Background(), "1.1.1.1")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if loc.City != "悉尼" {
		t.Fatalf("City = %q", loc.City)
	}

	// MustNew 在有效参数下应成功
	_ = MustNew(WithLocation("127.0.0.1", &geo.Location{Country: "内网"}))
}

func TestNewInvalidIP(t *testing.T) {
	t.Parallel()

	_, err := New(WithLocation("not-ip", &geo.Location{Country: "x"}))
	if !geo.IsInvalidIP(err) {
		t.Fatalf("New invalid ip error = %v", err)
	}
}
