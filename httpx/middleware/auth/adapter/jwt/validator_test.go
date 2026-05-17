package jwt

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/f2xme/gox/httpx/middleware/auth"
	goxjwt "github.com/f2xme/gox/jwt"
)

func TestClaimsImplementsAuthClaims(t *testing.T) {
	var _ auth.Claims = (*goxjwt.Claims)(nil)
}

func TestTokenValidatorValidate(t *testing.T) {
	t.Parallel()
	j := goxjwt.NewHS256([]byte("secret"))
	v := NewTokenValidator(j)
	token, err := j.Generate(&goxjwt.Claims{Subject: "99", ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := v.Validate(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.GetUID() != 99 {
		t.Fatalf("uid: %d", claims.GetUID())
	}
}

func TestTokenValidatorNilJWT(t *testing.T) {
	t.Parallel()
	v := &TokenValidator{}
	_, err := v.Validate(context.Background(), "x")
	if !errors.Is(err, goxjwt.ErrNilJWT) {
		t.Fatalf("got %v want %v", err, goxjwt.ErrNilJWT)
	}
}
