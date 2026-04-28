package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/f2xme/gox/session"
	"github.com/f2xme/gox/session/adapter/memory"
)

func newTestManager(t *testing.T, ttl time.Duration) session.Manager {
	t.Helper()

	store, err := memory.New()
	if err != nil {
		t.Fatalf("memory.New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	manager, err := session.New(store, session.WithTTL(ttl), session.WithIDLength(session.MinIDLength))
	if err != nil {
		t.Fatalf("session.New() error = %v", err)
	}
	return manager
}

func TestManagerCreateGetSaveDelete(t *testing.T) {
	manager := newTestManager(t, time.Minute)
	ctx := context.Background()

	sess, err := manager.Create(ctx)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(sess.ID) != session.MinIDLength {
		t.Fatalf("Create() id length = %d, want %d", len(sess.ID), session.MinIDLength)
	}

	sess.Values["user_id"] = "1001"
	if err := manager.Save(ctx, sess); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := manager.Get(ctx, sess.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Values["user_id"] != "1001" {
		t.Fatalf("Get() user_id = %v, want 1001", got.Values["user_id"])
	}

	if err := manager.Delete(ctx, sess.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := manager.Get(ctx, sess.ID); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("Get() after Delete error = %v, want ErrNotFound", err)
	}
}

func TestManagerRefresh(t *testing.T) {
	manager := newTestManager(t, 50*time.Millisecond)
	ctx := context.Background()

	sess, err := manager.Create(ctx)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	oldExpiresAt := sess.ExpiresAt

	time.Sleep(20 * time.Millisecond)
	refreshed, err := manager.Refresh(ctx, sess.ID)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if !refreshed.ExpiresAt.After(oldExpiresAt) {
		t.Fatalf("Refresh() ExpiresAt = %v, want after %v", refreshed.ExpiresAt, oldExpiresAt)
	}
}

func TestManagerExpiration(t *testing.T) {
	manager := newTestManager(t, 10*time.Millisecond)
	ctx := context.Background()

	sess, err := manager.Create(ctx)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	time.Sleep(25 * time.Millisecond)
	_, err = manager.Get(ctx, sess.ID)
	if !errors.Is(err, session.ErrNotFound) && !errors.Is(err, session.ErrExpired) {
		t.Fatalf("Get() after expiration error = %v, want ErrNotFound or ErrExpired", err)
	}
}

func TestNewValidation(t *testing.T) {
	store, err := memory.New()
	if err != nil {
		t.Fatalf("memory.New() error = %v", err)
	}
	defer store.Close()

	if _, err := session.New(nil); !errors.Is(err, session.ErrNilStore) {
		t.Fatalf("New(nil) error = %v, want ErrNilStore", err)
	}
	if _, err := session.New(store, session.WithTTL(0)); !errors.Is(err, session.ErrInvalidTTL) {
		t.Fatalf("New() invalid ttl error = %v, want ErrInvalidTTL", err)
	}
	if _, err := session.New(store, session.WithIDLength(0)); !errors.Is(err, session.ErrInvalidID) {
		t.Fatalf("New() invalid id length error = %v, want ErrInvalidID", err)
	}
	if _, err := session.New(store, session.WithIDLength(session.MinIDLength-1)); !errors.Is(err, session.ErrInvalidID) {
		t.Fatalf("New() short id length error = %v, want ErrInvalidID", err)
	}
}
