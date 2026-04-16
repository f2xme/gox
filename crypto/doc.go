// Package crypto 提供密码学工具，包括密码哈希和数字签名功能。
//
// # 功能特性
//
//   - 密码哈希：支持 bcrypt 和 Argon2id 算法
//   - 数字签名：支持 RSA、ECDSA 和 Ed25519 算法
//   - 安全的随机数生成和密钥管理
//   - 符合行业标准的加密参数配置
//
// # 快速开始
//
// 密码哈希 - bcrypt：
//
//	package main
//
//	import (
//		"fmt"
//		"log"
//
//		"github.com/f2xme/gox/crypto"
//	)
//
//	func main() {
//		// 哈希密码
//		hash, err := crypto.HashPassword("mypassword")
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// 验证密码
//		if crypto.VerifyPassword("mypassword", hash) {
//			fmt.Println("密码正确")
//		}
//	}
//
// 密码哈希 - Argon2id（推荐）：
//
//	// 使用默认参数
//	hash, err := crypto.HashPasswordArgon2("mypassword", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 验证密码
//	valid, err := crypto.VerifyPasswordArgon2("mypassword", hash)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if valid {
//		fmt.Println("密码正确")
//	}
//
// 数字签名 - Ed25519（推荐）：
//
//	// 生成密钥对
//	publicKey, privateKey, err := crypto.GenerateEd25519KeyPair()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	data := []byte("message to sign")
//
//	// 签名
//	signature := crypto.SignEd25519(privateKey, data)
//
//	// 验证
//	if crypto.VerifyEd25519(publicKey, data, signature) {
//		fmt.Println("签名有效")
//	}
//
// # 密码哈希算法选择
//
// bcrypt：
//
//   - 适用于需要兼容现有系统的场景
//   - 计算成本固定，无法调整内存使用
//   - 广泛支持，成熟稳定
//
// Argon2id（推荐用于新项目）：
//
//   - 2015 年密码哈希竞赛冠军
//   - 可配置内存和 CPU 成本，抵御 GPU/ASIC 攻击
//   - 结合了 Argon2i（抗侧信道攻击）和 Argon2d（抗 GPU 攻击）的优点
//
// 自定义 Argon2id 参数：
//
//	params := &crypto.Argon2Params{
//		Memory:      128 * 1024, // 128 MB
//		Iterations:  4,           // 迭代次数
//		Parallelism: 4,           // 并行度
//		SaltLength:  16,          // 盐长度
//		KeyLength:   32,          // 密钥长度
//	}
//	hash, err := crypto.HashPasswordArgon2("mypassword", params)
//
// # 数字签名算法选择
//
// RSA：
//
//   - 密钥长度至少 2048 位（推荐 3072 或 4096 位）
//   - 签名和验证速度较慢
//   - 密钥和签名体积较大
//   - 适用于需要与传统系统兼容的场景
//
// ECDSA：
//
//   - 使用 P-256 曲线（NIST 标准）
//   - 密钥和签名体积小
//   - 签名速度快，验证速度中等
//   - 广泛应用于 TLS、JWT 等场景
//
// Ed25519（推荐用于新项目）：
//
//   - 基于 Curve25519 椭圆曲线
//   - 签名和验证速度最快
//   - 密钥和签名体积最小（公钥 32 字节，签名 64 字节）
//   - 抗侧信道攻击，无需随机数生成器
//   - 适用于高性能场景和资源受限设备
//
// # 数字签名示例
//
// RSA 签名：
//
//	// 生成 2048 位密钥对
//	privateKey, err := crypto.GenerateRSAKeyPair(2048)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	data := []byte("message to sign")
//
//	// 签名
//	signature, err := crypto.SignRSA(privateKey, data)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 验证
//	err = crypto.VerifyRSA(&privateKey.PublicKey, data, signature)
//	if err != nil {
//		fmt.Println("签名无效:", err)
//	} else {
//		fmt.Println("签名有效")
//	}
//
// ECDSA 签名：
//
//	// 生成密钥对（P-256 曲线）
//	privateKey, err := crypto.GenerateECDSAKeyPair()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	data := []byte("message to sign")
//
//	// 签名
//	signature, err := crypto.SignECDSA(privateKey, data)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 验证
//	if crypto.VerifyECDSA(&privateKey.PublicKey, data, signature) {
//		fmt.Println("签名有效")
//	}
//
// # 安全建议
//
// 密码哈希：
//
//   - 永远不要存储明文密码
//   - 使用足够的计算成本参数以抵御暴力破解
//   - Argon2id 推荐参数：内存 64MB+，迭代 3+，并行度 2+
//   - bcrypt 推荐使用默认 cost（10）或更高
//   - 定期审查和更新哈希参数以应对硬件性能提升
//
// 数字签名：
//
//   - 妥善保管私钥，永远不要泄露或传输
//   - RSA 密钥长度至少 2048 位，推荐 3072 或 4096 位
//   - 使用安全的随机数生成器（本包已使用 crypto/rand）
//   - 验证签名时始终检查返回的错误
//   - 考虑使用密钥轮换策略
//
// # 注意事项
//
//   - 所有密码哈希函数都会自动生成随机盐
//   - Argon2id 哈希格式：$argon2id$v=19$m=65536,t=3,p=2$salt$hash
//   - bcrypt 哈希包含算法版本、cost 和盐信息
//   - 数字签名使用 SHA-256 作为哈希函数（Ed25519 除外，它内置哈希）
//   - RSA 签名使用 PSS 填充模式（比 PKCS#1 v1.5 更安全）
//   - ECDSA 签名使用 ASN.1 DER 编码格式
package crypto
