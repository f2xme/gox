package main

import (
	"fmt"
	"time"

	"github.com/f2xme/gox/jwt"
)

func main() {
	fmt.Println("=== JWT 使用示例 ===")

	// 1. 使用 HMAC (HS256) 算法创建 JWT 实例
	secret := []byte("your-secret-key-keep-it-safe")
	jwtInstance := jwt.NewHS256(secret)

	// 2. 使用 Options 模式配置 JWT
	fmt.Println("1. 使用 Options 模式创建 JWT:")
	jwtWithOptions := jwt.NewHS256(
		secret,
		jwt.WithIssuer("gox-example"),
		jwt.WithExpiration(24*time.Hour),
		jwt.WithAudience("web", "mobile"),
	)
	fmt.Println("已创建带配置的 JWT 实例")

	// 3. 生成 JWT token
	fmt.Println("\n2. 生成 JWT Token:")
	claims := &jwt.Claims{
		Subject: "user123",
		Custom: map[string]interface{}{
			"user_id":  12345,
			"username": "zhangsan",
			"role":     "admin",
		},
	}

	token, err := jwtWithOptions.Generate(claims)
	if err != nil {
		fmt.Printf("生成 token 失败: %v\n", err)
		return
	}
	fmt.Printf("Token: %s\n", token)

	// 4. 解析和验证 token
	fmt.Println("\n3. 解析和验证 Token:")
	parsedClaims, err := jwtWithOptions.Parse(token)
	if err != nil {
		fmt.Printf("解析 token 失败: %v\n", err)
		return
	}

	fmt.Printf("Subject: %s\n", parsedClaims.Subject)
	fmt.Printf("Issuer: %s\n", parsedClaims.Issuer)
	fmt.Printf("Audience: %v\n", parsedClaims.Audience)
	fmt.Printf("ExpiresAt: %s\n", parsedClaims.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("Custom Claims:\n")
	for k, v := range parsedClaims.Custom {
		fmt.Printf("  %s: %v\n", k, v)
	}

	// 5. 检查 token 是否过期
	fmt.Println("\n4. 检查 Token 状态:")
	fmt.Printf("Token 是否过期: %v\n", parsedClaims.IsExpired())
	fmt.Printf("Token 是否尚未生效: %v\n", parsedClaims.IsNotYetValid())

	// 6. 刷新 token（延长有效期）
	fmt.Println("\n5. 刷新 Token:")
	newToken, err := jwtInstance.Refresh(token, 48*time.Hour)
	if err != nil {
		fmt.Printf("刷新 token 失败: %v\n", err)
		return
	}
	fmt.Printf("新 Token: %s\n", newToken)

	// 验证新 token
	newClaims, err := jwtInstance.Parse(newToken)
	if err != nil {
		fmt.Printf("解析新 token 失败: %v\n", err)
		return
	}
	fmt.Printf("新 Token 过期时间: %s\n", newClaims.ExpiresAt.Format(time.RFC3339))

	// 7. 演示无效 token 的处理
	fmt.Println("\n6. 错误处理示例:")
	invalidToken := "invalid.token.string"
	_, err = jwtInstance.Parse(invalidToken)
	if err != nil {
		fmt.Printf("预期的错误: %v\n", err)
	}

	// 8. 使用不同的算法（HS512）
	fmt.Println("\n7. 使用 HS512 算法:")
	jwtHS512 := jwt.NewHS512(secret)
	tokenHS512, err := jwtHS512.Generate(&jwt.Claims{
		Subject:   "user456",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		IssuedAt:  time.Now(),
		Custom: map[string]interface{}{
			"level": "premium",
		},
	})
	if err != nil {
		fmt.Printf("生成 HS512 token 失败: %v\n", err)
		return
	}
	fmt.Printf("HS512 Token: %s\n", tokenHS512)

	fmt.Println("\nJWT 示例完成")
}
