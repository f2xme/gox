package graceful

import (
	"context"
	"errors"
	"testing"
)

func TestCloserFunc(t *testing.T) {
	called := false
	closer := CloserFunc(func(ctx context.Context) error {
		called = true
		return nil
	})

	err := closer.Close(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("closer function was not called")
	}
}

func TestCloserFuncWithError(t *testing.T) {
	expectedErr := errors.New("close error")
	closer := CloserFunc(func(ctx context.Context) error {
		return expectedErr
	})

	err := closer.Close(context.Background())
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestCloserFuncWithContext(t *testing.T) {
	var receivedCtx context.Context
	closer := CloserFunc(func(ctx context.Context) error {
		receivedCtx = ctx
		return nil
	})

	ctx := context.WithValue(context.Background(), "key", "value")
	closer.Close(ctx)

	if receivedCtx != ctx {
		t.Error("context was not passed correctly")
	}
}
