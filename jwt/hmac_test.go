package jwt

import (
	"testing"
	"time"
)

func TestHS256Generate(t *testing.T) {
	j := NewHS256([]byte("test-secret"))

	claims := &Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
	}

	token, err := j.Generate(claims)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestHS256Parse(t *testing.T) {
	j := NewHS256([]byte("test-secret"))

	originalClaims := &Claims{
		Subject:   "user123",
		Issuer:    "test-issuer",
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
		ID:        "token123",
		Custom: map[string]interface{}{
			"role": "admin",
		},
	}

	token, err := j.Generate(originalClaims)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	parsedClaims, err := j.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if parsedClaims.Subject != originalClaims.Subject {
		t.Errorf("Subject = %v, want %v", parsedClaims.Subject, originalClaims.Subject)
	}
	if parsedClaims.Issuer != originalClaims.Issuer {
		t.Errorf("Issuer = %v, want %v", parsedClaims.Issuer, originalClaims.Issuer)
	}
	if parsedClaims.ID != originalClaims.ID {
		t.Errorf("ID = %v, want %v", parsedClaims.ID, originalClaims.ID)
	}
	if parsedClaims.Custom["role"] != "admin" {
		t.Errorf("Custom role = %v, want admin", parsedClaims.Custom["role"])
	}
}

func TestHS256ParseInvalidToken(t *testing.T) {
	j := NewHS256([]byte("test-secret"))

	_, err := j.Parse("invalid.token.here")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestHS256ParseExpiredToken(t *testing.T) {
	j := NewHS256([]byte("test-secret"))

	claims := &Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(-time.Hour), // expired
		IssuedAt:  time.Now().Add(-2 * time.Hour),
	}

	token, err := j.Generate(claims)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	_, err = j.Parse(token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestHS256ParseWrongSecret(t *testing.T) {
	j1 := NewHS256([]byte("secret1"))
	j2 := NewHS256([]byte("secret2"))

	claims := &Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
	}

	token, err := j1.Generate(claims)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	_, err = j2.Parse(token)
	if err == nil {
		t.Error("expected error when parsing with wrong secret")
	}
}

func TestHS256Refresh(t *testing.T) {
	j := NewHS256([]byte("test-secret"))

	originalClaims := &Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
		Custom: map[string]interface{}{
			"role": "admin",
		},
	}

	token, err := j.Generate(originalClaims)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Wait a bit to ensure new token has different timestamps
	time.Sleep(10 * time.Millisecond)

	newToken, err := j.Refresh(token, 2*time.Hour)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if newToken == token {
		t.Error("expected different token after refresh")
	}

	newClaims, err := j.Parse(newToken)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if newClaims.Subject != originalClaims.Subject {
		t.Errorf("Subject = %v, want %v", newClaims.Subject, originalClaims.Subject)
	}
	if newClaims.Custom["role"] != "admin" {
		t.Errorf("Custom role = %v, want admin", newClaims.Custom["role"])
	}
	if newClaims.ExpiresAt.Before(time.Now().Add(time.Hour)) {
		t.Error("expected new expiration to be extended")
	}
}

func TestHS256WithOptions(t *testing.T) {
	j := NewHS256(
		[]byte("test-secret"),
		WithIssuer("test-issuer"),
		WithExpiration(time.Hour),
		WithAudience("app1", "app2"),
	)

	claims := &Claims{
		Subject: "user123",
	}

	token, err := j.Generate(claims)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	parsedClaims, err := j.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if parsedClaims.Issuer != "test-issuer" {
		t.Errorf("Issuer = %v, want test-issuer", parsedClaims.Issuer)
	}
	if len(parsedClaims.Audience) != 2 {
		t.Errorf("Audience length = %v, want 2", len(parsedClaims.Audience))
	}
	if parsedClaims.ExpiresAt.IsZero() {
		t.Error("expected ExpiresAt to be set")
	}
}

func TestHS384(t *testing.T) {
	j := NewHS384([]byte("test-secret"))

	claims := &Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
	}

	token, err := j.Generate(claims)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	parsedClaims, err := j.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if parsedClaims.Subject != claims.Subject {
		t.Errorf("Subject = %v, want %v", parsedClaims.Subject, claims.Subject)
	}
}

func TestHS512(t *testing.T) {
	j := NewHS512([]byte("test-secret"))

	claims := &Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
	}

	token, err := j.Generate(claims)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	parsedClaims, err := j.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if parsedClaims.Subject != claims.Subject {
		t.Errorf("Subject = %v, want %v", parsedClaims.Subject, claims.Subject)
	}
}

func TestHS256ParseNotYetValid(t *testing.T) {
	j := NewHS256([]byte("test-secret"))

	claims := &Claims{
		Subject:   "user123",
		NotBefore: time.Now().Add(time.Hour), // not yet valid
		ExpiresAt: time.Now().Add(2 * time.Hour),
		IssuedAt:  time.Now(),
	}

	token, err := j.Generate(claims)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	_, err = j.Parse(token)
	if err == nil {
		t.Error("expected error for not yet valid token")
	}
}

func TestHS256RefreshInvalidToken(t *testing.T) {
	j := NewHS256([]byte("test-secret"))

	_, err := j.Refresh("invalid.token.here", time.Hour)
	if err == nil {
		t.Error("expected error when refreshing invalid token")
	}
}

