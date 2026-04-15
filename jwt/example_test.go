package jwt_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/f2xme/gox/jwt"
)

func ExampleNewHS256() {
	// Create JWT with HMAC-SHA256
	j := jwt.NewHS256([]byte("my-secret-key"))

	// Create claims
	claims := &jwt.Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IssuedAt:  time.Now(),
		Custom: map[string]interface{}{
			"role": "admin",
		},
	}

	// Generate token
	token, err := j.Generate(claims)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Token generated: %t\n", len(token) > 0)
	// Output: Token generated: true
}

func ExampleNewHS256_withOptions() {
	// Create JWT with options
	j := jwt.NewHS256(
		[]byte("my-secret-key"),
		jwt.WithIssuer("my-app"),
		jwt.WithExpiration(24*time.Hour),
		jwt.WithAudience("web", "mobile"),
	)

	// Create claims (issuer, expiration, and audience will be set automatically)
	claims := &jwt.Claims{
		Subject: "user123",
	}

	// Generate token
	token, err := j.Generate(claims)
	if err != nil {
		panic(err)
	}

	// Parse token
	parsedClaims, err := j.Parse(token)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Subject: %s\n", parsedClaims.Subject)
	fmt.Printf("Issuer: %s\n", parsedClaims.Issuer)
	fmt.Printf("Audience: %v\n", parsedClaims.Audience)
	// Output:
	// Subject: user123
	// Issuer: my-app
	// Audience: [web mobile]
}

func ExampleJWT_Parse() {
	j := jwt.NewHS256([]byte("my-secret-key"))

	// Generate token
	claims := &jwt.Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
	}
	token, _ := j.Generate(claims)

	// Parse token
	parsedClaims, err := j.Parse(token)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Subject: %s\n", parsedClaims.Subject)
	// Output: Subject: user123
}

func ExampleJWT_Refresh() {
	j := jwt.NewHS256([]byte("my-secret-key"))

	// Generate token
	claims := &jwt.Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
	}
	token, _ := j.Generate(claims)

	// Refresh token with new expiration
	newToken, err := j.Refresh(token, 24*time.Hour)
	if err != nil {
		panic(err)
	}

	// Parse new token
	newClaims, _ := j.Parse(newToken)
	fmt.Printf("Subject: %s\n", newClaims.Subject)
	// Output: Subject: user123
}

func ExampleNewRS256() {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// Create JWT with RSA-SHA256
	j := jwt.NewRS256(privateKey, &privateKey.PublicKey)

	// Create claims
	claims := &jwt.Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
	}

	// Generate token
	token, err := j.Generate(claims)
	if err != nil {
		panic(err)
	}

	// Parse token
	parsedClaims, err := j.Parse(token)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Subject: %s\n", parsedClaims.Subject)
	// Output: Subject: user123
}

func ExampleNewES256() {
	// Generate ECDSA key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	// Create JWT with ECDSA-SHA256
	j := jwt.NewES256(privateKey, &privateKey.PublicKey)

	// Create claims
	claims := &jwt.Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
	}

	// Generate token
	token, err := j.Generate(claims)
	if err != nil {
		panic(err)
	}

	// Parse token
	parsedClaims, err := j.Parse(token)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Subject: %s\n", parsedClaims.Subject)
	// Output: Subject: user123
}

func ExampleClaims_IsExpired() {
	claims := &jwt.Claims{
		ExpiresAt: time.Now().Add(-time.Hour), // expired 1 hour ago
	}

	if claims.IsExpired() {
		fmt.Println("Token is expired")
	}
	// Output: Token is expired
}

func ExampleClaims_IsNotYetValid() {
	claims := &jwt.Claims{
		NotBefore: time.Now().Add(time.Hour), // valid in 1 hour
	}

	if claims.IsNotYetValid() {
		fmt.Println("Token is not yet valid")
	}
	// Output: Token is not yet valid
}
