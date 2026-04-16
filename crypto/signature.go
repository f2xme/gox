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

// GenerateRSAKeyPair generates an RSA key pair with the specified bit size.
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

// SignRSA signs data using RSA-PSS with SHA-256.
func SignRSA(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	signature, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, hash[:], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	return signature, nil
}

// VerifyRSA verifies an RSA-PSS signature with SHA-256.
func VerifyRSA(publicKey *rsa.PublicKey, data, signature []byte) error {
	hash := sha256.Sum256(data)
	err := rsa.VerifyPSS(publicKey, crypto.SHA256, hash[:], signature, nil)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	return nil
}

// GenerateECDSAKeyPair generates an ECDSA key pair using P-256 curve.
func GenerateECDSAKeyPair() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
	}
	return privateKey, nil
}

// SignECDSA signs data using ECDSA with SHA-256.
func SignECDSA(privateKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	return signature, nil
}

// VerifyECDSA verifies an ECDSA signature with SHA-256.
func VerifyECDSA(publicKey *ecdsa.PublicKey, data, signature []byte) bool {
	hash := sha256.Sum256(data)
	return ecdsa.VerifyASN1(publicKey, hash[:], signature)
}

// GenerateEd25519KeyPair generates an Ed25519 key pair.
func GenerateEd25519KeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Ed25519 key: %w", err)
	}
	return publicKey, privateKey, nil
}

// SignEd25519 signs data using Ed25519.
func SignEd25519(privateKey ed25519.PrivateKey, data []byte) []byte {
	return ed25519.Sign(privateKey, data)
}

// VerifyEd25519 verifies an Ed25519 signature.
func VerifyEd25519(publicKey ed25519.PublicKey, data, signature []byte) bool {
	return ed25519.Verify(publicKey, data, signature)
}
