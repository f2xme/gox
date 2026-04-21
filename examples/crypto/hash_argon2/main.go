package main

import (
	"fmt"
	"log"

	"github.com/f2xme/gox/crypto"
)

func main() {
	fmt.Println("=== Argon2id 密码哈希示例 ===")

	password := "mySecurePassword123"

	// 使用默认参数哈希密码
	fmt.Println("\n1. 使用默认参数:")
	hash, err := crypto.HashPasswordArgon2(password, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("原始密码: %s\n", password)
	fmt.Printf("Argon2id 哈希: %s\n", hash)

	// 验证密码
	valid, err := crypto.VerifyPasswordArgon2(password, hash)
	if err != nil {
		log.Fatal(err)
	}
	if valid {
		fmt.Println("✓ 密码验证成功")
	} else {
		fmt.Println("✗ 密码验证失败")
	}

	// 使用自定义参数
	fmt.Println("\n2. 使用自定义参数（更高安全性）:")
	params := &crypto.Argon2Params{
		Memory:      128 * 1024, // 128 MB
		Iterations:  4,          // 4 次迭代
		Parallelism: 4,          // 4 个并行线程
		SaltLength:  16,         // 16 字节盐
		KeyLength:   32,         // 32 字节密钥
	}
	hash2, err := crypto.HashPasswordArgon2(password, params)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("自定义参数哈希: %s\n", hash2)

	valid2, err := crypto.VerifyPasswordArgon2(password, hash2)
	if err != nil {
		log.Fatal(err)
	}
	if valid2 {
		fmt.Println("✓ 密码验证成功")
	} else {
		fmt.Println("✗ 密码验证失败")
	}

	// 验证错误的密码
	fmt.Println("\n3. 验证错误密码:")
	wrongPassword := "wrongPassword"
	valid3, err := crypto.VerifyPasswordArgon2(wrongPassword, hash)
	if err != nil {
		log.Fatal(err)
	}
	if valid3 {
		fmt.Println("✓ 错误密码验证成功（不应该发生）")
	} else {
		fmt.Println("✗ 错误密码验证失败（预期行为）")
	}
}
