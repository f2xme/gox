package httpx

import (
	"testing"
)

func TestData(t *testing.T) {
	ctx := &errorHandlerContext{}

	if err := Data(ctx, map[string]string{"name": "alice"}); err != nil {
		t.Fatal(err)
	}

	if ctx.status != 200 {
		t.Fatalf("status: want 200, got %d", ctx.status)
	}
	resp, ok := ctx.body.(Response)
	if !ok {
		t.Fatalf("body type: want Response, got %T", ctx.body)
	}
	if !resp.Success {
		t.Fatal("expected success=true")
	}
	if resp.Message != "ok" {
		t.Fatalf("message: want %q, got %q", "ok", resp.Message)
	}
	if resp.Data == nil {
		t.Fatal("expected data to be set")
	}
}

func TestDataCustomMessage(t *testing.T) {
	ctx := &errorHandlerContext{}

	if err := Data(ctx, map[string]string{"id": "1"}, "创建成功"); err != nil {
		t.Fatal(err)
	}

	resp := ctx.body.(Response)
	if resp.Message != "创建成功" {
		t.Fatalf("message: want %q, got %q", "创建成功", resp.Message)
	}
}

func TestDone(t *testing.T) {
	ctx := &errorHandlerContext{}

	if err := Done(ctx, "删除成功"); err != nil {
		t.Fatal(err)
	}

	resp, ok := ctx.body.(Response)
	if !ok {
		t.Fatalf("body type: want Response, got %T", ctx.body)
	}
	if !resp.Success {
		t.Fatal("expected success=true")
	}
	if resp.Message != "删除成功" {
		t.Fatalf("message: want %q, got %q", "删除成功", resp.Message)
	}
	if resp.Data != nil {
		t.Fatalf("data: want nil, got %v", resp.Data)
	}
}

func TestFail(t *testing.T) {
	ctx := &errorHandlerContext{}

	if err := Fail(ctx, "something went wrong"); err != nil {
		t.Fatal(err)
	}

	resp, ok := ctx.body.(Response)
	if !ok {
		t.Fatalf("body type: want Response, got %T", ctx.body)
	}
	if resp.Success {
		t.Fatal("expected success=false")
	}
	if resp.Message != "something went wrong" {
		t.Fatalf("message: want %q, got %q", "something went wrong", resp.Message)
	}
}

func TestFailWithData(t *testing.T) {
	ctx := &errorHandlerContext{}
	fieldErrors := map[string]string{"name": "必填"}

	if err := Fail(ctx, "表单校验失败", fieldErrors); err != nil {
		t.Fatal(err)
	}

	resp := ctx.body.(Response)
	if resp.Success {
		t.Fatal("expected success=false")
	}
	if resp.Data == nil {
		t.Fatal("expected data to be set")
	}
}
