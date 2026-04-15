package jwt

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// rsaJWT 使用 RSA 算法实现 JWT 接口
type rsaJWT struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	method     jwt.SigningMethod
	opts       Options
}

// NewRS256 使用 RS256 算法创建新的 JWT 实例
func NewRS256(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, opts ...Option) JWT {
	return newRSA(privateKey, publicKey, jwt.SigningMethodRS256, opts...)
}

// NewRS384 使用 RS384 算法创建新的 JWT 实例
func NewRS384(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, opts ...Option) JWT {
	return newRSA(privateKey, publicKey, jwt.SigningMethodRS384, opts...)
}

// NewRS512 使用 RS512 算法创建新的 JWT 实例
func NewRS512(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, opts ...Option) JWT {
	return newRSA(privateKey, publicKey, jwt.SigningMethodRS512, opts...)
}

func newRSA(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, method jwt.SigningMethod, opts ...Option) JWT {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &rsaJWT{
		privateKey: privateKey,
		publicKey:  publicKey,
		method:     method,
		opts:       options,
	}
}

// Generate 使用给定的声明创建新的 JWT 令牌
func (r *rsaJWT) Generate(claims *Claims) (string, error) {
	if claims == nil {
		return "", ErrNilClaims
	}

	// Create a shallow copy to avoid modifying input
	workingClaims := *claims

	applyDefaultClaims(&workingClaims, &r.opts)

	mapClaims := workingClaims.ToMapClaims()

	token := jwt.NewWithClaims(r.method, mapClaims)

	tokenString, err := token.SignedString(r.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// Parse 解析并验证 JWT 令牌字符串
func (r *rsaJWT) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != r.method {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}
		return r.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	claims := FromMapClaims(mapClaims)

	if err := checkRevocation(context.Background(), r.opts.Revoker, claims.ID); err != nil {
		return nil, err
	}

	return claims, nil
}

// Refresh 生成具有更新过期时间的新令牌
func (r *rsaJWT) Refresh(tokenString string, expiration time.Duration) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != r.method {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}
		return r.publicKey, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidClaims
	}

	claims := FromMapClaims(mapClaims)

	if err := checkRevocation(context.Background(), r.opts.Revoker, claims.ID); err != nil {
		return "", err
	}

	updateClaimsTimestamp(claims, expiration)

	return r.Generate(claims)
}
