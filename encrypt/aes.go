// encrypt/aes.go
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

const aesGCMNonceSize = 12

// AESEncrypt 使用 AES-GCM 和给定密钥加密明文
// 密钥必须是 16、24 或 32 字节，分别对应 AES-128、AES-192 或 AES-256
// 返回的密文格式：[nonce(12) || encrypted_data || auth_tag(16)]
func AESEncrypt(plaintext, key []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("%w: must be 16, 24, or 32 bytes, got %d", ErrInvalidKeySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCMNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// AESDecrypt 使用 AES-GCM 和给定密钥解密密文
// 密钥必须是 16、24 或 32 字节，分别对应 AES-128、AES-192 或 AES-256
// 密文必须是 AESEncrypt 生成的格式
func AESDecrypt(ciphertext, key []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("%w: must be 16, 24, or 32 bytes, got %d", ErrInvalidKeySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aesGCMNonceSize+gcm.Overhead() {
		return nil, fmt.Errorf("%w: ciphertext too short", ErrInvalidCiphertext)
	}

	nonce := ciphertext[:aesGCMNonceSize]
	encrypted := ciphertext[aesGCMNonceSize:]

	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}
