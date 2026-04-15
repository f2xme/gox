package graceful_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/f2xme/gox/graceful"
)

// silentLogger suppresses log output for examples
type silentLogger struct{}

func (l *silentLogger) Printf(format string, v ...interface{}) {}

func ExampleNew() {
	m := graceful.New(graceful.WithLogger(&silentLogger{}))

	// Register resources
	m.Register("cache", graceful.CloserFunc(func(ctx context.Context) error {
		fmt.Println("closing cache")
		return nil
	}))

	// Shutdown immediately for example
	m.Shutdown(context.Background())

	// Output:
	// closing cache
}

func ExampleHTTPServer() {
	server := &http.Server{
		Addr:    ":8080",
		Handler: http.DefaultServeMux,
	}

	m := graceful.New()
	m.Register("http-server", graceful.HTTPServer(server), graceful.WithPriority(10))

	// In real application, call m.Wait() to block until signal
	// m.Wait()
}

func ExampleWithPriority() {
	m := graceful.New(graceful.WithLogger(&silentLogger{}))

	// Higher priority resources are closed first
	m.Register("database", graceful.CloserFunc(func(ctx context.Context) error {
		fmt.Println("closing database")
		return nil
	}), graceful.WithPriority(1))

	m.Register("http-server", graceful.CloserFunc(func(ctx context.Context) error {
		fmt.Println("closing http server")
		return nil
	}), graceful.WithPriority(10))

	m.Shutdown(context.Background())

	// Output:
	// closing http server
	// closing database
}

func ExampleWithTimeout() {
	m := graceful.New(
		graceful.WithTimeout(5 * time.Second),
	)

	m.Register("service", graceful.CloserFunc(func(ctx context.Context) error {
		// This will timeout after 5 seconds
		return nil
	}))

	m.Shutdown(context.Background())
}

func ExampleOnBeforeShutdown() {
	m := graceful.New(
		graceful.WithLogger(&silentLogger{}),
		graceful.OnBeforeShutdown(func() {
			fmt.Println("starting shutdown")
		}),
		graceful.OnAfterShutdown(func() {
			fmt.Println("shutdown complete")
		}),
	)

	m.Register("service", graceful.CloserFunc(func(ctx context.Context) error {
		fmt.Println("closing service")
		return nil
	}))

	m.Shutdown(context.Background())

	// Output:
	// starting shutdown
	// closing service
	// shutdown complete
}

func ExampleGenericCloser() {
	m := graceful.New(graceful.WithLogger(&silentLogger{}))

	// Wrap a simple function
	m.Register("cleanup", graceful.GenericCloser(func() error {
		fmt.Println("cleanup complete")
		return nil
	}))

	m.Shutdown(context.Background())

	// Output:
	// cleanup complete
}
