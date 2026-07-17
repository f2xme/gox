package aliyun

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/f2xme/gox/idverify"
)

func TestAliyunVerifySuccess(t *testing.T) {
	var calls int32
	v, err := New(WithAccessKeyID("ak"), WithAccessKeySecret("sk"), WithEndpoints("ep-a", "ep-b"))
	if err != nil {
		t.Fatal(err)
	}
	v.WithCaller(func(_ context.Context, endpoint, name, id string) (callResult, error) {
		atomic.AddInt32(&calls, 1)
		if endpoint != "ep-a" || name != "张三" {
			t.Fatalf("endpoint=%s name=%s", endpoint, name)
		}
		return callResult{HTTPStatus: 200, Code: "200", BizCode: "1"}, nil
	})

	res, err := v.Verify(context.Background(), idverify.Request{Name: "张三", IDNumber: "110101199001011234"})
	if err != nil || !res.Matched {
		t.Fatalf("%+v %v", res, err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestAliyunBizCodes(t *testing.T) {
	tests := []struct {
		biz  string
		code string
	}{
		{"2", idverify.CodeNameMismatch},
		{"3", idverify.CodeIDInvalid},
	}
	for _, tt := range tests {
		v := MustNew(WithAccessKeyID("ak"), WithAccessKeySecret("sk"))
		v.WithCaller(func(context.Context, string, string, string) (callResult, error) {
			return callResult{HTTPStatus: 200, Code: "200", BizCode: tt.biz}, nil
		})
		res, err := v.Verify(context.Background(), idverify.Request{Name: "a", IDNumber: "1"})
		if err != nil || res.Matched || res.ErrorCode != tt.code {
			t.Fatalf("biz=%s res=%+v err=%v", tt.biz, res, err)
		}
	}
}

func TestAliyunFailoverAndClientFault(t *testing.T) {
	var calls []string
	v := MustNew(WithAccessKeyID("ak"), WithAccessKeySecret("sk"), WithEndpoints("ep-a", "ep-b"))
	v.WithCaller(func(_ context.Context, endpoint, _, _ string) (callResult, error) {
		calls = append(calls, endpoint)
		if endpoint == "ep-a" {
			return callResult{}, fmt.Errorf("network")
		}
		return callResult{HTTPStatus: 200, Code: "200", BizCode: "1"}, nil
	})
	res, err := v.Verify(context.Background(), idverify.Request{Name: "a", IDNumber: "1"})
	if err != nil || !res.Matched || len(calls) != 2 {
		t.Fatalf("res=%+v err=%v calls=%v", res, err, calls)
	}

	calls = nil
	var n int32
	v.WithCaller(func(context.Context, string, string, string) (callResult, error) {
		atomic.AddInt32(&n, 1)
		return callResult{HTTPStatus: 200, Code: "410", Message: "not open"}, nil
	})
	_, err = v.Verify(context.Background(), idverify.Request{Name: "a", IDNumber: "1"})
	if err == nil || atomic.LoadInt32(&n) != 1 {
		t.Fatalf("err=%v n=%d", err, n)
	}
}

func TestAliyunHTTPStatusAndNotConfigured(t *testing.T) {
	v := MustNew(WithAccessKeyID("ak"), WithAccessKeySecret("sk"), WithEndpoints("ep-a", "ep-b"))
	var calls int
	v.WithCaller(func(_ context.Context, endpoint, _, _ string) (callResult, error) {
		calls++
		if endpoint == "ep-a" {
			return callResult{HTTPStatus: 502, Code: "200", BizCode: "1"}, nil
		}
		return callResult{HTTPStatus: 200, Code: "200", BizCode: "1"}, nil
	})
	res, err := v.Verify(context.Background(), idverify.Request{Name: "a", IDNumber: "1"})
	if err != nil || !res.Matched || calls != 2 {
		t.Fatalf("res=%+v err=%v calls=%d", res, err, calls)
	}

	_, err = New(WithAccessKeyID(""), WithAccessKeySecret(""))
	if !errors.Is(err, idverify.ErrNotConfigured) {
		t.Fatalf("err=%v", err)
	}
}

func TestAliyunDefaultEndpoints(t *testing.T) {
	v := MustNew(WithAccessKeyID("ak"), WithAccessKeySecret("sk"))
	if len(v.options.Endpoints) != 2 {
		t.Fatalf("%v", v.options.Endpoints)
	}
}

func TestAliyunClientCache(t *testing.T) {
	v := MustNew(WithAccessKeyID("ak"), WithAccessKeySecret("sk"), WithEndpoints("ep-a"))
	c1, err := v.clientFor("ep-a")
	if err != nil {
		t.Fatal(err)
	}
	c2, err := v.clientFor("ep-a")
	if err != nil {
		t.Fatal(err)
	}
	if c1 == nil || c1 != c2 {
		t.Fatal("want cached client")
	}
}
