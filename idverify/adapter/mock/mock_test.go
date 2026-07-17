package mock

import (
	"context"
	"errors"
	"testing"

	"github.com/f2xme/gox/idverify"
)

func TestMockMatchDefault(t *testing.T) {
	v, err := New()
	if err != nil {
		t.Fatal(err)
	}
	res, err := v.Verify(context.Background(), idverify.Request{Name: "张三", IDNumber: "110101199001011234"})
	if err != nil || !res.Matched {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if v.Provider() != idverify.ProviderMock {
		t.Fatal(v.Provider())
	}
	if len(v.Calls()) != 1 {
		t.Fatalf("calls=%d", len(v.Calls()))
	}
}

func TestMockMismatchAndInvalid(t *testing.T) {
	v, err := New(
		WithMismatchNames("fail-mismatch"),
		WithInvalidIDNames("fail-id"),
	)
	if err != nil {
		t.Fatal(err)
	}

	res, err := v.Verify(context.Background(), idverify.Request{Name: "fail-mismatch", IDNumber: "110101199001011234"})
	if err != nil || res.Matched || res.ErrorCode != idverify.CodeNameMismatch {
		t.Fatalf("%+v %v", res, err)
	}

	res, err = v.Verify(context.Background(), idverify.Request{Name: "fail-id", IDNumber: "110101199001011234"})
	if err != nil || res.Matched || res.ErrorCode != idverify.CodeIDInvalid {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestMockSystemErrorAndInvalidArg(t *testing.T) {
	want := errors.New("boom")
	v := MustNew(WithVerifyError(want))
	_, err := v.Verify(context.Background(), idverify.Request{Name: "a", IDNumber: "1"})
	if !errors.Is(err, want) {
		t.Fatalf("err=%v", err)
	}

	v2 := MustNew()
	_, err = v2.Verify(context.Background(), idverify.Request{})
	if !errors.Is(err, idverify.ErrInvalidArgument) {
		t.Fatalf("err=%v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = v2.Verify(ctx, idverify.Request{Name: "a", IDNumber: "1"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
	if errors.Is(err, idverify.ErrUnavailable) {
		t.Fatal("canceled must not be ErrUnavailable")
	}
}
