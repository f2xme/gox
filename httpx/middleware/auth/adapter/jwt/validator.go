package jwt

import (
	"context"

	"github.com/f2xme/gox/httpx/middleware/auth"
	goxjwt "github.com/f2xme/gox/jwt"
)

// TokenValidator 将 gox/jwt.JWT 适配为 auth.Validator。
type TokenValidator struct {
	JWT goxjwt.JWT
}

var _ auth.Validator = (*TokenValidator)(nil)

// NewTokenValidator 使用给定的 JWT 实例创建 TokenValidator。
func NewTokenValidator(j goxjwt.JWT) *TokenValidator {
	return &TokenValidator{JWT: j}
}

// NewHS256Validator 使用 HS256 算法和指定密钥创建 TokenValidator。
func NewHS256Validator(secret []byte, opts ...goxjwt.Option) *TokenValidator {
	return NewTokenValidator(goxjwt.NewHS256(secret, opts...))
}

// NewHS384Validator 使用 HS384 算法和指定密钥创建 TokenValidator。
func NewHS384Validator(secret []byte, opts ...goxjwt.Option) *TokenValidator {
	return NewTokenValidator(goxjwt.NewHS384(secret, opts...))
}

// NewHS512Validator 使用 HS512 算法和指定密钥创建 TokenValidator。
func NewHS512Validator(secret []byte, opts ...goxjwt.Option) *TokenValidator {
	return NewTokenValidator(goxjwt.NewHS512(secret, opts...))
}

// Validate 解析并校验 token，返回可用于 auth 中间件的声明。
func (v *TokenValidator) Validate(_ context.Context, token string) (auth.Claims, error) {
	if v == nil || v.JWT == nil {
		return nil, goxjwt.ErrNilJWT
	}
	c, err := v.JWT.Parse(token)
	if err != nil {
		return nil, err
	}
	return c, nil
}
