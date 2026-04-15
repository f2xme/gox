// encrypt/aes_test.go
package encrypt

import (
	"bytes"
	"crypto/rand"
	"errors"
	"testing"
)

func TestAESEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		keySize   int
		plaintext []byte
	}{
		{"aes128", 16, []byte("hello aes-128")},
		{"aes192", 24, []byte("hello aes-192")},
		{"aes256", 32, []byte("hello aes-256")},
		{"empty_plaintext", 32, []byte("")},
		{"chinese", 32, []byte("你好世界 AES 加密测试")},
		{"large_data", 32, make([]byte, 1024*1024)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keySize)
			if _, err := rand.Read(key); err != nil {
				t.Fatalf("failed to generate key: %v", err)
			}

			ciphertext, err := AESEncrypt(tt.plaintext, key)
			if err != nil {
				t.Fatalf("AESEncrypt() error = %v", err)
			}

			// 密文应该比明文长（nonce + tag）
			if len(ciphertext) <= len(tt.plaintext) {
				t.Errorf("ciphertext should be longer than plaintext")
			}

			decrypted, err := AESDecrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("AESDecrypt() error = %v", err)
			}

			if !bytes.Equal(tt.plaintext, decrypted) {
				t.Errorf("decrypted does not match plaintext")
			}
		})
	}
}

func TestAESEncryptInvalidKeySize(t *testing.T) {
	tests := []struct {
		name    string
		keySize int
	}{
		{"too_short", 8},
		{"15_bytes", 15},
		{"17_bytes", 17},
		{"33_bytes", 33},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keySize)
			_, err := AESEncrypt([]byte("test"), key)
			if !errors.Is(err, ErrInvalidKeySize) {
				t.Errorf("AESEncrypt() error = %v, want ErrInvalidKeySize", err)
			}
		})
	}
}

func TestAESDecryptInvalidCiphertext(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	tests := []struct {
		name       string
		ciphertext []byte
	}{
		{"too_short", []byte("short")},
		{"empty", []byte{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AESDecrypt(tt.ciphertext, key)
			if !errors.Is(err, ErrInvalidCiphertext) {
				t.Errorf("AESDecrypt() error = %v, want ErrInvalidCiphertext", err)
			}
		})
	}
}

func TestAESDecryptWrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	rand.Read(key1)
	rand.Read(key2)

	ciphertext, err := AESEncrypt([]byte("secret"), key1)
	if err != nil {
		t.Fatalf("AESEncrypt() error = %v", err)
	}

	_, err = AESDecrypt(ciphertext, key2)
	if !errors.Is(err, ErrDecryptionFailed) {
		t.Errorf("AESDecrypt() error = %v, want ErrDecryptionFailed", err)
	}
}

func TestAESEncryptProducesDifferentCiphertexts(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)
	plaintext := []byte("same input")

	ct1, _ := AESEncrypt(plaintext, key)
	ct2, _ := AESEncrypt(plaintext, key)

	if bytes.Equal(ct1, ct2) {
		t.Error("encrypting the same plaintext twice should produce different ciphertexts")
	}
}
