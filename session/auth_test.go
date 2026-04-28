package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/f2xme/gox/session"
	"github.com/f2xme/gox/session/adapter/memory"
)

func newAuthTestManager(t *testing.T) session.Manager {
	t.Helper()

	store, err := memory.New()
	if err != nil {
		t.Fatalf("memory.New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	manager, err := session.New(store, session.WithTTL(time.Minute))
	if err != nil {
		t.Fatalf("session.New() error = %v", err)
	}
	return manager
}

func TestValidatorValidate(t *testing.T) {
	manager := newAuthTestManager(t)
	sess, err := manager.Create(t.Context())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	sess.Values[session.DefaultUIDKey] = int64(1001)
	sess.Values["role"] = "admin"
	if err := manager.Save(t.Context(), sess); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	claims, err := (session.Validator{Manager: manager}).Validate(t.Context(), sess.ID)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if claims.GetUID() != 1001 {
		t.Fatalf("GetUID() = %d, want 1001", claims.GetUID())
	}
	if role, ok := claims.Get("role"); !ok || role != "admin" {
		t.Fatalf("Get(role) = %v, %v, want admin, true", role, ok)
	}
	if sid, ok := claims.Get(session.ClaimsSessionIDKey); !ok || sid != sess.ID {
		t.Fatalf("Get(session_id) = %v, %v, want %s, true", sid, ok, sess.ID)
	}
}

func TestNewValidator(t *testing.T) {
	manager := newAuthTestManager(t)
	validator := session.NewValidator(
		manager,
		session.WithValidatorUIDKey("uid"),
		session.WithRefreshThreshold(time.Minute),
	)

	sess, err := manager.Create(t.Context())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	sess.Values["uid"] = int64(1001)
	if err := manager.Save(t.Context(), sess); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	claims, err := validator.Validate(t.Context(), sess.ID)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if claims.GetUID() != 1001 {
		t.Fatalf("GetUID() = %d, want 1001", claims.GetUID())
	}
}

func TestValidatorCustomUIDKey(t *testing.T) {
	manager := newAuthTestManager(t)
	sess, err := manager.Create(t.Context())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	sess.Values["uid"] = "42"
	if err := manager.Save(t.Context(), sess); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	claims, err := (session.Validator{Manager: manager, UIDKey: "uid"}).Validate(t.Context(), sess.ID)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if claims.GetUID() != 42 {
		t.Fatalf("GetUID() = %d, want 42", claims.GetUID())
	}
}

func TestValidatorRefreshesNearExpiration(t *testing.T) {
	store, err := memory.New()
	if err != nil {
		t.Fatalf("memory.New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	manager, err := session.New(store, session.WithTTL(80*time.Millisecond))
	if err != nil {
		t.Fatalf("session.New() error = %v", err)
	}
	sess, err := manager.Create(t.Context())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	sess.Values[session.DefaultUIDKey] = int64(1001)
	if err := manager.Save(t.Context(), sess); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	oldExpiresAt := sess.ExpiresAt

	time.Sleep(50 * time.Millisecond)
	claims, err := (session.Validator{
		Manager:          manager,
		RefreshThreshold: 40 * time.Millisecond,
	}).Validate(t.Context(), sess.ID)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if claims.GetUID() != 1001 {
		t.Fatalf("GetUID() = %d, want 1001", claims.GetUID())
	}

	refreshed, err := manager.Get(t.Context(), sess.ID)
	if err != nil {
		t.Fatalf("Get() refreshed error = %v", err)
	}
	if !refreshed.ExpiresAt.After(oldExpiresAt) {
		t.Fatalf("refreshed ExpiresAt = %v, want after %v", refreshed.ExpiresAt, oldExpiresAt)
	}
}

func TestValidatorDoesNotRefreshInvalidSession(t *testing.T) {
	store, err := memory.New()
	if err != nil {
		t.Fatalf("memory.New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	manager, err := session.New(store, session.WithTTL(80*time.Millisecond))
	if err != nil {
		t.Fatalf("session.New() error = %v", err)
	}
	sess, err := manager.Create(t.Context())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	sess.Values[session.DefaultUIDKey] = "bad"
	if err := manager.Save(t.Context(), sess); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	oldExpiresAt := sess.ExpiresAt

	time.Sleep(50 * time.Millisecond)
	_, err = (session.Validator{
		Manager:          manager,
		RefreshThreshold: 40 * time.Millisecond,
	}).Validate(t.Context(), sess.ID)
	if !errors.Is(err, session.ErrInvalidSession) {
		t.Fatalf("Validate() error = %v, want ErrInvalidSession", err)
	}

	got, err := manager.Get(t.Context(), sess.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !got.ExpiresAt.Equal(oldExpiresAt) {
		t.Fatalf("invalid session was refreshed: got %v, want %v", got.ExpiresAt, oldExpiresAt)
	}
}

func TestValidatorValidateReceivesContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := (session.Validator{Manager: contextAwareManager{}}).Validate(ctx, "sid")
	if !errors.Is(err, session.ErrInvalidSession) || !errors.Is(err, context.Canceled) {
		t.Fatalf("Validate() error = %v, want ErrInvalidSession wrapping context.Canceled", err)
	}
}

type contextAwareManager struct{}

func (contextAwareManager) Create(ctx context.Context) (*session.Session, error) {
	return nil, nil
}

func (contextAwareManager) Get(ctx context.Context, id string) (*session.Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &session.Session{
		ID:        id,
		Values:    map[string]any{session.DefaultUIDKey: int64(1001)},
		ExpiresAt: time.Now().Add(time.Minute),
	}, nil
}

func (contextAwareManager) Save(ctx context.Context, sess *session.Session) error {
	return nil
}

func (contextAwareManager) Refresh(ctx context.Context, id string) (*session.Session, error) {
	return nil, nil
}

func (contextAwareManager) Delete(ctx context.Context, id string) error {
	return nil
}

func (contextAwareManager) Destroy(ctx context.Context, id string) error {
	return nil
}

func TestValidatorInvalidSession(t *testing.T) {
	manager := newAuthTestManager(t)

	if _, err := (session.Validator{Manager: manager}).Validate(t.Context(), "missing"); !errors.Is(err, session.ErrInvalidSession) {
		t.Fatalf("Validate() missing error = %v, want ErrInvalidSession", err)
	}

	sess, err := manager.Create(t.Context())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	sess.Values[session.DefaultUIDKey] = "bad"
	if err := manager.Save(t.Context(), sess); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if _, err := (session.Validator{Manager: manager}).Validate(t.Context(), sess.ID); !errors.Is(err, session.ErrInvalidSession) {
		t.Fatalf("Validate() bad uid error = %v, want ErrInvalidSession", err)
	}
}

func TestValidatorSupportsJSONNumberShape(t *testing.T) {
	manager := newAuthTestManager(t)
	sess, err := manager.Create(t.Context())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	sess.Values[session.DefaultUIDKey] = float64(1001)
	if err := manager.Save(t.Context(), sess); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	claims, err := (session.Validator{Manager: manager}).Validate(t.Context(), sess.ID)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if claims.GetUID() != 1001 {
		t.Fatalf("GetUID() = %d, want 1001", claims.GetUID())
	}

	sess.Values[session.DefaultUIDKey] = float64(10.5)
	if err := manager.Save(t.Context(), sess); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if _, err := (session.Validator{Manager: manager}).Validate(t.Context(), sess.ID); !errors.Is(err, session.ErrInvalidSession) {
		t.Fatalf("Validate() non-integer float error = %v, want ErrInvalidSession", err)
	}
}
