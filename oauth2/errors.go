package oauth2

import (
	"errors"
	"fmt"
)

// 通用错误定义。
var (
	// ErrInvalidCode 表示授权码为空或已失效。
	ErrInvalidCode = errors.New("oauth2: invalid authorization code")
	// ErrInvalidToken 表示访问令牌为空、失效或缺少必要字段。
	ErrInvalidToken = errors.New("oauth2: invalid token")
	// ErrMissingRefreshToken 表示刷新令牌为空。
	ErrMissingRefreshToken = errors.New("oauth2: missing refresh token")
	// ErrProviderResponse 表示第三方平台返回了错误响应。
	ErrProviderResponse = errors.New("oauth2: provider response error")
	// ErrInvalidEndpoint 表示 OAuth2 端点地址无效。
	ErrInvalidEndpoint = errors.New("oauth2: invalid endpoint")
)

// ProviderError 表示第三方平台返回的业务错误。
type ProviderError struct {
	// Provider 服务提供商名称。
	Provider string
	// Code 平台错误码。
	Code string
	// Message 平台错误信息。
	Message string
	// Raw 平台原始响应。
	Raw []byte
	// StatusCode HTTP 响应状态码，非 HTTP 错误时为 0。
	StatusCode int
}

// Error 返回错误描述。
func (e *ProviderError) Error() string {
	if e == nil {
		return ErrProviderResponse.Error()
	}
	if e.Code == "" {
		return fmt.Sprintf("%s: %s", ErrProviderResponse, e.Message)
	}
	if e.Message == "" {
		return fmt.Sprintf("%s: %s code %s", ErrProviderResponse, e.Provider, e.Code)
	}
	return fmt.Sprintf("%s: %s code %s: %s", ErrProviderResponse, e.Provider, e.Code, e.Message)
}

// Unwrap 返回可用于 errors.Is 判断的根错误。
func (e *ProviderError) Unwrap() error {
	return ErrProviderResponse
}
