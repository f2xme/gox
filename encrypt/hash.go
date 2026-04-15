// encrypt/hash.go
package encrypt

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"

	"github.com/zeebo/blake3"
)

// MD5 计算数据的 MD5 哈希值并返回小写十六进制字符串
//
// 已弃用：MD5 在密码学上已被破解。新应用请使用 SHA256 或 Blake3
func MD5(data []byte) string {
	h := md5.Sum(data)
	return hex.EncodeToString(h[:])
}

// SHA1 计算数据的 SHA-1 哈希值并返回小写十六进制字符串
//
// 已弃用：SHA1 在密码学上较弱。新应用请使用 SHA256 或 Blake3
func SHA1(data []byte) string {
	h := sha1.Sum(data)
	return hex.EncodeToString(h[:])
}

// SHA256 计算数据的 SHA-256 哈希值并返回小写十六进制字符串
func SHA256(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// SHA512 计算数据的 SHA-512 哈希值并返回小写十六进制字符串
func SHA512(data []byte) string {
	h := sha512.Sum512(data)
	return hex.EncodeToString(h[:])
}

// Blake3 计算数据的 BLAKE3 哈希值并返回小写十六进制字符串
// Blake3 比 MD5/SHA1 更快更安全，推荐用于新应用
func Blake3(data []byte) string {
	h := blake3.Sum256(data)
	return hex.EncodeToString(h[:])
}
