package traceid

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/mock"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		incomingID string
		wantID     string
		wantCalls  int
	}{
		{
			name:      "generates missing trace ID",
			wantID:    "generated-trace-id",
			wantCalls: 1,
		},
		{
			name:       "preserves incoming trace ID",
			incomingID: "incoming-trace-id",
			wantID:     "incoming-trace-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generatorCalls := 0
			middleware := New(WithGenerator(func() string {
				generatorCalls++
				return "generated-trace-id"
			}))

			ctx := mock.NewMockContext("GET", "/test")
			ctx.Headers[defaultHeaderKey] = tt.incomingID

			var gotFromHandler string
			handler := middleware(func(ctx httpx.Context) error {
				gotFromHandler = Get(ctx)
				return nil
			})
			if err := handler(ctx); err != nil {
				t.Fatal(err)
			}

			if gotFromHandler != tt.wantID {
				t.Errorf("Get() = %q, want %q", gotFromHandler, tt.wantID)
			}
			if got := ctx.RespHeaders.Get(defaultHeaderKey); got != tt.wantID {
				t.Errorf("response header = %q, want %q", got, tt.wantID)
			}
			if got := ctx.Store[contextKey]; got != tt.wantID {
				t.Errorf("context value = %q, want %q", got, tt.wantID)
			}
			if generatorCalls != tt.wantCalls {
				t.Errorf("generator calls = %d, want %d", generatorCalls, tt.wantCalls)
			}
		})
	}
}

func TestNew_WithHeaderKey(t *testing.T) {
	const customHeader = "Traceparent-ID"
	middleware := New(
		WithHeaderKey(customHeader),
		WithGenerator(func() string { return "custom-trace-id" }),
	)
	ctx := mock.NewMockContext("GET", "/test")

	if err := middleware(func(httpx.Context) error { return nil })(ctx); err != nil {
		t.Fatal(err)
	}

	if got := ctx.RespHeaders.Get(customHeader); got != "custom-trace-id" {
		t.Errorf("response header = %q, want %q", got, "custom-trace-id")
	}
	if got := ctx.RespHeaders.Get(defaultHeaderKey); got != "" {
		t.Errorf("default response header = %q, want empty", got)
	}
}

func TestNew_PropagatesHandlerError(t *testing.T) {
	wantErr := errors.New("handler failed")
	middleware := New(WithGenerator(func() string { return "trace-id" }))
	ctx := mock.NewMockContext("GET", "/test")

	err := middleware(func(httpx.Context) error { return wantErr })(ctx)
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

func TestDefaultGenerator(t *testing.T) {
	first := defaultGenerator()
	second := defaultGenerator()

	if len(first) != 32 {
		t.Errorf("ID length = %d, want 32", len(first))
	}
	if _, err := hex.DecodeString(first); err != nil {
		t.Errorf("ID is not hexadecimal: %v", err)
	}
	if first == second {
		t.Error("generated IDs must be unique")
	}
}

func TestGet_FallsBackToRequestHeader(t *testing.T) {
	ctx := mock.NewMockContext("GET", "/test")
	ctx.Headers[defaultHeaderKey] = "incoming-trace-id"

	if got := Get(ctx); got != "incoming-trace-id" {
		t.Errorf("Get() = %q, want %q", got, "incoming-trace-id")
	}
}
