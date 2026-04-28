package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/f2xme/gox/httpx/mock"
	"github.com/f2xme/gox/session"
	"github.com/f2xme/gox/session/adapter/memory"
)

func TestSessionValidatorValidate(t *testing.T) {
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
	sess, err := manager.Create(t.Context())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	sess.Values[DefaultSessionUIDKey] = int64(1001)
	if err := manager.Save(t.Context(), sess); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	claims, err := (SessionValidator{Manager: manager}).Validate(t.Context(), sess.ID)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if claims.GetUID() != 1001 {
		t.Fatalf("GetUID() = %d, want 1001", claims.GetUID())
	}
}

func TestNewSessionValidator(t *testing.T) {
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
	sess, err := manager.Create(t.Context())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	sess.Values["uid"] = int64(1001)
	if err := manager.Save(t.Context(), sess); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	validator := NewSessionValidator(manager, session.WithValidatorUIDKey("uid"))
	claims, err := validator.Validate(t.Context(), sess.ID)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if claims.GetUID() != 1001 {
		t.Fatalf("GetUID() = %d, want 1001", claims.GetUID())
	}
}

func TestNewSessionExtractor(t *testing.T) {
	ctx := mock.NewMockContext(http.MethodGet, "/profile")
	ctx.SetCookie(&http.Cookie{Name: DefaultSessionCookieName, Value: "sid-1"})

	token := NewSessionExtractor("")(ctx)
	if token != "sid-1" {
		t.Fatalf("NewSessionExtractor() = %q, want sid-1", token)
	}

	missing := NewSessionExtractor("missing")(ctx)
	if missing != "" {
		t.Fatalf("NewSessionExtractor() missing = %q, want empty", missing)
	}
}
