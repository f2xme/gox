package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/f2xme/gox/crypto"
)

func main() {
	fmt.Println("=== RSA 数字签名示例 ===")

	// 生成 2048 位 RSA 密钥对
	fmt.Println("生成 2048 位 RSA 密钥对...")
	privateKey, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("密钥生成成功\n")
	fmt.Printf("公钥模数位数: %d\n", privateKey.PublicKey.N.BitLen())

	// 待签名的数据
	data := []byte("这是一条需要签名的重要消息")
	fmt.Printf("\n原始数据: %s\n", string(data))

	// 签名
	signature, err := crypto.SignRSA(privateKey, data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("签名长度: %d 字节\n", len(signature))
	fmt.Printf("签名 (base64): %s...\n", base64.StdEncoding.EncodeToString(signature)[:64])

	// 验证签名
	fmt.Println("\n验证签名:")
	err = crypto.VerifyRSA(&privateKey.PublicKey, data, signature)
	if err != nil {
		fmt.Printf("✗ 签名验证失败: %v\n", err)
	} else {
		fmt.Println("✓ 签名验证成功")
	}

	// 验证被篡改的数据
	fmt.Println("\n验证被篡改的数据:")
	tamperedData := []byte("这是一条被篡改的消息")
	err = crypto.VerifyRSA(&privateKey.PublicKey, tamperedData, signature)
	if err != nil {
		fmt.Println("✗ 签名验证失败（预期行为）")
	} else {
		fmt.Println("✓ 签名验证成功（不应该发生）")
	}

	// 验证错误的签名
	fmt.Println("\n验证错误的签名:")
	wrongSignature := make([]byte, len(signature))
	err = crypto.VerifyRSA(&privateKey.PublicKey, data, wrongSignature)
	if err != nil {
		fmt.Println("✗ 签名验证失败（预期行为）")
	} else {
		fmt.Println("✓ 签名验证成功（不应该发生）")
	}
}
