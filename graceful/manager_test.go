package graceful

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestManagerRegister(t *testing.T) {
	m := New().(*manager)

	closer := CloserFunc(func(ctx context.Context) error {
		return nil
	})

	m.Register("test", closer)

	if len(m.resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(m.resources))
	}

	if m.resources[0].name != "test" {
		t.Errorf("expected name 'test', got %q", m.resources[0].name)
	}
}

func TestManagerShutdown(t *testing.T) {
	m := New()

	var closed int32
	closer := CloserFunc(func(ctx context.Context) error {
		atomic.AddInt32(&closed, 1)
		return nil
	})

	m.Register("test", closer)

	err := m.Shutdown(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&closed) != 1 {
		t.Error("closer was not called")
	}
}

func TestManagerShutdownPriority(t *testing.T) {
	m := New()

	var order []string
	createCloser := func(name string) Closer {
		return CloserFunc(func(ctx context.Context) error {
			order = append(order, name)
			return nil
		})
	}

	m.Register("low", createCloser("low"), WithPriority(1))
	m.Register("high", createCloser("high"), WithPriority(10))
	m.Register("medium", createCloser("medium"), WithPriority(5))

	err := m.Shutdown(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := []string{"high", "medium", "low"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d closers, got %d", len(expected), len(order))
	}

	for i, name := range expected {
		if order[i] != name {
			t.Errorf("index %d: expected %q, got %q", i, name, order[i])
		}
	}
}

func TestManagerShutdownTimeout(t *testing.T) {
	m := New(WithTimeout(100 * time.Millisecond))

	closer := CloserFunc(func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	m.Register("slow", closer)

	start := time.Now()
	err := m.Shutdown(context.Background())
	duration := time.Since(start)

	if err == nil {
		t.Error("expected timeout error")
	}

	if duration > 150*time.Millisecond {
		t.Errorf("shutdown took too long: %v", duration)
	}
}

func TestManagerShutdownError(t *testing.T) {
	m := New()

	expectedErr := errors.New("close error")
	closer := CloserFunc(func(ctx context.Context) error {
		return expectedErr
	})

	m.Register("failing", closer)

	err := m.Shutdown(context.Background())
	if err == nil {
		t.Error("expected error")
	}
}

func TestManagerMultipleResources(t *testing.T) {
	m := New()

	var closed int32
	for i := 0; i < 5; i++ {
		closer := CloserFunc(func(ctx context.Context) error {
			atomic.AddInt32(&closed, 1)
			return nil
		})
		m.Register("test", closer)
	}

	err := m.Shutdown(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&closed) != 5 {
		t.Errorf("expected 5 closers called, got %d", closed)
	}
}

func TestWithPriority(t *testing.T) {
	m := New().(*manager)

	closer := CloserFunc(func(ctx context.Context) error {
		return nil
	})

	m.Register("test", closer, WithPriority(100))

	if m.resources[0].priority != 100 {
		t.Errorf("expected priority 100, got %d", m.resources[0].priority)
	}
}

func TestWithResourceTimeout(t *testing.T) {
	m := New().(*manager)

	closer := CloserFunc(func(ctx context.Context) error {
		return nil
	})

	m.Register("test", closer, WithResourceTimeout(5*time.Second))

	if m.resources[0].timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %v", m.resources[0].timeout)
	}
}

func TestHooks(t *testing.T) {
	var beforeCalled, afterCalled bool
	var timedOutResource string

	m := New(
		OnBeforeShutdown(func() {
			beforeCalled = true
		}),
		OnAfterShutdown(func() {
			afterCalled = true
		}),
		OnTimeout(func(name string) {
			timedOutResource = name
		}),
		WithTimeout(50*time.Millisecond),
	)

	// Add a normal closer
	m.Register("normal", CloserFunc(func(ctx context.Context) error {
		return nil
	}))

	// Add a slow closer that will timeout
	m.Register("slow", CloserFunc(func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}))

	m.Shutdown(context.Background())

	if !beforeCalled {
		t.Error("before shutdown hook was not called")
	}

	if !afterCalled {
		t.Error("after shutdown hook was not called")
	}

	if timedOutResource != "slow" {
		t.Errorf("expected timeout for 'slow', got %q", timedOutResource)
	}
}
