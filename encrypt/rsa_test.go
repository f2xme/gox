// encrypt/rsa_test.go
package encrypt

import (
	"strings"
	"testing"
)

func TestGenerateRSAKeyPair(t *testing.T) {
	tests := []struct {
		name string
		bits int
	}{
		{"2048", 2048},
		{"3072", 3072},
		{"4096", 4096},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pubKey, privKey, err := GenerateRSAKeyPair(tt.bits)
			if err != nil {
				t.Fatalf("GenerateRSAKeyPair() error = %v", err)
			}

			if !strings.Contains(pubKey, "BEGIN PUBLIC KEY") {
				t.Error("public key should be in PEM format")
			}
			if !strings.Contains(privKey, "BEGIN RSA PRIVATE KEY") {
				t.Error("private key should be in PEM format")
			}
		})
	}
}

func TestParsePublicKeyInvalidPEM(t *testing.T) {
	_, err := parsePublicKey("not a pem")
	if err == nil {
		t.Error("parsePublicKey() should fail with invalid PEM")
	}
}

func TestParsePrivateKeyInvalidPEM(t *testing.T) {
	_, err := parsePrivateKey("not a pem")
	if err == nil {
		t.Error("parsePrivateKey() should fail with invalid PEM")
	}
}

func TestCleanPEM(t *testing.T) {
	messy := "  -----BEGIN PUBLIC KEY-----  \n  abc  \n  def  \n  -----END PUBLIC KEY-----  "
	cleaned := cleanPEM(messy)
	lines := strings.Split(cleaned, "\n")
	for _, line := range lines {
		if line != strings.TrimSpace(line) {
			t.Errorf("line has leading/trailing whitespace: %q", line)
		}
	}
}

func TestRSAEncryptDecrypt(t *testing.T) {
	pubKey, privKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair() error = %v", err)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"hello", []byte("Hello, RSA encryption!")},
		{"empty", []byte("")},
		{"chinese", []byte("你好，这是一个 RSA 加密测试！")},
		{"binary", []byte{0x00, 0xff, 0x01, 0xfe}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := RSAEncrypt(tt.plaintext, pubKey)
			if err != nil {
				t.Fatalf("RSAEncrypt() error = %v", err)
			}

			decrypted, err := RSADecrypt(ciphertext, privKey)
			if err != nil {
				t.Fatalf("RSADecrypt() error = %v", err)
			}

			if string(decrypted) != string(tt.plaintext) {
				t.Errorf("decrypted = %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestRSAEncryptDecryptPKCS1v15(t *testing.T) {
	pubKey, privKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair() error = %v", err)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"hello", []byte("Hello, PKCS1v15!")},
		{"chinese", []byte("你好，PKCS1v15 测试")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := RSAEncryptPKCS1v15(tt.plaintext, pubKey)
			if err != nil {
				t.Fatalf("RSAEncryptPKCS1v15() error = %v", err)
			}

			decrypted, err := RSADecryptPKCS1v15(ciphertext, privKey)
			if err != nil {
				t.Fatalf("RSADecryptPKCS1v15() error = %v", err)
			}

			if string(decrypted) != string(tt.plaintext) {
				t.Errorf("decrypted = %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestRSAEncryptWithInvalidKey(t *testing.T) {
	_, err := RSAEncrypt([]byte("test"), "invalid key")
	if err == nil {
		t.Error("RSAEncrypt() should fail with invalid key")
	}

	_, err = RSADecrypt([]byte("test"), "invalid key")
	if err == nil {
		t.Error("RSADecrypt() should fail with invalid key")
	}

	_, err = RSAEncryptPKCS1v15([]byte("test"), "invalid key")
	if err == nil {
		t.Error("RSAEncryptPKCS1v15() should fail with invalid key")
	}

	_, err = RSADecryptPKCS1v15([]byte("test"), "invalid key")
	if err == nil {
		t.Error("RSADecryptPKCS1v15() should fail with invalid key")
	}
}

func TestRSADecryptWrongKey(t *testing.T) {
	pubKey1, _, _ := GenerateRSAKeyPair(2048)
	_, privKey2, _ := GenerateRSAKeyPair(2048)

	ciphertext, err := RSAEncrypt([]byte("secret"), pubKey1)
	if err != nil {
		t.Fatalf("RSAEncrypt() error = %v", err)
	}

	_, err = RSADecrypt(ciphertext, privKey2)
	if err == nil {
		t.Error("RSADecrypt() should fail with wrong key")
	}
}

func TestRSAEncryptWithWhitespaceInPEM(t *testing.T) {
	pubKey, privKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair() error = %v", err)
	}

	// 在 PEM 每行前后添加空白字符
	messyPub := strings.ReplaceAll(pubKey, "\n", "  \n  ")
	messyPriv := strings.ReplaceAll(privKey, "\n", "  \n  ")

	plaintext := []byte("whitespace test")
	ciphertext, err := RSAEncrypt(plaintext, messyPub)
	if err != nil {
		t.Fatalf("RSAEncrypt() with messy PEM error = %v", err)
	}

	decrypted, err := RSADecrypt(ciphertext, messyPriv)
	if err != nil {
		t.Fatalf("RSADecrypt() with messy PEM error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}
