package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 算法哈希密码，使用默认 cost。
//
// 参数：
//   - password: 待哈希的明文密码
//
// 返回值：
//   - string: bcrypt 哈希字符串
//   - error: 哈希失败时返回错误
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword 验证密码是否匹配 bcrypt 哈希。
//
// 参数：
//   - password: 待验证的明文密码
//   - hash: bcrypt 哈希字符串
//
// 返回值：
//   - bool: 密码匹配返回 true，否则返回 false
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Argon2Params 定义 Argon2id 哈希算法的参数。
type Argon2Params struct {
	// Memory 内存使用量（KB），推荐 64MB 以上
	Memory uint32
	// Iterations 迭代次数，推荐 3 次以上
	Iterations uint32
	// Parallelism 并行度，推荐 2 以上
	Parallelism uint8
	// SaltLength 盐的长度（字节），推荐 16 字节
	SaltLength uint32
	// KeyLength 生成密钥的长度（字节），推荐 32 字节
	KeyLength uint32
}

// DefaultArgon2Params 返回推荐的 Argon2id 参数配置。
//
// 默认配置：
//   - 内存：64 MB
//   - 迭代次数：3
//   - 并行度：2
//   - 盐长度：16 字节
//   - 密钥长度：32 字节
func DefaultArgon2Params() *Argon2Params {
	return &Argon2Params{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// HashPasswordArgon2 使用 Argon2id 算法哈希密码。
//
// 参数：
//   - password: 待哈希的明文密码
//   - params: Argon2id 参数，传 nil 使用默认参数
//
// 返回值：
//   - string: Argon2id 哈希字符串，格式：$argon2id$v=19$m=65536,t=3,p=2$salt$hash
//   - error: 哈希失败时返回错误
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

// VerifyPasswordArgon2 验证密码是否匹配 Argon2id 哈希。
//
// 参数：
//   - password: 待验证的明文密码
//   - encodedHash: Argon2id 哈希字符串
//
// 返回值：
//   - bool: 密码匹配返回 true，否则返回 false
//   - error: 哈希格式无效或解码失败时返回错误
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
