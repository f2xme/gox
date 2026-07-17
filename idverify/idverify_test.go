package idverify

import (
	"errors"
	"testing"
)

func TestRequestNormalizeAndValid(t *testing.T) {
	req := Request{Name: " 张三 ", IDNumber: " 11010119900101123x "}
	n := req.Normalize()
	if n.Name != "张三" || n.IDNumber != "11010119900101123X" {
		t.Fatalf("normalize = %+v", n)
	}
	if !n.Valid() {
		t.Fatal("want valid")
	}
	if (Request{}).Valid() {
		t.Fatal("empty should be invalid")
	}
}

func TestErrorWrapIs(t *testing.T) {
	err := Wrap(ProviderBaidu, "verify", ErrNotConfigured)
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("Is failed: %v", err)
	}
	var e *Error
	if !errors.As(err, &e) || e.Provider != ProviderBaidu || e.Op != "verify" {
		t.Fatalf("As failed: %+v", e)
	}
}
