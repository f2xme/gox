package main

import (
	"crypto/rand"
	"fmt"
	"log"

	"github.com/f2xme/gox/encrypt"
)

func main() {
	fmt.Println("=== encrypt 包使用示例 ===")

	// 示例 1: 哈希函数
	fmt.Println("\n示例 1: 哈希函数")
	data := []byte("hello world")

	fmt.Printf("MD5:    %s\n", encrypt.MD5(data))
	fmt.Printf("SHA1:   %s\n", encrypt.SHA1(data))
	fmt.Printf("SHA256: %s\n", encrypt.SHA256(data))
	fmt.Printf("SHA512: %s\n", encrypt.SHA512(data))
	fmt.Printf("Blake3: %s\n", encrypt.Blake3(data))

	// 示例 2: AES 对称加密
	fmt.Println("\n示例 2: AES 对称加密")

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatal(err)
	}

	plaintext := []byte("这是需要加密的敏感数据")
	fmt.Printf("原文: %s\n", plaintext)

	ciphertext, err := encrypt.AESEncrypt(plaintext, key)
	if err != nil {
		log.Fatalf("加密失败: %v", err)
	}
	fmt.Printf("密文长度: %d 字节\n", len(ciphertext))

	// 解密
	decrypted, err := encrypt.AESDecrypt(ciphertext, key)
	if err != nil {
		log.Fatalf("解密失败: %v", err)
	}
	fmt.Printf("解密后: %s\n", decrypted)

	// 示例 3: RSA 非对称加密
	fmt.Println("\n示例 3: RSA 非对称加密")

	publicKey, privateKey, err := encrypt.GenerateRSAKeyPair(2048)
	if err != nil {
		log.Fatalf("生成密钥对失败: %v", err)
	}

	fmt.Println("公钥（前 100 字符）:")
	fmt.Println(publicKey[:100] + "...")

	message := []byte("RSA 加密的消息")
	fmt.Printf("\n原文: %s\n", message)

	encrypted, err := encrypt.RSAEncrypt(message, publicKey)
	if err != nil {
		log.Fatalf("RSA 加密失败: %v", err)
	}
	fmt.Printf("密文长度: %d 字节\n", len(encrypted))

	decryptedMsg, err := encrypt.RSADecrypt(encrypted, privateKey)
	if err != nil {
		log.Fatalf("RSA 解密失败: %v", err)
	}
	fmt.Printf("解密后: %s\n", decryptedMsg)

	// 示例 4: Base64 编码
	fmt.Println("\n示例 4: Base64 编码")

	original := []byte("需要编码的数据")
	encoded := encrypt.EncodeBase64(original)
	fmt.Printf("原文: %s\n", original)
	fmt.Printf("Base64: %s\n", encoded)

	decoded, err := encrypt.DecodeBase64(encoded)
	if err != nil {
		log.Fatalf("Base64 解码失败: %v", err)
	}
	fmt.Printf("解码后: %s\n", decoded)

	// 示例 5: Hex 编码
	fmt.Println("\n示例 5: Hex 编码")

	hexEncoded := encrypt.EncodeHex(original)
	fmt.Printf("原文: %s\n", original)
	fmt.Printf("Hex: %s\n", hexEncoded)

	hexDecoded, err := encrypt.DecodeHex(hexEncoded)
	if err != nil {
		log.Fatalf("Hex 解码失败: %v", err)
	}
	fmt.Printf("解码后: %s\n", hexDecoded)

	// 示例 6: 组合使用 - AES 加密 + Base64 编码
	fmt.Println("\n示例 6: AES 加密 + Base64 编码")

	secret := []byte("机密信息")
	fmt.Printf("原文: %s\n", secret)

	encryptedData, err := encrypt.AESEncrypt(secret, key)
	if err != nil {
		log.Fatalf("加密失败: %v", err)
	}

	encodedData := encrypt.EncodeBase64(encryptedData)
	fmt.Printf("加密并编码后: %s\n", encodedData)

	decodedData, err := encrypt.DecodeBase64(encodedData)
	if err != nil {
		log.Fatalf("解码失败: %v", err)
	}

	finalData, err := encrypt.AESDecrypt(decodedData, key)
	if err != nil {
		log.Fatalf("解密失败: %v", err)
	}
	fmt.Printf("解码并解密后: %s\n", finalData)

	// 示例 7: 密码哈希（推荐使用 SHA256 或 Blake3）
	fmt.Println("\n示例 7: 密码哈希")

	password := []byte("user_password_123")
	fmt.Printf("密码: %s\n", password)

	// 使用 SHA256 哈希
	hashedPassword := encrypt.SHA256(password)
	fmt.Printf("SHA256 哈希: %s\n", hashedPassword)

	// 使用 Blake3 哈希（更快）
	blake3Hash := encrypt.Blake3(password)
	fmt.Printf("Blake3 哈希: %s\n", blake3Hash)

	fmt.Println("\n=== 示例结束 ===")
}
