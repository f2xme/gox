package oss

import (
	"errors"
	"fmt"
)

// Error OSS 错误
type Error struct {
	// Code 错误码
	Code string
	// Message 错误消息
	Message string
	// Key 对象键（如果适用）
	Key string
	// Err 原始错误
	Err error
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
	// ErrCodeNotFound 表示对象不存在
	ErrCodeNotFound = "NotFound"
	// ErrCodeAccessDenied 表示访问被拒绝
	ErrCodeAccessDenied = "AccessDenied"
	// ErrCodeInvalidArgument 表示参数无效
	ErrCodeInvalidArgument = "InvalidArgument"
	// ErrCodeBucketNotEmpty 表示存储桶不为空
	ErrCodeBucketNotEmpty = "BucketNotEmpty"
	// ErrCodeBucketExists 表示存储桶已存在
	ErrCodeBucketExists = "BucketAlreadyExists"
	// ErrCodeInternal 表示内部错误
	ErrCodeInternal = "InternalError"
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

// ErrorCode 返回统一错误码
func ErrorCode(err error) string {
	var ossErr *Error
	if errors.As(err, &ossErr) {
		return ossErr.Code
	}
	return ""
}

// IsCode 判断错误是否为指定错误码
func IsCode(err error, code string) bool {
	return ErrorCode(err) == code
}

// IsNotFound 判断错误是否为对象不存在
func IsNotFound(err error) bool {
	return IsCode(err, ErrCodeNotFound)
}

// IsAccessDenied 判断错误是否为访问被拒绝
func IsAccessDenied(err error) bool {
	return IsCode(err, ErrCodeAccessDenied)
}

// IsPermissionDenied 判断错误是否为访问被拒绝
//
// Deprecated: 使用 IsAccessDenied。
func IsPermissionDenied(err error) bool {
	return IsAccessDenied(err)
}
