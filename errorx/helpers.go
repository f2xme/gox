package errorx

import (
	"errors"
	"fmt"
	"strings"
)

// IsKind 检查错误是否为指定类别
func IsKind(err error, kind Kind) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind == kind
	}
	return false
}

// IsRetryable 检查错误是否可重试
func IsRetryable(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind.IsRetryable()
	}
	return false
}

// IsTimeout 检查错误是否为超时错误
func IsTimeout(err error) bool {
	return IsKind(err, KindTimeout)
}

// GetCode 如果错误是 Error 类型则返回错误码
func GetCode(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

// GetStack 如果错误是 Error 类型则返回堆栈跟踪
func GetStack(err error) []StackFrame {
	var e *Error
	if errors.As(err, &e) {
		return e.Stack
	}
	return nil
}

// Format 格式化错误及其详细信息
func Format(err error) string {
	var e *Error
	if !errors.As(err, &e) {
		return err.Error()
	}

	var sb strings.Builder
	sb.WriteString(e.Error())

	if e.Kind != KindUnknown {
		sb.WriteString(fmt.Sprintf(" [Kind: %s]", e.Kind))
	}

	if len(e.Metadata) > 0 {
		sb.WriteString(" [Metadata:")
		for k, v := range e.Metadata {
			sb.WriteString(fmt.Sprintf(" %s=%v", k, v))
		}
		sb.WriteString("]")
	}

	return sb.String()
}

// FormatWithStack 格式化错误及其堆栈跟踪
func FormatWithStack(err error) string {
	var e *Error
	if !errors.As(err, &e) {
		return err.Error()
	}

	var sb strings.Builder
	sb.WriteString(Format(err))

	if len(e.Stack) > 0 {
		sb.WriteString("\nStack trace:\n")
		for i, frame := range e.Stack {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, frame.String()))
		}
	}

	return sb.String()
}
