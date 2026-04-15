package logx

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCtx_Info(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	Ctx(context.Background()).Info("hello", NewKV("k", "v"))
	if len(ml.records) != 1 {
		t.Fatal("expected 1 record")
	}
	if ml.records[0].msg != "hello" {
		t.Errorf("expected msg 'hello', got %q", ml.records[0].msg)
	}
	hasCaller := false
	hasK := false
	for _, m := range ml.records[0].metas {
		if m.Key() == "caller" {
			hasCaller = true
		}
		if m.Key() == "k" {
			hasK = true
		}
	}
	if !hasCaller {
		t.Error("expected caller meta")
	}
	if !hasK {
		t.Error("expected user meta 'k'")
	}
}

func TestCtx_NilContext(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	Ctx(context.Background()).Info("test")
	if len(ml.records) != 1 {
		t.Error("expected 1 record")
	}
}

func TestCtx_With(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	cl := Ctx(context.Background()).With(NewKV("service", "api"))
	cl.Info("msg")

	found := false
	for _, m := range ml.records[0].metas {
		if m.Key() == "service" && m.Value() == "api" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'service' meta from With()")
	}
}

func TestCtx_With_Immutable(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	cl1 := Ctx(context.Background())
	cl2 := cl1.With(NewKV("extra", "field"))

	cl1.Info("msg1")
	cl2.Info("msg2")

	for _, m := range ml.records[0].metas {
		if m.Key() == "extra" {
			t.Error("cl1 should not have 'extra' meta (immutability violation)")
		}
	}
}

func TestCtx_WithCaller(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	Ctx(context.Background()).WithCaller("custom/path.go:42").Info("msg")

	found := false
	for _, m := range ml.records[0].metas {
		if m.Key() == "caller" && m.Value() == "custom/path.go:42" {
			found = true
		}
	}
	if !found {
		t.Error("expected overridden caller")
	}
}

func TestCtx_Error_NilErr(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	Ctx(context.Background()).Error(nil)
	if len(ml.records) != 0 {
		t.Error("expected no record for nil error")
	}
}

func TestSetContextExtractor(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})
	defer SetContextExtractor(nil)

	type traceKey struct{}
	SetContextExtractor(func(ctx context.Context) []Meta {
		if tid, ok := ctx.Value(traceKey{}).(string); ok {
			return []Meta{NewKV("trace_id", tid)}
		}
		return nil
	})

	ctx := context.WithValue(context.Background(), traceKey{}, "abc-123")
	Ctx(ctx).Info("request")

	found := false
	for _, m := range ml.records[0].metas {
		if m.Key() == "trace_id" && m.Value() == "abc-123" {
			found = true
		}
	}
	if !found {
		t.Error("expected trace_id from extractor")
	}
}

func TestInfoCtx_Shortcut(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	InfoCtx(context.Background(), "hello")
	if len(ml.records) != 1 {
		t.Fatal("expected 1 record")
	}
	hasCaller := false
	for _, m := range ml.records[0].metas {
		if m.Key() == "caller" && strings.Contains(m.Value().(string), "context_test.go") {
			hasCaller = true
		}
	}
	if !hasCaller {
		t.Error("expected caller pointing to context_test.go")
	}
}

func TestErrorCtx_Shortcut(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	ErrorCtx(context.Background(), errors.New("fail"), NewKV("action", "test"))
	if len(ml.records) != 1 {
		t.Fatal("expected 1 record")
	}
	if ml.records[0].err.Error() != "fail" {
		t.Error("expected error 'fail'")
	}
}
