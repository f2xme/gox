// Package random 提供安全的随机字符生成功能。
//
// 基于 crypto/rand 实现，提供密码学安全的随机性，适用于生成验证码、
// 临时密码、API Token 等安全敏感场景。
//
// # 功能特性
//
//   - 密码学安全：使用 crypto/rand 提供密码学级别的随机性
//   - 多种字符集：支持数字、字母、字母数字等预定义字符集
//   - 自定义字符集：可以指定任意字符集生成随机字符串
//   - 并发安全：所有函数都是并发安全的
//   - 零依赖：除标准库外无其他依赖
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"fmt"
//		"log"
//
//		"github.com/f2xme/gox/random"
//	)
//
//	func main() {
//		// 生成 6 位数字验证码
//		code, err := random.Numeric(6)
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Println("验证码:", code) // 例如: 847295
//
//		// 生成 16 位字母数字混合 Token
//		token, err := random.AlphaNumeric(16)
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Println("Token:", token) // 例如: aB3xK9mP2qR7sT4u
//
//		// 生成 8 位小写字母 ID
//		id, err := random.AlphaLower(8)
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Println("ID:", id) // 例如: xkpmqrst
//	}
//
// # 使用场景
//
// 验证码生成：
//
//	// 生成 6 位数字验证码
//	code, _ := random.Numeric(6)
//
// API Token 生成：
//
//	// 生成 32 位字母数字 Token
//	token, _ := random.AlphaNumeric(32)
//
// 临时密码生成：
//
//	// 生成 12 位字母数字混合密码
//	password, _ := random.AlphaNumeric(12)
//
// 自定义字符集：
//
//	// 生成 16 位十六进制字符串
//	hex, _ := random.String(16, "0123456789ABCDEF")
//
//	// 生成 8 位不含易混淆字符的字符串
//	safe, _ := random.String(8, "23456789ABCDEFGHJKLMNPQRSTUVWXYZ")
//
// # 注意事项
//
//   - 所有函数使用 crypto/rand 提供密码学安全的随机性，适用于安全敏感场景
//   - 如果 length <= 0 或 charset 为空，返回空字符串和 nil 错误
//   - 错误通常来自底层的 crypto/rand.Reader，在正常系统上极少发生
//   - 生成的随机字符串不保证唯一性，如需唯一 ID 请使用 idgen 包
package random
