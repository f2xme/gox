// Package jwt 提供 JWT 令牌的生成、解析和验证功能。
//
// 基于 golang-jwt/jwt 封装，支持多种签名算法（HMAC、RSA、ECDSA），
// 提供统一的 API 和开箱即用的配置。
//
// # 功能特性
//
//   - 支持多种签名算法：HMAC (HS256/384/512)、RSA (RS256/384/512)、ECDSA (ES256/384/512)
//   - 令牌生成、解析和验证
//   - 令牌刷新（延长过期时间）
//   - 令牌撤销（基于 Redis）
//   - 自定义声明支持
//   - 线程安全，支持并发使用
//
// # 快速开始
//
// 基本使用（HMAC 算法）：
//
//	package main
//
//	import (
//		"fmt"
//		"time"
//
//		"github.com/f2xme/gox/jwt"
//	)
//
//	func main() {
//		// 创建 JWT 实例
//		j := jwt.NewHS256([]byte("your-secret-key"))
//
//		// 生成令牌
//		token, err := j.Generate(&jwt.Claims{
//			Subject:   "user123",
//			ExpiresAt: time.Now().Add(24 * time.Hour),
//			Custom: map[string]interface{}{
//				"role": "admin",
//			},
//		})
//		if err != nil {
//			panic(err)
//		}
//		fmt.Println("Token:", token)
//
//		// 解析令牌
//		claims, err := j.Parse(token)
//		if err != nil {
//			panic(err)
//		}
//		fmt.Println("Subject:", claims.Subject)
//		fmt.Println("Role:", claims.Custom["role"])
//	}
//
// # 使用配置选项
//
// 使用 Options 模式配置 JWT：
//
//	j := jwt.NewHS256(
//		[]byte("secret"),
//		jwt.WithIssuer("my-app"),
//		jwt.WithExpiration(2*time.Hour),
//		jwt.WithAudience("web", "mobile"),
//	)
//
//	// 生成令牌时会自动应用配置的默认值
//	token, _ := j.Generate(&jwt.Claims{
//		Subject: "user123",
//		// Issuer、ExpiresAt、Audience 会自动填充
//	})
//
// # RSA 签名
//
// 适用于分布式系统，公钥验证无需共享密钥：
//
//	import (
//		"crypto/rsa"
//		"crypto/x509"
//		"encoding/pem"
//		"os"
//	)
//
//	// 加载私钥
//	privateKeyData, _ := os.ReadFile("private.pem")
//	block, _ := pem.Decode(privateKeyData)
//	privateKey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
//
//	// 加载公钥
//	publicKeyData, _ := os.ReadFile("public.pem")
//	block, _ = pem.Decode(publicKeyData)
//	publicKey, _ := x509.ParsePKCS1PublicKey(block.Bytes)
//
//	// 创建 RSA JWT
//	j := jwt.NewRS256(privateKey, publicKey)
//
//	token, _ := j.Generate(claims)
//	claims, _ := j.Parse(token)
//
// # ECDSA 签名
//
// 适用于移动应用，密钥短、签名快：
//
//	import (
//		"crypto/ecdsa"
//		"crypto/elliptic"
//		"crypto/rand"
//	)
//
//	// 生成密钥对
//	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
//	publicKey := &privateKey.PublicKey
//
//	// 创建 ECDSA JWT
//	j := jwt.NewES256(privateKey, publicKey)
//
//	token, _ := j.Generate(claims)
//	claims, _ := j.Parse(token)
//
// # 令牌刷新
//
// 延长令牌的过期时间：
//
//	// 刷新令牌，延长 24 小时
//	newToken, err := j.Refresh(oldToken, 24*time.Hour)
//	if err != nil {
//		panic(err)
//	}
//
// # 令牌撤销
//
// 使用 Redis 实现令牌撤销（用于登出场景）：
//
//	import (
//		"context"
//
//		"github.com/f2xme/gox/jwt"
//		"github.com/redis/go-redis/v9"
//	)
//
//	// 创建 Redis 客户端
//	rdb := redis.NewClient(&redis.Options{
//		Addr: "localhost:6379",
//	})
//
//	// 创建撤销器
//	revoker := jwt.NewRedisRevoker(rdb)
//
//	// 创建 JWT 时配置撤销器
//	j := jwt.NewHS256(
//		[]byte("secret"),
//		jwt.WithRevoker(revoker),
//	)
//
//	// 撤销令牌（用户登出）
//	ctx := context.Background()
//	err := revoker.Revoke(ctx, claims.ID, time.Until(claims.ExpiresAt))
//
//	// 解析时会自动检查令牌是否被撤销
//	claims, err := j.Parse(token)
//	// 如果令牌已被撤销，会返回 jwt.ErrTokenRevoked
//
// # 算法选择指南
//
//	| 场景           | 推荐算法          | 原因                     |
//	|----------------|-------------------|--------------------------|
//	| 单体应用       | HMAC (HS256)      | 简单高效，密钥管理容易   |
//	| 分布式系统     | RSA (RS256)       | 公钥验证，无需共享密钥   |
//	| 移动应用       | ECDSA (ES256)     | 密钥短，签名快           |
//	| 高安全要求     | RSA/ECDSA 384/512 | 更长的密钥长度           |
//
// # 最佳实践
//
// 1. 设置合理的过期时间
//
//	// 访问令牌：短期（15 分钟 - 1 小时）
//	accessToken, _ := j.Generate(&jwt.Claims{
//		Subject:   userID,
//		ExpiresAt: time.Now().Add(15 * time.Minute),
//	})
//
//	// 刷新令牌：长期（7 天 - 30 天）
//	refreshToken, _ := j.Generate(&jwt.Claims{
//		Subject:   userID,
//		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
//		Custom:    map[string]interface{}{"type": "refresh"},
//	})
//
// 2. 使用 HTTPS 传输令牌
//
//	// 在 HTTP 头中传输
//	req.Header.Set("Authorization", "Bearer "+token)
//
// 3. 验证令牌后检查声明
//
//	claims, err := j.Parse(token)
//	if err != nil {
//		return err
//	}
//
//	// 检查过期时间
//	if claims.IsExpired() {
//		return errors.New("token expired")
//	}
//
//	// 检查自定义字段
//	if claims.Custom["role"] != "admin" {
//		return errors.New("insufficient permissions")
//	}
//
// 4. 安全存储密钥
//
//	// 不要硬编码密钥，使用环境变量或密钥管理服务
//	secret := []byte(os.Getenv("JWT_SECRET"))
//	j := jwt.NewHS256(secret)
//
// # 错误处理
//
//	token, err := j.Parse(tokenString)
//	if err != nil {
//		switch {
//		case errors.Is(err, jwt.ErrInvalidToken):
//			// 令牌格式无效
//		case errors.Is(err, jwt.ErrTokenRevoked):
//			// 令牌已被撤销
//		case errors.Is(err, jwt.ErrUnexpectedSigningMethod):
//			// 签名算法不匹配
//		default:
//			// 其他错误
//		}
//	}
//
// # 线程安全
//
// 所有 JWT 实现都是线程安全的，可以在多个 goroutine 中并发使用。
package jwt
