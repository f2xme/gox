package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisRevoker(t *testing.T) {
	// Start miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	revoker := NewRedisRevoker(client)
	ctx := context.Background()

	// Test Revoke
	err = revoker.Revoke(ctx, "token123", time.Hour)
	if err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}

	// Test IsRevoked - should be revoked
	revoked, err := revoker.IsRevoked(ctx, "token123")
	if err != nil {
		t.Fatalf("IsRevoked() error = %v", err)
	}
	if !revoked {
		t.Error("expected token to be revoked")
	}

	// Test IsRevoked - should not be revoked
	revoked, err = revoker.IsRevoked(ctx, "token456")
	if err != nil {
		t.Fatalf("IsRevoked() error = %v", err)
	}
	if revoked {
		t.Error("expected token to not be revoked")
	}
}

func TestRedisRevokerExpiration(t *testing.T) {
	// Start miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	revoker := NewRedisRevoker(client)
	ctx := context.Background()

	// Revoke with short expiration
	err = revoker.Revoke(ctx, "token123", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}

	// Should be revoked immediately
	revoked, err := revoker.IsRevoked(ctx, "token123")
	if err != nil {
		t.Fatalf("IsRevoked() error = %v", err)
	}
	if !revoked {
		t.Error("expected token to be revoked")
	}

	// Fast forward time in miniredis
	mr.FastForward(200 * time.Millisecond)

	// Should not be revoked after expiration
	revoked, err = revoker.IsRevoked(ctx, "token123")
	if err != nil {
		t.Fatalf("IsRevoked() error = %v", err)
	}
	if revoked {
		t.Error("expected token to not be revoked after expiration")
	}
}

func TestRedisRevokerWithPrefix(t *testing.T) {
	// Start miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	revoker := NewRedisRevoker(client, WithPrefix("test:"))
	ctx := context.Background()

	// Revoke token
	err = revoker.Revoke(ctx, "token123", time.Hour)
	if err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}

	// Check if key exists with prefix
	exists := mr.Exists("test:revoked:token123")
	if !exists {
		t.Error("expected key to exist with prefix")
	}

	// Should be revoked
	revoked, err := revoker.IsRevoked(ctx, "token123")
	if err != nil {
		t.Fatalf("IsRevoked() error = %v", err)
	}
	if !revoked {
		t.Error("expected token to be revoked")
	}
}

func TestJWTWithRevoker(t *testing.T) {
	// Start miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	revoker := NewRedisRevoker(client)

	// Create JWT with revoker
	j := NewHS256([]byte("test-secret"), WithRevoker(revoker))

	claims := &Claims{
		Subject:   "user123",
		ID:        "token123",
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
	}

	// Generate token
	token, err := j.Generate(claims)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Parse token - should succeed
	parsedClaims, err := j.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsedClaims.Subject != "user123" {
		t.Errorf("Subject = %v, want user123", parsedClaims.Subject)
	}

	// Revoke token
	ctx := context.Background()
	err = revoker.Revoke(ctx, "token123", time.Hour)
	if err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}

	// Parse token - should fail
	_, err = j.Parse(token)
	if err == nil {
		t.Error("expected error when parsing revoked token")
	}
}

func TestRedisRevokerErrors(t *testing.T) {
	// Start miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	// Close miniredis to simulate connection error
	mr.Close()

	revoker := NewRedisRevoker(client)
	ctx := context.Background()

	// Test Revoke with connection error
	err = revoker.Revoke(ctx, "token123", time.Hour)
	if err == nil {
		t.Error("expected error when Redis is unavailable")
	}

	// Test IsRevoked with connection error
	_, err = revoker.IsRevoked(ctx, "token123")
	if err == nil {
		t.Error("expected error when Redis is unavailable")
	}
}
