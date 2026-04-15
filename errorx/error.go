package errorx

import (
	"fmt"
)

// Error 表示带有额外上下文的增强错误
type Error struct {
	// Code 错误码（例如 "ERR001"）
	Code string
	// Message 错误消息
	Message string
	// Kind 错误类别
	Kind Kind
	// Stack 堆栈跟踪
	Stack []StackFrame
	// Cause 底层错误
	Cause error
	// Metadata 包含额外的上下文信息
	Metadata map[string]any
}

func (e *Error) Error() string {
	if e.Code != "" {
		if e.Cause != nil {
			return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
		}
		return fmt.Sprintf("[%s] %s", e.Code, e.Message)
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// WithKind 设置错误类别并返回错误以支持链式调用
func (e *Error) WithKind(kind Kind) *Error {
	e.Kind = kind
	return e
}

// WithCode 设置错误码并返回错误以支持链式调用
func (e *Error) WithCode(code string) *Error {
	e.Code = code
	return e
}

// WithMetadata 添加元数据并返回错误以支持链式调用
func (e *Error) WithMetadata(key string, value any) *Error {
	if e.Metadata == nil {
		e.Metadata = make(map[string]any)
	}
	e.Metadata[key] = value
	return e
}

// New 创建带有给定消息的新错误
func New(message string) *Error {
	return &Error{
		Message: message,
		Kind:    KindUnknown,
		Stack:   captureStack(2),
	}
}

// NewWithKind 创建带有给定类别和消息的新错误
func NewWithKind(kind Kind, message string) *Error {
	return &Error{
		Message: message,
		Kind:    kind,
		Stack:   captureStack(2),
	}
}

// NewCode 创建带有给定错误码和消息的新错误
func NewCode(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Kind:    KindUnknown,
		Stack:   captureStack(2),
	}
}

// Wrap 用额外的上下文包装现有错误
// 如果 err 为 nil，Wrap 返回 nil
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Message: message,
		Kind:    KindUnknown,
		Cause:   err,
		Stack:   captureStack(2),
	}
}
