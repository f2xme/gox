package main

import (
	"fmt"
	"log"

	"github.com/f2xme/gox/crypto"
)

func main() {
	fmt.Println("=== bcrypt 密码哈希示例 ===")

	password := "mySecurePassword123"

	// 哈希密码
	hash, err := crypto.HashPassword(password)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("原始密码: %s\n", password)
	fmt.Printf("bcrypt 哈希: %s\n", hash)

	// 验证正确的密码
	if crypto.VerifyPassword(password, hash) {
		fmt.Println("✓ 密码验证成功")
	} else {
		fmt.Println("✗ 密码验证失败")
	}

	// 验证错误的密码
	wrongPassword := "wrongPassword"
	if crypto.VerifyPassword(wrongPassword, hash) {
		fmt.Println("✓ 错误密码验证成功（不应该发生）")
	} else {
		fmt.Println("✗ 错误密码验证失败（预期行为）")
	}
}
