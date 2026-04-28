package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/f2xme/gox/session"
)

func TestStoreSetGetDelete(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	sess := &session.Session{
		ID:        "sid",
		Values:    map[string]any{"name": "Alice"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Minute),
	}

	if err := store.Set(ctx, sess, time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := store.Get(ctx, "sid")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Values["name"] != "Alice" {
		t.Fatalf("Get() name = %v, want Alice", got.Values["name"])
	}

	got.Values["name"] = "Bob"
	gotAgain, err := store.Get(ctx, "sid")
	if err != nil {
		t.Fatalf("Get() again error = %v", err)
	}
	if gotAgain.Values["name"] != "Alice" {
		t.Fatalf("stored value mutated through returned session")
	}

	if err := store.Delete(ctx, "sid"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := store.Get(ctx, "sid"); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("Get() deleted error = %v, want ErrNotFound", err)
	}
}

func TestStoreExpiration(t *testing.T) {
	store, err := New(WithCleanupInterval(10 * time.Millisecond))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	sess := &session.Session{ID: "sid", Values: map[string]any{}}
	if err := store.Set(ctx, sess, 10*time.Millisecond); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	time.Sleep(25 * time.Millisecond)
	if _, err := store.Get(ctx, "sid"); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("Get() expired error = %v, want ErrNotFound", err)
	}
}

func TestStoreSetPreservesExpiresAt(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	expiresAt := time.Now().Add(10 * time.Minute).Truncate(time.Microsecond)
	sess := &session.Session{
		ID:        "sid",
		Values:    map[string]any{},
		ExpiresAt: expiresAt,
	}
	if err := store.Set(ctx, sess, time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := store.Get(ctx, "sid")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !got.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("ExpiresAt = %v, want %v", got.ExpiresAt, expiresAt)
	}
}

func TestStoreConcurrentAccess(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("sid-%d", i)
			sess := &session.Session{ID: id, Values: map[string]any{"i": i}}
			if err := store.Set(ctx, sess, time.Minute); err != nil {
				t.Errorf("Set() error = %v", err)
				return
			}
			if _, err := store.Get(ctx, id); err != nil {
				t.Errorf("Get() error = %v", err)
			}
		}(i)
	}
	wg.Wait()
}
