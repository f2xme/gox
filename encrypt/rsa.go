// encrypt/rsa.go
package encrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
)

// GenerateRSAKeyPair 生成 RSA 密钥对并返回 PEM 编码的公钥和私钥
func GenerateRSAKeyPair(bits int) (publicKeyPEM, privateKeyPEM string, err error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", err
	}

	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return string(pubKeyPEM), string(privKeyPEM), nil
}

// RSAEncrypt 使用 RSA-OAEP 和 SHA256 加密明文
// publicKeyPEM 必须是 PEM 格式
func RSAEncrypt(plaintext []byte, publicKeyPEM string) ([]byte, error) {
	pubKey, err := parsePublicKey(publicKeyPEM)
	if err != nil {
		return nil, err
	}

	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, plaintext, nil)
}

// RSADecrypt 使用 RSA-OAEP 和 SHA256 解密密文
// privateKeyPEM 必须是 PEM 格式
func RSADecrypt(ciphertext []byte, privateKeyPEM string) ([]byte, error) {
	privKey, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}
	return plaintext, nil
}

// RSAEncryptPKCS1v15 使用 RSA PKCS1v15 加密明文
//
// 已弃用：PKCS1v15 不如 OAEP 安全。新应用请使用 RSAEncrypt
func RSAEncryptPKCS1v15(plaintext []byte, publicKeyPEM string) ([]byte, error) {
	pubKey, err := parsePublicKey(publicKeyPEM)
	if err != nil {
		return nil, err
	}

	return rsa.EncryptPKCS1v15(rand.Reader, pubKey, plaintext)
}

// RSADecryptPKCS1v15 使用 RSA PKCS1v15 解密密文
//
// 已弃用：PKCS1v15 不如 OAEP 安全。新应用请使用 RSADecrypt
func RSADecryptPKCS1v15(ciphertext []byte, privateKeyPEM string) ([]byte, error) {
	privKey, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}
	return plaintext, nil
}

func parsePublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	publicKeyPEM = cleanPEM(publicKeyPEM)

	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("%w: failed to decode public key PEM block", ErrInvalidPEM)
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%w: not an RSA public key", ErrInvalidKeyType)
	}

	return rsaPub, nil
}

func parsePrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	privateKeyPEM = cleanPEM(privateKeyPEM)

	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("%w: failed to decode private key PEM block", ErrInvalidPEM)
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("%w: not an RSA private key", ErrInvalidKeyType)
		}
		return rsaKey, nil
	}

	return privKey, nil
}

func cleanPEM(pemStr string) string {
	var b strings.Builder
	for line := range strings.SplitSeq(pemStr, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		b.WriteString(trimmed)
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}
