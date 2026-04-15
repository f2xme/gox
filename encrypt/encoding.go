// encrypt/encoding.go
package encrypt

import (
	"encoding/base64"
	"encoding/hex"
)

// EncodeBase64 将数据编码为标准 Base64 字符串
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 解码标准 Base64 字符串
func DecodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// EncodeHex 将数据编码为小写十六进制字符串
func EncodeHex(data []byte) string {
	return hex.EncodeToString(data)
}

// DecodeHex 解码十六进制字符串，接受大写和小写
func DecodeHex(s string) ([]byte, error) {
	return hex.DecodeString(s)
}
