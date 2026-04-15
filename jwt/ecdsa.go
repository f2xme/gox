package jwt

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ecdsaJWT 使用 ECDSA 算法实现 JWT 接口
type ecdsaJWT struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	method     jwt.SigningMethod
	opts       Options
}

// NewES256 使用 ES256 算法创建新的 JWT 实例
func NewES256(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, opts ...Option) JWT {
	return newECDSA(privateKey, publicKey, jwt.SigningMethodES256, opts...)
}

// NewES384 使用 ES384 算法创建新的 JWT 实例
func NewES384(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, opts ...Option) JWT {
	return newECDSA(privateKey, publicKey, jwt.SigningMethodES384, opts...)
}

// NewES512 使用 ES512 算法创建新的 JWT 实例
func NewES512(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, opts ...Option) JWT {
	return newECDSA(privateKey, publicKey, jwt.SigningMethodES512, opts...)
}

func newECDSA(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, method jwt.SigningMethod, opts ...Option) JWT {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &ecdsaJWT{
		privateKey: privateKey,
		publicKey:  publicKey,
		method:     method,
		opts:       options,
	}
}

// Generate 使用给定的声明创建新的 JWT 令牌
func (e *ecdsaJWT) Generate(claims *Claims) (string, error) {
	if claims == nil {
		return "", ErrNilClaims
	}

	// Create a shallow copy to avoid modifying input
	workingClaims := *claims

	applyDefaultClaims(&workingClaims, &e.opts)

	mapClaims := workingClaims.ToMapClaims()

	token := jwt.NewWithClaims(e.method, mapClaims)

	tokenString, err := token.SignedString(e.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// Parse 解析并验证 JWT 令牌字符串
func (e *ecdsaJWT) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != e.method {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}
		return e.publicKey, nil
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

	if err := checkRevocation(context.Background(), e.opts.Revoker, claims.ID); err != nil {
		return nil, err
	}

	return claims, nil
}

// Refresh 生成具有更新过期时间的新令牌
func (e *ecdsaJWT) Refresh(tokenString string, expiration time.Duration) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != e.method {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}
		return e.publicKey, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidClaims
	}

	claims := FromMapClaims(mapClaims)

	if err := checkRevocation(context.Background(), e.opts.Revoker, claims.ID); err != nil {
		return "", err
	}

	updateClaimsTimestamp(claims, expiration)

	return e.Generate(claims)
}
