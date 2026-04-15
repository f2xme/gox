package trace

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestStartSpan(t *testing.T) {
	ctx := context.Background()
	span := StartSpan(ctx, SpanKindService, "TestOperation")

	if span.Name() != "TestOperation" {
		t.Errorf("expected name TestOperation, got %s", span.Name())
	}

	if span.Kind() != SpanKindService {
		t.Errorf("expected kind service, got %s", span.Kind())
	}

	if span.Context() != ctx {
		t.Error("context mismatch")
	}
}

func TestSpanSet(t *testing.T) {
	ctx := context.Background()
	span := StartSpan(ctx, SpanKindDAO, "GetUser")

	span.Set("user_id", 123).Set("name", "test")

	attrs := span.Attrs()
	if attrs["user_id"] != 123 {
		t.Errorf("expected user_id 123, got %v", attrs["user_id"])
	}
	if attrs["name"] != "test" {
		t.Errorf("expected name test, got %v", attrs["name"])
	}
}

func TestSpanEnd(t *testing.T) {
	ctx := context.Background()
	span := StartSpan(ctx, SpanKindService, "TestOp")

	time.Sleep(10 * time.Millisecond)

	result := span.End(nil)

	if !result.Success() {
		t.Error("expected success")
	}

	if result.DurationMs() < 10 {
		t.Errorf("expected duration >= 10ms, got %d", result.DurationMs())
	}
}

func TestSpanEndWithError(t *testing.T) {
	ctx := context.Background()
	span := StartSpan(ctx, SpanKindDAO, "GetUser")

	testErr := errors.New("not found")
	result := span.End(testErr)

	if result.Success() {
		t.Error("expected failure")
	}

	if result.Error != testErr {
		t.Errorf("expected error %v, got %v", testErr, result.Error)
	}
}

func TestServiceHelper(t *testing.T) {
	ctx := context.Background()
	var capturedResult *SpanResult

	SetCallback(func(r *SpanResult) {
		capturedResult = r
	})
	defer SetCallback(nil)

	var err error
	func() {
		defer Service(ctx, "GetUser", "id", 123)(&err)
	}()

	if capturedResult == nil {
		t.Fatal("callback not called")
	}

	if capturedResult.Name() != "GetUser" {
		t.Errorf("expected name GetUser, got %s", capturedResult.Name())
	}

	if capturedResult.Kind() != SpanKindService {
		t.Errorf("expected kind service, got %s", capturedResult.Kind())
	}

	if capturedResult.Attrs()["id"] != 123 {
		t.Errorf("expected id 123, got %v", capturedResult.Attrs()["id"])
	}
}

func TestDAOHelper(t *testing.T) {
	ctx := context.Background()
	var capturedResult *SpanResult

	SetCallback(func(r *SpanResult) {
		capturedResult = r
	})
	defer SetCallback(nil)

	var err error
	func() {
		defer DAO(ctx, "GetUserByID", "user_id", 456)(&err)
	}()

	if capturedResult == nil {
		t.Fatal("callback not called")
	}

	if capturedResult.Kind() != SpanKindDAO {
		t.Errorf("expected kind dao, got %s", capturedResult.Kind())
	}
}

func TestHelperWithError(t *testing.T) {
	ctx := context.Background()
	var capturedResult *SpanResult

	SetCallback(func(r *SpanResult) {
		capturedResult = r
	})
	defer SetCallback(nil)

	testErr := errors.New("db error")
	err := testErr
	func() {
		defer Service(ctx, "CreateUser")(&err)
	}()

	if capturedResult == nil {
		t.Fatal("callback not called")
	}

	if capturedResult.Success() {
		t.Error("expected failure")
	}

	if capturedResult.Error != testErr {
		t.Errorf("expected error %v, got %v", testErr, capturedResult.Error)
	}
}

func TestSpanKinds(t *testing.T) {
	tests := []struct {
		helper func(context.Context, string, ...any) EndFunc
		kind   SpanKind
	}{
		{Service, SpanKindService},
		{DAO, SpanKindDAO},
		{Cache, SpanKindCache},
		{RPC, SpanKindRPC},
		{MQ, SpanKindMQ},
	}

	for _, tt := range tests {
		var capturedResult *SpanResult
		SetCallback(func(r *SpanResult) {
			capturedResult = r
		})

		var err error
		func() {
			defer tt.helper(context.Background(), "Test")(&err)
		}()

		if capturedResult.Kind() != tt.kind {
			t.Errorf("expected kind %s, got %s", tt.kind, capturedResult.Kind())
		}
	}

	SetCallback(nil)
}
