package idverify

import (
	"errors"
	"fmt"
)

// 哨兵错误，可用 errors.Is 判断。
var (
	// ErrNotConfigured 表示密钥或必要配置缺失。
	ErrNotConfigured = errors.New("idverify: not configured")
	// ErrInvalidArgument 表示请求参数无效。
	ErrInvalidArgument = errors.New("idverify: invalid argument")
	// ErrUnavailable 表示上游服务不可用或全部节点失败。
	ErrUnavailable = errors.New("idverify: service unavailable")
)

// Error 带提供方上下文的核验错误。
type Error struct {
	// Provider 提供方编码。
	Provider string
	// Op 操作名，例如 verify、token。
	Op string
	// Err 底层错误。
	Err error
}

// Error 实现 error。
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Provider != "" && e.Op != "" {
		return fmt.Sprintf("idverify %s %s: %v", e.Provider, e.Op, e.Err)
	}
	if e.Provider != "" {
		return fmt.Sprintf("idverify %s: %v", e.Provider, e.Err)
	}
	return fmt.Sprintf("idverify: %v", e.Err)
}

// Unwrap 支持 errors.Is / errors.As。
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Wrap 包装错误并附带 provider/op。
func Wrap(provider, op string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{Provider: provider, Op: op, Err: err}
}
