package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/f2xme/gox/crypto"
)

func main() {
	fmt.Println("=== ECDSA 数字签名示例 ===")

	// 生成 ECDSA 密钥对（P-256 曲线）
	fmt.Println("生成 ECDSA 密钥对（P-256 曲线）...")
	privateKey, err := crypto.GenerateECDSAKeyPair()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("密钥生成成功\n")
	fmt.Printf("曲线: %s\n", privateKey.Curve.Params().Name)

	// 待签名的数据
	data := []byte("这是一条需要签名的重要消息")
	fmt.Printf("\n原始数据: %s\n", string(data))

	// 签名
	signature, err := crypto.SignECDSA(privateKey, data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("签名长度: %d 字节\n", len(signature))
	fmt.Printf("签名 (base64): %s\n", base64.StdEncoding.EncodeToString(signature))

	// 验证签名
	fmt.Println("\n验证签名:")
	if crypto.VerifyECDSA(&privateKey.PublicKey, data, signature) {
		fmt.Println("✓ 签名验证成功")
	} else {
		fmt.Println("✗ 签名验证失败")
	}

	// 验证被篡改的数据
	fmt.Println("\n验证被篡改的数据:")
	tamperedData := []byte("这是一条被篡改的消息")
	if crypto.VerifyECDSA(&privateKey.PublicKey, tamperedData, signature) {
		fmt.Println("✓ 签名验证成功（不应该发生）")
	} else {
		fmt.Println("✗ 签名验证失败（预期行为）")
	}

	// 验证错误的签名
	fmt.Println("\n验证错误的签名:")
	wrongSignature := make([]byte, len(signature))
	if crypto.VerifyECDSA(&privateKey.PublicKey, data, wrongSignature) {
		fmt.Println("✓ 签名验证成功（不应该发生）")
	} else {
		fmt.Println("✗ 签名验证失败（预期行为）")
	}
}
