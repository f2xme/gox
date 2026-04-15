package auth

import (
	"net/http"
	"slices"
	"testing"

	"github.com/f2xme/gox/httpx"
)

func TestDefaultUser_GetID(t *testing.T) {
	user := &DefaultUser{ID: "user-123"}
	if got := user.GetID(); got != "user-123" {
		t.Errorf("GetID() = %q, want %q", got, "user-123")
	}
}

func TestDefaultUser_GetRole(t *testing.T) {
	user := &DefaultUser{Role: "admin"}
	if got := user.GetRole(); got != "admin" {
		t.Errorf("GetRole() = %q, want %q", got, "admin")
	}
}

func TestDefaultUser_GetRoles(t *testing.T) {
	t.Run("returns copy of roles", func(t *testing.T) {
		roles := []string{"admin", "editor"}
		user := &DefaultUser{Roles: roles}
		got := user.GetRoles()
		if len(got) != 2 || got[0] != "admin" || got[1] != "editor" {
			t.Errorf("GetRoles() = %v, want %v", got, roles)
		}

		// Verify it's a copy, not the original slice
		got[0] = "modified"
		if user.Roles[0] != "admin" {
			t.Error("GetRoles() returned original slice, want copy")
		}
	})

	t.Run("returns nil for nil roles", func(t *testing.T) {
		user := &DefaultUser{Roles: nil}
		got := user.GetRoles()
		if got != nil {
			t.Errorf("GetRoles() = %v, want nil", got)
		}
	})
}

func TestDefaultUser_HasRole(t *testing.T) {
	tests := []struct {
		name     string
		user     *DefaultUser
		role     string
		expected bool
	}{
		{
			name:     "primary role match",
			user:     &DefaultUser{Role: "admin"},
			role:     "admin",
			expected: true,
		},
		{
			name:     "roles array match",
			user:     &DefaultUser{Role: "user", Roles: []string{"editor", "viewer"}},
			role:     "editor",
			expected: true,
		},
		{
			name:     "no match",
			user:     &DefaultUser{Role: "user", Roles: []string{"viewer"}},
			role:     "admin",
			expected: false,
		},
		{
			name:     "empty roles",
			user:     &DefaultUser{},
			role:     "admin",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.HasRole(tt.role); got != tt.expected {
				t.Errorf("HasRole(%q) = %v, want %v", tt.role, got, tt.expected)
			}
		})
	}
}

func TestDefaultUser_HasPermission(t *testing.T) {
	tests := []struct {
		name     string
		user     *DefaultUser
		perm     string
		expected bool
	}{
		{
			name:     "permission exists",
			user:     &DefaultUser{Permissions: []string{"user:read", "user:write"}},
			perm:     "user:read",
			expected: true,
		},
		{
			name:     "permission not exists",
			user:     &DefaultUser{Permissions: []string{"user:read"}},
			perm:     "user:delete",
			expected: false,
		},
		{
			name:     "empty permissions",
			user:     &DefaultUser{},
			perm:     "user:read",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.HasPermission(tt.perm); got != tt.expected {
				t.Errorf("HasPermission(%q) = %v, want %v", tt.perm, got, tt.expected)
			}
		})
	}
}

