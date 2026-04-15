package graceful

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPServer(t *testing.T) {
	server := &http.Server{
		Addr: ":0",
	}

	// Start server in background
	go server.ListenAndServe()
	time.Sleep(10 * time.Millisecond)

	closer := HTTPServer(server)
	err := closer.Close(context.Background())

	// Shutdown may return ErrServerClosed which is expected
	if err != nil && err != http.ErrServerClosed {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHTTPServerWithHandler(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Create a new server with the test server's listener
	httpServer := &http.Server{
		Addr:    server.Listener.Addr().String(),
		Handler: handler,
	}

	closer := HTTPServer(httpServer)
	err := closer.Close(context.Background())

	if err != nil && err != http.ErrServerClosed {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDBCloser(t *testing.T) {
	// We can't easily test with a real DB, so we'll skip this
	// In a real scenario, you'd use a test database
	t.Skip("requires database connection")
}

func TestIOCloser(t *testing.T) {
	mock := &mockCloser{}
	ioCloser := io.Closer(&closerAdapter{mock})

	closer := IOCloser(ioCloser)
	err := closer.Close(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !mock.closed {
		t.Error("closer was not called")
	}
}

type closerAdapter struct {
	mock *mockCloser
}

type mockCloser struct {
	closed bool
	err    error
}

func (c *closerAdapter) Close() error {
	c.mock.closed = true
	return c.mock.err
}

func TestIOCloserWithError(t *testing.T) {
	expectedErr := errors.New("close error")
	mock := &mockCloser{err: expectedErr}
	ioCloser := io.Closer(&closerAdapter{mock})

	closer := IOCloser(ioCloser)
	err := closer.Close(context.Background())

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestGenericCloser(t *testing.T) {
	called := false
	fn := func() error {
		called = true
		return nil
	}

	closer := GenericCloser(fn)
	err := closer.Close(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !called {
		t.Error("function was not called")
	}
}

func TestGenericCloserWithError(t *testing.T) {
	expectedErr := errors.New("close error")
	fn := func() error {
		return expectedErr
	}

	closer := GenericCloser(fn)
	err := closer.Close(context.Background())

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
