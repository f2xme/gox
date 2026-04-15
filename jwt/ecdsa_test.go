package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"
)

func TestES256Generate(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key: %v", err)
	}

	j := NewES256(privateKey, &privateKey.PublicKey)

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

func TestES256Parse(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key: %v", err)
	}

	j := NewES256(privateKey, &privateKey.PublicKey)

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

func TestES256ParseWrongKey(t *testing.T) {
	privateKey1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key: %v", err)
	}

	privateKey2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key: %v", err)
	}

	j1 := NewES256(privateKey1, &privateKey1.PublicKey)
	j2 := NewES256(privateKey2, &privateKey2.PublicKey)

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
		t.Error("expected error when parsing with wrong key")
	}
}

func TestES256Refresh(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key: %v", err)
	}

	j := NewES256(privateKey, &privateKey.PublicKey)

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
}

func TestES384(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key: %v", err)
	}

	j := NewES384(privateKey, &privateKey.PublicKey)

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

func TestES512(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key: %v", err)
	}

	j := NewES512(privateKey, &privateKey.PublicKey)

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

func TestES256RefreshInvalidToken(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key: %v", err)
	}

	j := NewES256(privateKey, &privateKey.PublicKey)

	_, err = j.Refresh("invalid.token.here", time.Hour)
	if err == nil {
		t.Error("expected error when refreshing invalid token")
	}
}

func TestES256ParseNotYetValid(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key: %v", err)
	}

	j := NewES256(privateKey, &privateKey.PublicKey)

	claims := &Claims{
		Subject:   "user123",
		NotBefore: time.Now().Add(time.Hour),
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