func TestDefaultUser_IsBanned(t *testing.T) {
	tests := []struct {
		name     string
		user     *DefaultUser
		expected bool
	}{
		{
			name:     "user is banned",
			user:     &DefaultUser{ID: "user-1", Banned: true},
			expected: true,
		},
		{
			name:     "user is not banned",
			user:     &DefaultUser{ID: "user-2", Banned: false},
			expected: false,
		},
		{
			name:     "default value is not banned",
			user:     &DefaultUser{ID: "user-3"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.IsBanned(); got != tt.expected {
				t.Errorf("IsBanned() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSetUser_GetUser(t *testing.T) {
	ctx := newMockContext()
	user := &DefaultUser{
		ID:   "user-123",
		Role: "admin",
	}

	SetUser(ctx, user)

	got, ok := GetUser(ctx)
	if !ok {
		t.Fatal("GetUser() returned false, want true")
	}
	if got.GetID() != "user-123" {
		t.Errorf("GetUser().GetID() = %q, want %q", got.GetID(), "user-123")
	}
	if got.GetRole() != "admin" {
		t.Errorf("GetUser().GetRole() = %q, want %q", got.GetRole(), "admin")
	}
}

func TestGetUser_NotSet(t *testing.T) {
	ctx := newMockContext()
	_, ok := GetUser(ctx)
	if ok {
		t.Error("GetUser() returned true for empty context, want false")
	}
}

// customUserImpl is a test implementation of User interface
type customUserImpl struct {
	uid         string
	role        string
	permissions []string
}

func (c *customUserImpl) GetID() string             { return c.uid }
func (c *customUserImpl) GetRole() string           { return c.role }
func (c *customUserImpl) GetRoles() []string        { return []string{c.role} }
func (c *customUserImpl) HasRole(role string) bool  { return c.role == role }
func (c *customUserImpl) HasPermission(string) bool { return true } // wildcard
func (c *customUserImpl) IsBanned() bool            { return false }

func TestUserInterface_CustomImplementation(t *testing.T) {
	// Test that custom User implementations work
	impl := &customUserImpl{
		uid:         "custom-123",
		role:        "superadmin",
		permissions: []string{"*"},
	}

	ctx := newMockContext()
	SetUser(ctx, impl)

	user, ok := GetUser(ctx)
	if !ok {
		t.Fatal("GetUser() returned false, want true")
	}
	if got := user.GetID(); got != "custom-123" {
		t.Errorf("user.GetID() = %q, want %q", got, "custom-123")
	}
	if !user.HasRole("superadmin") {
		t.Error("user.HasRole(superadmin) = false, want true")
	}
	if !user.HasPermission("anything") {
		t.Error("user.HasPermission(anything) = false, want true")
	}
}

// userClaims is a test implementation of Claims and User interfaces
type userClaims struct {
	UserID      string
	Role        string
	Roles       []string
	Permissions []string
}

func (c *userClaims) GetSubject() string         { return c.UserID }
func (c *userClaims) Get(key string) (any, bool) { return nil, false }
func (c *userClaims) GetID() string              { return c.UserID }
func (c *userClaims) GetRole() string            { return c.Role }
func (c *userClaims) GetRoles() []string         { return c.Roles }
func (c *userClaims) HasRole(role string) bool {
	if c.Role == role {
		return true
	}
	return slices.Contains(c.Roles, role)
}
func (c *userClaims) HasPermission(perm string) bool {
	return slices.Contains(c.Permissions, perm)
}
func (c *userClaims) IsBanned() bool { return false }

func TestAuth_Integration_WithUser(t *testing.T) {
	// Test integration: Claims that also implement User interface
	claims := &userClaims{
		UserID:      "user-789",
		Role:        "admin",
		Roles:       []string{"admin", "editor"},
		Permissions: []string{"user:read", "user:write", "user:delete"},
	}

	validator := &mockValidator{
		tokenToClaims: map[string]Claims{
			"user-token": claims,
		},
	}

	middleware := New(WithValidator(validator))

	handler := middleware(func(ctx httpx.Context) error {
		// Test Claims access (existing)
		c := GetClaims(ctx)
		if c == nil {
			t.Fatal("expected claims in context")
		}
		if c.GetSubject() != "user-789" {
			t.Errorf("GetSubject() = %q, want %q", c.GetSubject(), "user-789")
		}

		// Test User access (new)
		// Manually set user since middleware doesn't auto-convert yet
		if u, ok := c.(User); ok {
			SetUser(ctx, u)
		}

		user, ok := GetUser(ctx)
		if !ok {
			t.Fatal("expected user in context")
		}
		if got := user.GetID(); got != "user-789" {
			t.Errorf("user.GetID() = %q, want %q", got, "user-789")
		}
		if !user.HasRole("admin") {
			t.Error("user.HasRole(admin) = false, want true")
		}
		if !user.HasPermission("user:delete") {
			t.Error("user.HasPermission(user:delete) = false, want true")
		}

		return ctx.JSON(http.StatusOK, map[string]any{"ok": true})
	})

	ctx := newMockContext()
	ctx.headers["Authorization"] = "Bearer user-token"

	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ctx.respCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, ctx.respCode)
	}
}
