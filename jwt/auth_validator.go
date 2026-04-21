package jwt

import "github.com/f2xme/gox/httpx/middleware/auth"

// TokenValidator 实现 httpx/middleware/auth.TokenValidator，将 JWT 解析为 *Claims（同时满足 auth.Claims）。
type TokenValidator struct {
	JWT JWT
}

// NewTokenValidator 使用给定的 JWT 实例创建 TokenValidator。
func NewTokenValidator(j JWT) *TokenValidator {
	return &TokenValidator{JWT: j}
}

// NewHS256Validator 使用 HS256 算法和指定密钥创建 TokenValidator，是 NewTokenValidator(NewHS256(secret, opts...)) 的快捷方式。
func NewHS256Validator(secret []byte, opts ...Option) *TokenValidator {
	return NewTokenValidator(NewHS256(secret, opts...))
}

// NewHS384Validator 使用 HS384 算法和指定密钥创建 TokenValidator，是 NewTokenValidator(NewHS384(secret, opts...)) 的快捷方式。
func NewHS384Validator(secret []byte, opts ...Option) *TokenValidator {
	return NewTokenValidator(NewHS384(secret, opts...))
}

// NewHS512Validator 使用 HS512 算法和指定密钥创建 TokenValidator，是 NewTokenValidator(NewHS512(secret, opts...)) 的快捷方式。
func NewHS512Validator(secret []byte, opts ...Option) *TokenValidator {
	return NewTokenValidator(NewHS512(secret, opts...))
}

// Validate 解析并校验 token，返回可用于 httpx 认证上下文的声明。
func (v *TokenValidator) Validate(token string) (auth.Claims, error) {
	if v == nil || v.JWT == nil {
		return nil, ErrNilJWT
	}
	c, err := v.JWT.Parse(token)
	if err != nil {
		return nil, err
	}
	return c, nil
}
