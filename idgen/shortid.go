package idgen

import (
	"crypto/rand"
)

const (
	// base62Alphabet 是 Base62 编码字符集（0-9, A-Z, a-z）
	base62Alphabet       = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	defaultShortIDLength = 8
)

// ShortID 生成默认长度为 8 的短小 URL 安全 ID
// 如果随机数生成失败则返回错误
func ShortID() (string, error) {
	return ShortIDWithLength(defaultShortIDLength)
}

// ShortIDWithLength 生成指定长度的短小 URL 安全 ID
// 如果随机数生成失败则返回错误
func ShortIDWithLength(length int) (string, error) {
	if length <= 0 {
		length = defaultShortIDLength
	}

	result := make([]byte, length)
	buf := make([]byte, length)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		result[i] = base62Alphabet[int(buf[i])%len(base62Alphabet)]
	}

	return string(result), nil
}
