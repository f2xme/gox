// Package crypto provides cryptographic utilities including password hashing and digital signatures.
package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt with default cost.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a bcrypt hash.
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Argon2Params defines parameters for Argon2id hashing.
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Params returns recommended Argon2id parameters.
func DefaultArgon2Params() *Argon2Params {
	return &Argon2Params{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// HashPasswordArgon2 hashes a password using Argon2id.
func HashPasswordArgon2(password string, params *Argon2Params) (string, error) {
	if params == nil {
		params = DefaultArgon2Params()
	}

	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	// Encode as: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.Memory,
		params.Iterations,
		params.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// VerifyPasswordArgon2 verifies a password against an Argon2id hash.
func VerifyPasswordArgon2(password, encodedHash string) (bool, error) {
	var version int
	var memory, iterations uint32
	var parallelism uint8
	var salt, hash string

	_, err := fmt.Sscanf(encodedHash, "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		&version, &memory, &iterations, &parallelism, &salt, &hash)
	if err != nil {
		return false, fmt.Errorf("invalid hash format: %w", err)
	}

	saltBytes, err := base64.RawStdEncoding.DecodeString(salt)
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	hashBytes, err := base64.RawStdEncoding.DecodeString(hash)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	computedHash := argon2.IDKey([]byte(password), saltBytes, iterations, memory, parallelism, uint32(len(hashBytes)))

	if len(computedHash) != len(hashBytes) {
		return false, nil
	}

	for i := range computedHash {
		if computedHash[i] != hashBytes[i] {
			return false, nil
		}
	}

	return true, nil
}
