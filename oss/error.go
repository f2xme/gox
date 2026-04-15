package oss

import "fmt"

// Error OSS 错误
type Error struct {
	Code    string // 错误码
	Message string // 错误消息
	Key     string // 对象键（如果适用）
	Err     error  // 原始错误
}

func (e *Error) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("oss: %s (key=%s): %s", e.Code, e.Key, e.Message)
	}
	return fmt.Sprintf("oss: %s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

const (
	ErrCodeNotFound        = "NotFound"            // 对象不存在
	ErrCodeAccessDenied    = "AccessDenied"        // 访问被拒绝
	ErrCodeInvalidArgument = "InvalidArgument"     // 无效参数
	ErrCodeBucketNotEmpty  = "BucketNotEmpty"      // 存储桶不为空
	ErrCodeBucketExists    = "BucketAlreadyExists" // 存储桶已存在
	ErrCodeInternal        = "InternalError"       // 内部错误
)

// NewError 创建一个新的 OSS 错误
func NewError(code, message string, key ...string) *Error {
	e := &Error{Code: code, Message: message}
	if len(key) > 0 {
		e.Key = key[0]
	}
	return e
}

// WrapError 包装一个错误为 OSS 错误
func WrapError(code, message string, err error, key ...string) *Error {
	e := &Error{Code: code, Message: message, Err: err}
	if len(key) > 0 {
		e.Key = key[0]
	}
	return e
}
