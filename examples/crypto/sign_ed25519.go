package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/f2xme/gox/crypto"
)

func main() {
	fmt.Println("=== Ed25519 数字签名示例 ===")

	// 生成密钥对
	publicKey, privateKey, err := crypto.GenerateEd25519KeyPair()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("公钥长度: %d 字节\n", len(publicKey))
	fmt.Printf("私钥长度: %d 字节\n", len(privateKey))
	fmt.Printf("公钥 (base64): %s\n", base64.StdEncoding.EncodeToString(publicKey))

	// 待签名的数据
	data := []byte("这是一条需要签名的重要消息")
	fmt.Printf("\n原始数据: %s\n", string(data))

	// 签名
	signature := crypto.SignEd25519(privateKey, data)
	fmt.Printf("签名长度: %d 字节\n", len(signature))
	fmt.Printf("签名 (base64): %s\n", base64.StdEncoding.EncodeToString(signature))

	// 验证签名
	fmt.Println("\n验证签名:")
	if crypto.VerifyEd25519(publicKey, data, signature) {
		fmt.Println("✓ 签名验证成功")
	} else {
		fmt.Println("✗ 签名验证失败")
	}

	// 验证被篡改的数据
	fmt.Println("\n验证被篡改的数据:")
	tamperedData := []byte("这是一条被篡改的消息")
	if crypto.VerifyEd25519(publicKey, tamperedData, signature) {
		fmt.Println("✓ 签名验证成功（不应该发生）")
	} else {
		fmt.Println("✗ 签名验证失败（预期行为）")
	}

	// 验证错误的签名
	fmt.Println("\n验证错误的签名:")
	wrongSignature := make([]byte, 64)
	if crypto.VerifyEd25519(publicKey, data, wrongSignature) {
		fmt.Println("✓ 签名验证成功（不应该发生）")
	} else {
		fmt.Println("✗ 签名验证失败（预期行为）")
	}
}
