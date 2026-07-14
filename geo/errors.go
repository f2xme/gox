package geo

import (
	"errors"
	"fmt"
)

// Error 表示 IP 地区查询错误。
type Error struct {
	// Code 错误码。
	Code string
	// Message 错误消息。
	Message string
	// IP 关联的 IP 地址（如果适用）。
	IP string
	// Err 原始错误。
	Err error
}

// Error 实现 error 接口。
func (e *Error) Error() string {
	if e.IP != "" {
		return fmt.Sprintf("geo: %s (ip=%s): %s", e.Code, e.IP, e.Message)
	}
	return fmt.Sprintf("geo: %s: %s", e.Code, e.Message)
}

// Unwrap 返回被包装的原始错误。
func (e *Error) Unwrap() error {
	return e.Err
}

const (
	// ErrCodeInvalidIP 表示 IP 地址格式无效。
	ErrCodeInvalidIP = "InvalidIP"
	// ErrCodeNotFound 表示未找到对应地区信息。
	ErrCodeNotFound = "NotFound"
	// ErrCodeInvalidArgument 表示参数无效。
	ErrCodeInvalidArgument = "InvalidArgument"
	// ErrCodeUnavailable 表示上游服务或数据源不可用。
	ErrCodeUnavailable = "Unavailable"
	// ErrCodeInternal 表示内部错误。
	ErrCodeInternal = "InternalError"
)

// NewError 创建一个新的 IP 地区查询错误。
func NewError(code, message string, ip ...string) *Error {
	e := &Error{Code: code, Message: message}
	if len(ip) > 0 {
		e.IP = ip[0]
	}
	return e
}

// WrapError 包装一个错误为 IP 地区查询错误。
func WrapError(code, message string, err error, ip ...string) *Error {
	e := &Error{Code: code, Message: message, Err: err}
	if len(ip) > 0 {
		e.IP = ip[0]
	}
	return e
}

// ErrorCode 返回统一错误码。
func ErrorCode(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

// IsCode 判断错误是否为指定错误码。
func IsCode(err error, code string) bool {
	return ErrorCode(err) == code
}

// IsInvalidIP 判断错误是否为 IP 格式无效。
func IsInvalidIP(err error) bool {
	return IsCode(err, ErrCodeInvalidIP)
}

// IsNotFound 判断错误是否为未找到地区信息。
func IsNotFound(err error) bool {
	return IsCode(err, ErrCodeNotFound)
}

// IsUnavailable 判断错误是否为服务不可用。
func IsUnavailable(err error) bool {
	return IsCode(err, ErrCodeUnavailable)
}
