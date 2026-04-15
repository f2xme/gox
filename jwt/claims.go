package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var standardClaimKeys = map[string]bool{
	"sub": true, "iss": true, "aud": true, "exp": true,
	"nbf": true, "iat": true, "jti": true,
}

// Claims 表示包含标准字段和自定义字段的 JWT 声明
type Claims struct {
	// 标准声明
	Subject   string    `json:"sub,omitempty"` // 主体（通常是用户 ID）
	Issuer    string    `json:"iss,omitempty"` // 签发者
	Audience  []string  `json:"aud,omitempty"` // 受众
	ExpiresAt time.Time `json:"exp,omitempty"` // 过期时间
	NotBefore time.Time `json:"nbf,omitempty"` // 生效时间
	IssuedAt  time.Time `json:"iat,omitempty"` // 签发时间
	ID        string    `json:"jti,omitempty"` // JWT ID

	// 自定义声明
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// ToMapClaims 将 Claims 转换为 jwt.MapClaims 用于令牌生成
func (c *Claims) ToMapClaims() jwt.MapClaims {
	claims := make(jwt.MapClaims, 7+len(c.Custom))

	if c.Subject != "" {
		claims["sub"] = c.Subject
	}
	if c.Issuer != "" {
		claims["iss"] = c.Issuer
	}
	if len(c.Audience) > 0 {
		claims["aud"] = c.Audience
	}
	if !c.ExpiresAt.IsZero() {
		claims["exp"] = c.ExpiresAt.Unix()
	}
	if !c.NotBefore.IsZero() {
		claims["nbf"] = c.NotBefore.Unix()
	}
	if !c.IssuedAt.IsZero() {
		claims["iat"] = c.IssuedAt.Unix()
	}
	if c.ID != "" {
		claims["jti"] = c.ID
	}

	// Add custom claims
	for k, v := range c.Custom {
		claims[k] = v
	}

	return claims
}

// FromMapClaims 将 jwt.MapClaims 转换为 Claims
func FromMapClaims(mc jwt.MapClaims) *Claims {
	claims := &Claims{}

	if sub, ok := mc["sub"].(string); ok {
		claims.Subject = sub
	}
	if iss, ok := mc["iss"].(string); ok {
		claims.Issuer = iss
	}
	// Handle audience - can be []interface{}, []string, or string
	switch aud := mc["aud"].(type) {
	case []interface{}:
		claims.Audience = make([]string, 0, len(aud))
		for _, a := range aud {
			if s, ok := a.(string); ok {
				claims.Audience = append(claims.Audience, s)
			}
		}
	case []string:
		claims.Audience = aud
	case string:
		claims.Audience = []string{aud}
	}
	if exp, ok := mc["exp"].(float64); ok {
		claims.ExpiresAt = time.Unix(int64(exp), 0)
	}
	if nbf, ok := mc["nbf"].(float64); ok {
		claims.NotBefore = time.Unix(int64(nbf), 0)
	}
	if iat, ok := mc["iat"].(float64); ok {
		claims.IssuedAt = time.Unix(int64(iat), 0)
	}
	if jti, ok := mc["jti"].(string); ok {
		claims.ID = jti
	}

	// Extract custom claims (only allocate map if needed)
	for k, v := range mc {
		if !standardClaimKeys[k] {
			if claims.Custom == nil {
				claims.Custom = make(map[string]interface{})
			}
			claims.Custom[k] = v
		}
	}

	return claims
}

// IsExpired 检查令牌是否已过期
func (c *Claims) IsExpired() bool {
	if c.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(c.ExpiresAt)
}

// IsNotYetValid 检查令牌是否尚未生效
func (c *Claims) IsNotYetValid() bool {
	if c.NotBefore.IsZero() {
		return false
	}
	return time.Now().Before(c.NotBefore)
}
