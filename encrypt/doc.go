// Package encrypt 提供加密相关的工具函数，包括哈希、对称加密和非对称加密。
//
// encrypt 包提供一组简单的函数式 API，涵盖常用的加密操作。
//
// # 功能特性
//
//   - 哈希函数：支持 MD5、SHA1、SHA256、SHA512 和 Blake3
//   - 对称加密：AES-GCM 模式（AES-128/192/256）
//   - 非对称加密：RSA OAEP 和 PKCS1v15
//   - 编码工具：Base64 和 Hex 编码/解码
//   - 线程安全：所有函数可在多个 goroutine 中并发使用
//
// # 快速开始
//
// 哈希计算：
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/f2xme/gox/encrypt"
//	)
//
//	func main() {
//		// SHA256 哈希
//		hash := encrypt.SHA256([]byte("hello"))
//		fmt.Println(hash)
//		// 输出: 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
//
//		// Blake3 哈希（更快更安全）
//		hash = encrypt.Blake3([]byte("hello"))
//		fmt.Println(hash)
//	}
//
// AES 对称加密：
//
//	package main
//
//	import (
//		"crypto/rand"
//		"fmt"
//		"github.com/f2xme/gox/encrypt"
//	)
//
//	func main() {
//		// 生成 AES-256 密钥
//		key := make([]byte, 32)
//		rand.Read(key)
//
//		// 加密
//		plaintext := []byte("secret message")
//		ciphertext, err := encrypt.AESEncrypt(plaintext, key)
//		if err != nil {
//			panic(err)
//		}
//
//		// 解密
//		decrypted, err := encrypt.AESDecrypt(ciphertext, key)
//		if err != nil {
//			panic(err)
//		}
//		fmt.Println(string(decrypted)) // 输出: secret message
//	}
//
// RSA 非对称加密：
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/f2xme/gox/encrypt"
//	)
//
//	func main() {
//		// 生成 RSA 密钥对
//		pubKey, privKey, err := encrypt.GenerateRSAKeyPair(2048)
//		if err != nil {
//			panic(err)
//		}
//
//		// 加密
//		plaintext := []byte("secret")
//		ciphertext, err := encrypt.RSAEncrypt(plaintext, pubKey)
//		if err != nil {
//			panic(err)
//		}
//
//		// 解密
//		decrypted, err := encrypt.RSADecrypt(ciphertext, privKey)
//		if err != nil {
//			panic(err)
//		}
//		fmt.Println(string(decrypted)) // 输出: secret
//	}
//
// # 编码工具
//
// Base64 和 Hex 编码可与加密函数自由组合：
//
//	// Base64 编码密文
//	encoded := encrypt.EncodeBase64(ciphertext)
//
//	// Base64 解码
//	decoded, err := encrypt.DecodeBase64(encoded)
//
//	// Hex 编码
//	hexStr := encrypt.EncodeHex(data)
//
//	// Hex 解码
//	data, err := encrypt.DecodeHex(hexStr)
//
// # 最佳实践
//
// 选择合适的哈希算法：
//
//	// 推荐：SHA256 或 Blake3
//	hash := encrypt.SHA256(data)
//	hash := encrypt.Blake3(data)
//
//	// 不推荐：MD5 或 SHA1（已不安全，仅用于兼容性）
//	hash := encrypt.MD5(data)
//
// 使用足够长的密钥：
//
//	// AES-256（推荐）
//	key := make([]byte, 32)
//	rand.Read(key)
//
//	// RSA-2048 或更高
//	pubKey, privKey, _ := encrypt.GenerateRSAKeyPair(2048)
//
// 安全存储密钥：
//
//	// 不要硬编码密钥
//	// 使用环境变量或密钥管理服务
//	key := []byte(os.Getenv("ENCRYPTION_KEY"))
//
// 使用 OAEP 而非 PKCS1v15：
//
//	// 推荐：OAEP（更安全）
//	ciphertext, _ := encrypt.RSAEncrypt(data, pubKey)
//
//	// 不推荐：PKCS1v15（已过时）
//	ciphertext, _ := encrypt.RSAEncryptPKCS1v15(data, pubKey)
//
// # 注意事项
//
//   - MD5 和 SHA1 已在密码学上被破解，仅用于兼容旧系统
//   - RSA PKCS1v15 已过时，新应用应使用 OAEP
//   - AES 密钥必须是 16、24 或 32 字节（对应 AES-128/192/256）
//   - RSA 密钥建议至少 2048 位
//   - 所有函数都是线程安全的
package encrypt
