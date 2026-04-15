package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// hmacJWT 使用 HMAC 算法实现 JWT 接口
type hmacJWT struct {
	secret []byte
	method jwt.SigningMethod
	opts   Options
}

// NewHS256 使用 HS256 算法创建新的 JWT 实例
func NewHS256(secret []byte, opts ...Option) JWT {
	return newHMAC(secret, jwt.SigningMethodHS256, opts...)
}

// NewHS384 使用 HS384 算法创建新的 JWT 实例
func NewHS384(secret []byte, opts ...Option) JWT {
	return newHMAC(secret, jwt.SigningMethodHS384, opts...)
}

// NewHS512 使用 HS512 算法创建新的 JWT 实例
func NewHS512(secret []byte, opts ...Option) JWT {
	return newHMAC(secret, jwt.SigningMethodHS512, opts...)
}

func newHMAC(secret []byte, method jwt.SigningMethod, opts ...Option) JWT {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &hmacJWT{
		secret: secret,
		method: method,
		opts:   options,
	}
}

// Generate 使用给定的声明创建新的 JWT 令牌
func (h *hmacJWT) Generate(claims *Claims) (string, error) {
	if claims == nil {
		return "", ErrNilClaims
	}

	// Create a shallow copy to avoid modifying input
	workingClaims := *claims

	applyDefaultClaims(&workingClaims, &h.opts)

	mapClaims := workingClaims.ToMapClaims()

	token := jwt.NewWithClaims(h.method, mapClaims)

	tokenString, err := token.SignedString(h.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// Parse 解析并验证 JWT 令牌字符串
func (h *hmacJWT) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != h.method {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}
		return h.secret, nil
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

	if err := checkRevocation(context.Background(), h.opts.Revoker, claims.ID); err != nil {
		return nil, err
	}

	return claims, nil
}

// Refresh 生成具有更新过期时间的新令牌
func (h *hmacJWT) Refresh(tokenString string, expiration time.Duration) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != h.method {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}
		return h.secret, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidClaims
	}

	claims := FromMapClaims(mapClaims)

	if err := checkRevocation(context.Background(), h.opts.Revoker, claims.ID); err != nil {
		return "", err
	}

	updateClaimsTimestamp(claims, expiration)

	return h.Generate(claims)
}
