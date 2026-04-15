package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestClaimsToMapClaims(t *testing.T) {
	now := time.Now()
	claims := &Claims{
		Subject:   "user123",
		Issuer:    "test-issuer",
		Audience:  []string{"app1", "app2"},
		ExpiresAt: now.Add(time.Hour),
		NotBefore: now,
		IssuedAt:  now,
		ID:        "token123",
		Custom: map[string]interface{}{
			"role": "admin",
			"org":  "test-org",
		},
	}

	mc := claims.ToMapClaims()

	if mc["sub"] != "user123" {
		t.Errorf("expected sub 'user123', got %v", mc["sub"])
	}
	if mc["iss"] != "test-issuer" {
		t.Errorf("expected iss 'test-issuer', got %v", mc["iss"])
	}
	if mc["jti"] != "token123" {
		t.Errorf("expected jti 'token123', got %v", mc["jti"])
	}
	if mc["role"] != "admin" {
		t.Errorf("expected role 'admin', got %v", mc["role"])
	}
	if mc["org"] != "test-org" {
		t.Errorf("expected org 'test-org', got %v", mc["org"])
	}
}

func TestFromMapClaims(t *testing.T) {
	now := time.Now()
	mc := jwt.MapClaims{
		"sub": "user123",
		"iss": "test-issuer",
		"aud": []interface{}{"app1", "app2"},
		"exp": float64(now.Add(time.Hour).Unix()),
		"nbf": float64(now.Unix()),
		"iat": float64(now.Unix()),
		"jti": "token123",
		"role": "admin",
		"org":  "test-org",
	}

	claims := FromMapClaims(mc)

	if claims.Subject != "user123" {
		t.Errorf("expected Subject 'user123', got %s", claims.Subject)
	}
	if claims.Issuer != "test-issuer" {
		t.Errorf("expected Issuer 'test-issuer', got %s", claims.Issuer)
	}
	if len(claims.Audience) != 2 {
		t.Errorf("expected 2 audiences, got %d", len(claims.Audience))
	}
	if claims.ID != "token123" {
		t.Errorf("expected ID 'token123', got %s", claims.ID)
	}
	if claims.Custom["role"] != "admin" {
		t.Errorf("expected role 'admin', got %v", claims.Custom["role"])
	}
	if claims.Custom["org"] != "test-org" {
		t.Errorf("expected org 'test-org', got %v", claims.Custom["org"])
	}
}

func TestFromMapClaimsWithStringAudience(t *testing.T) {
	mc := jwt.MapClaims{
		"sub": "user123",
		"aud": "single-app",
	}

	claims := FromMapClaims(mc)

	if len(claims.Audience) != 1 {
		t.Errorf("expected 1 audience, got %d", len(claims.Audience))
	}
	if claims.Audience[0] != "single-app" {
		t.Errorf("expected audience 'single-app', got %s", claims.Audience[0])
	}
}

func TestClaimsIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "expired token",
			expiresAt: time.Now().Add(-time.Hour),
			expected:  true,
		},
		{
			name:      "valid token",
			expiresAt: time.Now().Add(time.Hour),
			expected:  false,
		},
		{
			name:      "no expiration",
			expiresAt: time.Time{},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{ExpiresAt: tt.expiresAt}
			if got := claims.IsExpired(); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClaimsIsNotYetValid(t *testing.T) {
	tests := []struct {
		name      string
		notBefore time.Time
		expected  bool
	}{
		{
			name:      "not yet valid",
			notBefore: time.Now().Add(time.Hour),
			expected:  true,
		},
		{
			name:      "already valid",
			notBefore: time.Now().Add(-time.Hour),
			expected:  false,
		},
		{
			name:      "no not-before",
			notBefore: time.Time{},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{NotBefore: tt.notBefore}
			if got := claims.IsNotYetValid(); got != tt.expected {
				t.Errorf("IsNotYetValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClaimsRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	original := &Claims{
		Subject:   "user123",
		Issuer:    "test-issuer",
		Audience:  []string{"app1", "app2"},
		ExpiresAt: now.Add(time.Hour),
		NotBefore: now,
		IssuedAt:  now,
		ID:        "token123",
		Custom: map[string]interface{}{
			"role": "admin",
			"org":  "test-org",
		},
	}

	// Convert to MapClaims and back
	mc := original.ToMapClaims()
	restored := FromMapClaims(mc)

	if restored.Subject != original.Subject {
		t.Errorf("Subject mismatch: got %s, want %s", restored.Subject, original.Subject)
	}
	if restored.Issuer != original.Issuer {
		t.Errorf("Issuer mismatch: got %s, want %s", restored.Issuer, original.Issuer)
	}
	if len(restored.Audience) != len(original.Audience) {
		t.Errorf("Audience length mismatch: got %d, want %d", len(restored.Audience), len(original.Audience))
	}
	if restored.ID != original.ID {
		t.Errorf("ID mismatch: got %s, want %s", restored.ID, original.ID)
	}
	if restored.Custom["role"] != original.Custom["role"] {
		t.Errorf("Custom role mismatch: got %v, want %v", restored.Custom["role"], original.Custom["role"])
	}
}
