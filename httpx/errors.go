package httpx

import (
	"errors"
	"fmt"
	"net/http"
)

// HTTPError 表示带有 HTTP 状态码的错误。
type HTTPError struct {
	Code    int
	Message string
	Err     error
}

// NewHTTPError 创建一个新的 HTTPError。
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{Code: code, Message: message}
}

// Error 实现 error 接口。
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d, message=%s, err=%v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

// Unwrap 返回被包装的原始错误。
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// WithError 包装一个底层错误。
func (e *HTTPError) WithError(err error) *HTTPError {
	e.Err = err
	return e
}

// FirstMsg 返回 msg 中第一个非空字符串，若均为空则返回 def。
func FirstMsg(msg []string, def string) string {
	if len(msg) > 0 && msg[0] != "" {
		return msg[0]
	}
	return def
}

// DefaultErrorHandler 是默认的错误处理器。
// 将 HTTPError 映射到对应状态码，其他错误回退到 500。
func DefaultErrorHandler(ctx Context, err error) {
	var he *HTTPError
	if errors.As(err, &he) {
		_ = ctx.JSON(he.Code, NewFailResponse(he.Message))
	} else {
		_ = ctx.JSON(http.StatusInternalServerError, NewFailResponse(err.Error()))
	}
}
