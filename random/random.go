// Package random 提供安全的随机字符生成功能
//
// 支持多种字符集：
//   - 数字 (0-9)
//   - 小写字母 (a-z)
//   - 大写字母 (A-Z)
//   - 字母数字混合
//   - 自定义字符集
//
// 所有函数使用 crypto/rand 提供密码学安全的随机性
package random

import (
	"crypto/rand"
	"math/big"
)

const (
	// Digits 数字字符集
	Digits = "0123456789"
	// LowerLetters 小写字母字符集
	LowerLetters = "abcdefghijklmnopqrstuvwxyz"
	// UpperLetters 大写字母字符集
	UpperLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Letters 字母字符集（大小写）
	Letters = LowerLetters + UpperLetters
	// Alphanumeric 字母数字字符集
	Alphanumeric = Letters + Digits
)

// String 生成指定长度的随机字符串，使用给定的字符集
// 如果 length <= 0 或 charset 为空，返回空字符串
func String(length int, charset string) (string, error) {
	if length <= 0 || charset == "" {
		return "", nil
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range length {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

// Numeric 生成指定长度的随机数字字符串
func Numeric(length int) (string, error) {
	return String(length, Digits)
}

// Alpha 生成指定长度的随机字母字符串（大小写混合）
func Alpha(length int) (string, error) {
	return String(length, Letters)
}

// AlphaLower 生成指定长度的随机小写字母字符串
func AlphaLower(length int) (string, error) {
	return String(length, LowerLetters)
}

// AlphaUpper 生成指定长度的随机大写字母字符串
func AlphaUpper(length int) (string, error) {
	return String(length, UpperLetters)
}

// AlphaNumeric 生成指定长度的随机字母数字字符串
func AlphaNumeric(length int) (string, error) {
	return String(length, Alphanumeric)
}
