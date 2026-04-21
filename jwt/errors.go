package jwt

import "errors"

var (
	// ErrNilClaims 当声明参数为 nil 时返回
	ErrNilClaims = errors.New("jwt: claims cannot be nil")

	// ErrInvalidToken 当令牌验证失败时返回
	ErrInvalidToken = errors.New("jwt: invalid token")

	// ErrInvalidClaims 当声明类型断言失败时返回
	ErrInvalidClaims = errors.New("jwt: invalid claims type")

	// ErrTokenRevoked 当令牌已被撤销时返回
	ErrTokenRevoked = errors.New("jwt: token has been revoked")

	// ErrEmptyTokenID 当令牌 ID 为空时返回
	ErrEmptyTokenID = errors.New("jwt: token ID cannot be empty")

	// ErrUnexpectedSigningMethod 当令牌使用意外的签名方法时返回
	ErrUnexpectedSigningMethod = errors.New("jwt: unexpected signing method")

	// ErrNilJWT 当 TokenValidator 未配置 JWT 实例时返回
	ErrNilJWT = errors.New("jwt: JWT instance cannot be nil")
)
