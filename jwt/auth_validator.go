package jwt

import "github.com/f2xme/gox/httpx/middleware/auth"

// TokenValidator 实现 httpx/middleware/auth.TokenValidator，将 JWT 解析为 *Claims（同时满足 auth.Claims）。
type TokenValidator struct {
	JWT JWT
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
