package jwt

import (
	"time"
)

// JWT 定义 JWT 令牌操作接口
type JWT interface {
	// Generate 使用给定的声明创建新的 JWT 令牌
	Generate(claims *Claims) (string, error)

	// Parse 解析并验证 JWT 令牌字符串
	// 如果令牌有效，返回声明
	Parse(token string) (*Claims, error)

	// Refresh 生成具有更新过期时间的新令牌
	// 主体和其他声明从原始令牌中保留
	Refresh(token string, expiration time.Duration) (string, error)
}
