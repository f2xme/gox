package jwt

import (
	"errors"
	"testing"
	"time"

	"github.com/f2xme/gox/httpx/middleware/auth"
)

func TestClaimsImplementsAuthClaims(t *testing.T) {
	var _ auth.Claims = (*Claims)(nil)
}

func TestClaimsGetUID(t *testing.T) {
	t.Parallel()
	if (&Claims{Subject: " 42 "}).GetUID() != 42 {
		t.Fatal("expected trimmed subject to parse")
	}
	if (&Claims{Subject: "x"}).GetUID() != 0 {
		t.Fatal("invalid subject should yield 0")
	}
	if (&Claims{}).GetUID() != 0 {
		t.Fatal("empty subject should yield 0")
	}
}

func TestClaimsGet(t *testing.T) {
	t.Parallel()
	now := time.Unix(1000, 0)
	c := &Claims{
		Subject:   "1",
		Issuer:    "iss",
		Audience:  []string{"a"},
		ExpiresAt: now,
		NotBefore: now,
		IssuedAt:  now,
		ID:        "jid",
		Custom:    map[string]interface{}{"role": "admin"},
	}
	if v, ok := c.Get("sub"); !ok || v != "1" {
		t.Fatalf("sub: %v %v", v, ok)
	}
	if _, ok := (&Claims{}).Get("iss"); ok {
		t.Fatal("empty iss should be missing")
	}
	if v, ok := c.Get("role"); !ok || v != "admin" {
		t.Fatalf("custom: %v %v", v, ok)
	}
}

func TestTokenValidatorValidate(t *testing.T) {
	t.Parallel()
	j := NewHS256([]byte("secret"))
	v := &TokenValidator{JWT: j}
	token, err := j.Generate(&Claims{Subject: "99", ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	ac, err := v.Validate(token)
	if err != nil {
		t.Fatal(err)
	}
	if ac.GetUID() != 99 {
		t.Fatalf("uid: %d", ac.GetUID())
	}
}

func TestTokenValidatorNilJWT(t *testing.T) {
	t.Parallel()
	v := &TokenValidator{}
	_, err := v.Validate("x")
	if !errors.Is(err, ErrNilJWT) {
		t.Fatalf("got %v want %v", err, ErrNilJWT)
	}
}
