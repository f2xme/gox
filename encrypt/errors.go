package encrypt

import "errors"

// ErrInvalidKeySize 当密钥大小对算法无效时返回
var ErrInvalidKeySize = errors.New("encrypt: invalid key size")

// ErrInvalidCiphertext 当密文格式错误或太短时返回
var ErrInvalidCiphertext = errors.New("encrypt: invalid ciphertext")

// ErrDecryptionFailed 当解密因认证或填充错误失败时返回
var ErrDecryptionFailed = errors.New("encrypt: decryption failed")

// ErrInvalidPEM 当 PEM 数据无法解码时返回
var ErrInvalidPEM = errors.New("encrypt: invalid PEM format")

// ErrInvalidKeyType 当解析的密钥不是预期类型时返回
var ErrInvalidKeyType = errors.New("encrypt: invalid key type")
