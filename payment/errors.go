package payment

import (
	"errors"
	"fmt"
)

var (
	// ErrNotImplemented 表示适配器能力尚未实现。
	// Deprecated: 内置支付宝和微信 adapter 已实现真实网关。
	ErrNotImplemented = errors.New("payment adapter is not implemented")
	// ErrInvalidConfig 表示支付配置无效。
	ErrInvalidConfig = errors.New("payment: invalid config")
	// ErrInvalidRequest 表示支付请求参数无效。
	ErrInvalidRequest = errors.New("payment: invalid request")
	// ErrGateway 表示支付网关调用失败。
	ErrGateway = errors.New("payment: gateway error")
	// ErrInvalidSignature 表示签名校验失败。
	ErrInvalidSignature = errors.New("payment: invalid signature")
	// ErrUnknownStatus 表示服务商返回未知状态。
	ErrUnknownStatus = errors.New("payment: unknown status")
	// ErrExpired 表示支付码已经过期。
	ErrExpired = errors.New("payment: expired")
	// ErrUnsupportedClient 表示扫码客户端不受支持。
	ErrUnsupportedClient = errors.New("payment: unsupported client")
	// ErrInvalidOAuthState 表示 OAuth state 校验失败。
	ErrInvalidOAuthState = errors.New("payment: invalid oauth state")
)

// ProviderError 表示支付服务商操作错误。
type ProviderError struct {
	// Provider 是支付服务商。
	Provider Provider
	// Operation 是失败操作。
	Operation string
	// Code 是服务商错误码。
	Code string
	// Message 是可安全展示的错误摘要。
	Message string
	// Err 是原始原因。
	Err error
}

// Error 返回不包含凭据和原始请求体的错误描述。
func (e *ProviderError) Error() string {
	if e == nil {
		return ErrGateway.Error()
	}
	base := fmt.Sprintf("%s %s", e.Provider, e.Operation)
	if e.Code != "" {
		base += " code=" + e.Code
	}
	if e.Message != "" {
		base += ": " + e.Message
	}
	return base
}

// Unwrap 返回原始原因；未设置时返回 ErrGateway。
func (e *ProviderError) Unwrap() error {
	if e != nil && e.Err != nil {
		return e.Err
	}
	return ErrGateway
}
