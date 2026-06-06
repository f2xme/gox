package validator

import (
	"errors"
	"strings"
)

// ErrValidation 表示结构体验证失败。
//
// 可配合 errors.Is 判断验证错误：
//
//	if errors.Is(err, validator.ErrValidation) {
//	    // 处理参数验证失败
//	}
var ErrValidation = errors.New("validator: validation failed")

// FieldError 表示单个字段的验证错误。
type FieldError struct {
	// Namespace 是字段的完整命名空间。
	Namespace string
	// Field 是字段名，受 WithFieldNameTag 配置影响。
	Field string
	// StructField 是结构体原始字段名。
	StructField string
	// Tag 是触发失败的验证标签。
	Tag string
	// Param 是验证标签参数。
	Param string
	// Message 是已翻译的用户友好错误消息。
	Message string
}

// ValidationError 表示一次结构体验证中的一个或多个字段错误。
type ValidationError struct {
	fields []FieldError
}

// Error 实现 error 接口。
func (e *ValidationError) Error() string {
	if e == nil {
		return ""
	}

	messages := make([]string, 0, len(e.fields))
	for _, field := range e.fields {
		messages = append(messages, field.Message)
	}
	return strings.Join(messages, "; ")
}

// Unwrap 返回验证错误哨兵，便于 errors.Is 判断。
func (e *ValidationError) Unwrap() error {
	return ErrValidation
}

// Fields 返回字段级验证错误。
func (e *ValidationError) Fields() []FieldError {
	if e == nil {
		return nil
	}

	fields := make([]FieldError, len(e.fields))
	copy(fields, e.fields)
	return fields
}

// Messages 返回所有字段错误消息。
func (e *ValidationError) Messages() []string {
	if e == nil {
		return nil
	}

	messages := make([]string, 0, len(e.fields))
	for _, field := range e.fields {
		messages = append(messages, field.Message)
	}
	return messages
}

// IsValidationError 判断 err 是否为验证错误。
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidation)
}

// AsValidationError 提取 ValidationError。
func AsValidationError(err error) (*ValidationError, bool) {
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return validationErr, true
	}
	return nil, false
}
