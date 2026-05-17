package logx

import (
	"errors"
	"sync"
	"testing"
	"time"
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

type blockingLogger struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once

	mu      sync.Mutex
	records []recordedMsg
}

func newBlockingLogger() *blockingLogger {
	return &blockingLogger{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (l *blockingLogger) Info(msg string, metas ...Meta) {
	l.once.Do(func() { close(l.started) })
	<-l.release

	l.mu.Lock()
	defer l.mu.Unlock()
	l.records = append(l.records, recordedMsg{level: "info", msg: msg, metas: metas})
}

func (l *blockingLogger) Warn(msg string, metas ...Meta) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.records = append(l.records, recordedMsg{level: "warn", msg: msg, metas: metas})
}

func (l *blockingLogger) Error(err error, metas ...Meta) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.records = append(l.records, recordedMsg{level: "error", err: err, metas: metas})
}

func (l *blockingLogger) Fatal(err error, metas ...Meta) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.records = append(l.records, recordedMsg{level: "fatal", err: err, metas: metas})
}

func (l *blockingLogger) snapshot() []recordedMsg {
	l.mu.Lock()
	defer l.mu.Unlock()

	records := make([]recordedMsg, len(l.records))
	copy(records, l.records)
	return records
}

func TestInit_WithAsync_InfoDoesNotBlockOnUnderlyingLogger(t *testing.T) {
	bl := newBlockingLogger()
	Init(bl, WithAsync(), WithAsyncBufferSize(1))
	defer Init(nopLogger{})

	done := make(chan struct{})
	go func() {
		Info("async")
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected Info to return before underlying logger finishes")
	}

	select {
	case <-bl.started:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected async worker to call underlying logger")
	}

	close(bl.release)
	if err := Flush(); err != nil {
		t.Fatalf("Flush returned error: %v", err)
	}
	if records := bl.snapshot(); len(records) != 1 || records[0].msg != "async" {
		t.Fatalf("expected async record, got %#v", records)
	}
}

func TestAsyncLogger_FlushWaitsForQueuedRecords(t *testing.T) {
	ml := &mockLogger{}
	Init(ml, WithAsync())
	defer Init(nopLogger{})

	Info("one")
	Warn("two")
	Error(errors.New("three"))

	if err := Flush(); err != nil {
		t.Fatalf("Flush returned error: %v", err)
	}
	if len(ml.records) != 3 {
		t.Fatalf("expected 3 records after Flush, got %d", len(ml.records))
	}
}

func TestAsyncLogger_StopWaitsForQueuedRecords(t *testing.T) {
	ml := &mockFlushStopSync{}
	Init(ml, WithAsync())
	defer Init(nopLogger{})

	Info("one")
	if err := Stop(); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if len(ml.records) != 1 {
		t.Fatalf("expected 1 record after Stop, got %d", len(ml.records))
	}
	if !ml.stopped {
		t.Fatal("expected underlying Stop to be called")
	}
}

func TestAsyncLogger_CopiesMetaSlice(t *testing.T) {
	ml := &mockLogger{}
	Init(ml, WithAsync())
	defer Init(nopLogger{})

	metas := []Meta{NewKV("key", "before")}
	Info("copy", metas...)
	metas[0] = NewKV("key", "after")

	if err := Flush(); err != nil {
		t.Fatalf("Flush returned error: %v", err)
	}
	if got := ml.records[0].metas[0].Value(); got != "before" {
		t.Fatalf("expected copied meta value before mutation, got %v", got)
	}
}
