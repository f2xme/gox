package logx

import (
	"errors"
	"testing"
)

type recordedMsg struct {
	level string
	msg   string
	err   error
	metas []Meta
}

type mockLogger struct {
	records []recordedMsg
}

func (m *mockLogger) Info(msg string, metas ...Meta) {
	m.records = append(m.records, recordedMsg{level: "info", msg: msg, metas: metas})
}
func (m *mockLogger) Warn(msg string, metas ...Meta) {
	m.records = append(m.records, recordedMsg{level: "warn", msg: msg, metas: metas})
}
func (m *mockLogger) Error(err error, metas ...Meta) {
	m.records = append(m.records, recordedMsg{level: "error", err: err, metas: metas})
}
func (m *mockLogger) Fatal(err error, metas ...Meta) {
	m.records = append(m.records, recordedMsg{level: "fatal", err: err, metas: metas})
}

type mockFlushStopSync struct {
	mockLogger
	flushed bool
	stopped bool
	synced  bool
}

func (m *mockFlushStopSync) Flush() error { m.flushed = true; return nil }
func (m *mockFlushStopSync) Stop() error  { m.stopped = true; return nil }
func (m *mockFlushStopSync) Sync() error  { m.synced = true; return nil }

func TestInit_And_Info(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	Info("hello", NewKV("k", "v"))
	if len(ml.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(ml.records))
	}
	if ml.records[0].msg != "hello" {
		t.Errorf("expected msg 'hello', got %q", ml.records[0].msg)
	}
}

func TestWarn(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	Warn("warning")
	if ml.records[0].level != "warn" {
		t.Error("expected warn level")
	}
}

func TestError_NilErr(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	Error(nil)
	if len(ml.records) != 0 {
		t.Error("expected no record for nil error")
	}
}

func TestError_WithoutCaller(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	Error(errors.New("oops"))
	if len(ml.records) != 1 {
		t.Fatal("expected 1 record")
	}
	metas := ml.records[0].metas
	if len(metas) != 0 {
		t.Fatalf("expected 0 metas when no metas provided, got %d", len(metas))
	}
}

func TestError_WithMetas(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	Error(errors.New("oops"), NewKV("action", "test"))
	metas := ml.records[0].metas
	if len(metas) != 1 || metas[0].Key() != "action" {
		t.Error("expected only user-provided metas")
	}
}

func TestFlush_Stop_Sync(t *testing.T) {
	ml := &mockFlushStopSync{}
	Init(ml)
	defer Init(nopLogger{})

	Flush()
	if !ml.flushed {
		t.Error("expected Flush called")
	}
	Stop()
	if !ml.stopped {
		t.Error("expected Stop called")
	}
	Sync()
	if !ml.synced {
		t.Error("expected Sync called")
	}
}

func TestFlush_NoOp_WhenNotImplemented(t *testing.T) {
	ml := &mockLogger{}
	Init(ml)
	defer Init(nopLogger{})

	if err := Flush(); err != nil {
		t.Error("expected nil error for non-Flusher")
	}
	if err := Stop(); err != nil {
		t.Error("expected nil error for non-Stopper")
	}
	if err := Sync(); err != nil {
		t.Error("expected nil error for non-Syncer")
	}
}

func TestDefaultLogger_NoOp(t *testing.T) {
	Init(nopLogger{})
	Info("test")
	Warn("test")
	Error(errors.New("test"))
}
