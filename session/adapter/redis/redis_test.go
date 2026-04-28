package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/f2xme/gox/session"
	goredis "github.com/redis/go-redis/v9"
)

func setupRedis(t *testing.T) (session.Store, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	store, err := New(WithClient(client), WithKeyPrefix("test:session:"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() {
		if closer, ok := store.(session.Closer); ok {
			_ = closer.Close()
		}
	})
	return store, mr
}

func TestStoreSetGetDelete(t *testing.T) {
	store, mr := setupRedis(t)
	ctx := context.Background()
	now := time.Now()
	sess := &session.Session{
		ID:        "sid",
		Values:    map[string]any{"user_id": "1001"},
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(time.Minute),
	}

	if err := store.Set(ctx, sess, time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if !mr.Exists("test:session:sid") {
		t.Fatalf("redis key was not written with prefix")
	}

	got, err := store.Get(ctx, "sid")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Values["user_id"] != "1001" {
		t.Fatalf("Get() user_id = %v, want 1001", got.Values["user_id"])
	}

	if err := store.Delete(ctx, "sid"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := store.Get(ctx, "sid"); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("Get() deleted error = %v, want ErrNotFound", err)
	}
}

func TestStoreExpiration(t *testing.T) {
	store, mr := setupRedis(t)
	ctx := context.Background()
	sess := &session.Session{ID: "sid", Values: map[string]any{}}

	if err := store.Set(ctx, sess, time.Second); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	mr.FastForward(2 * time.Second)

	if _, err := store.Get(ctx, "sid"); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("Get() expired error = %v, want ErrNotFound", err)
	}
}

func TestStoreSetPreservesExpiresAt(t *testing.T) {
	store, _ := setupRedis(t)
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

func TestStoreValidation(t *testing.T) {
	store, _ := setupRedis(t)
	ctx := context.Background()

	if err := store.Set(ctx, nil, time.Minute); !errors.Is(err, session.ErrInvalidID) {
		t.Fatalf("Set(nil) error = %v, want ErrInvalidID", err)
	}
	if err := store.Set(ctx, &session.Session{ID: "sid"}, 0); !errors.Is(err, session.ErrInvalidTTL) {
		t.Fatalf("Set() invalid ttl error = %v, want ErrInvalidTTL", err)
	}
}

func TestCloseDoesNotCloseCallerOwnedClient(t *testing.T) {
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	store, err := New(WithClient(client))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	closer, ok := store.(session.Closer)
	if !ok {
		t.Fatal("store does not implement session.Closer")
	}
	if err := closer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("caller-owned client was closed: %v", err)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	mr := miniredis.RunT(t)
	store, err := New(WithAddr(mr.Addr()))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	closer, ok := store.(session.Closer)
	if !ok {
		t.Fatal("store does not implement session.Closer")
	}

	if err := closer.Close(); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
	if err := closer.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
}

func TestCloseCanRetryAfterFailure(t *testing.T) {
	client := &closeFailClient{err: errors.New("close failed")}
	store := &Store{
		client:     client,
		ownsClient: true,
	}

	if err := store.Close(); !errors.Is(err, client.err) {
		t.Fatalf("first Close() error = %v, want %v", err, client.err)
	}
	if client.calls != 1 {
		t.Fatalf("Close() calls after first close = %d, want 1", client.calls)
	}

	client.err = nil
	if err := store.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if client.calls != 2 {
		t.Fatalf("Close() calls after retry = %d, want 2", client.calls)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("third Close() error = %v", err)
	}
	if client.calls != 2 {
		t.Fatalf("Close() calls after idempotent close = %d, want 2", client.calls)
	}
}

type closeFailClient struct {
	goredis.UniversalClient
	err   error
	calls int
}

func (c *closeFailClient) Close() error {
	c.calls++
	return c.err
}
