package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

// GenerateRSAKeyPair 生成指定位数的 RSA 密钥对。
//
// 参数：
//   - bits: 密钥位数，至少 2048 位
//
// 返回值：
//   - *rsa.PrivateKey: RSA 私钥
//   - error: 生成失败时返回错误
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, error) {
	if bits < 2048 {
		return nil, fmt.Errorf("RSA key size must be at least 2048 bits")
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}
	return privateKey, nil
}

// SignRSA 使用 RSA-PSS 和 SHA-256 对数据进行签名。
//
// 参数：
//   - privateKey: RSA 私钥
//   - data: 待签名的数据
//
// 返回值：
//   - []byte: 签名字节
//   - error: 签名失败时返回错误
func SignRSA(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	signature, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, hash[:], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	return signature, nil
}

// VerifyRSA 使用 RSA-PSS 和 SHA-256 验证签名。
//
// 参数：
//   - publicKey: RSA 公钥
//   - data: 原始数据
//   - signature: 签名字节
//
// 返回值：
//   - error: 签名无效时返回错误，有效时返回 nil
func VerifyRSA(publicKey *rsa.PublicKey, data, signature []byte) error {
	hash := sha256.Sum256(data)
	err := rsa.VerifyPSS(publicKey, crypto.SHA256, hash[:], signature, nil)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	return nil
}

// GenerateECDSAKeyPair 生成 ECDSA 密钥对，使用 P-256 曲线。
//
// 返回值：
//   - *ecdsa.PrivateKey: ECDSA 私钥
//   - error: 生成失败时返回错误
func GenerateECDSAKeyPair() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
	}
	return privateKey, nil
}

// SignECDSA 使用 ECDSA 和 SHA-256 对数据进行签名。
//
// 参数：
//   - privateKey: ECDSA 私钥
//   - data: 待签名的数据
//
// 返回值：
//   - []byte: ASN.1 DER 编码的签名
//   - error: 签名失败时返回错误
func SignECDSA(privateKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	return signature, nil
}

// VerifyECDSA 使用 ECDSA 和 SHA-256 验证签名。
//
// 参数：
//   - publicKey: ECDSA 公钥
//   - data: 原始数据
//   - signature: ASN.1 DER 编码的签名
//
// 返回值：
//   - bool: 签名有效返回 true，否则返回 false
func VerifyECDSA(publicKey *ecdsa.PublicKey, data, signature []byte) bool {
	hash := sha256.Sum256(data)
	return ecdsa.VerifyASN1(publicKey, hash[:], signature)
}

// GenerateEd25519KeyPair 生成 Ed25519 密钥对。
//
// 返回值：
//   - ed25519.PublicKey: Ed25519 公钥（32 字节）
//   - ed25519.PrivateKey: Ed25519 私钥（64 字节）
//   - error: 生成失败时返回错误
func GenerateEd25519KeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Ed25519 key: %w", err)
	}
	return publicKey, privateKey, nil
}

// SignEd25519 使用 Ed25519 对数据进行签名。
//
// 参数：
//   - privateKey: Ed25519 私钥
//   - data: 待签名的数据
//
// 返回值：
//   - []byte: 签名字节（64 字节）
func SignEd25519(privateKey ed25519.PrivateKey, data []byte) []byte {
	return ed25519.Sign(privateKey, data)
}

// VerifyEd25519 使用 Ed25519 验证签名。
//
// 参数：
//   - publicKey: Ed25519 公钥
//   - data: 原始数据
//   - signature: 签名字节
//
// 返回值：
//   - bool: 签名有效返回 true，否则返回 false
func VerifyEd25519(publicKey ed25519.PublicKey, data, signature []byte) bool {
	return ed25519.Verify(publicKey, data, signature)
}
