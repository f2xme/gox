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

// BizError 表示业务层错误，HTTP 状态码固定为 200，通过 success:false 标识失败。
type BizError struct {
	Message string
	Err     error
}

// NewBizError 创建一个新的 BizError。
func NewBizError(message string) *BizError {
	return &BizError{Message: message}
}

// Error 实现 error 接口。
func (e *BizError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("biz_error: message=%s, err=%v", e.Message, e.Err)
	}
	return fmt.Sprintf("biz_error: message=%s", e.Message)
}

// Unwrap 返回被包装的原始错误。
func (e *BizError) Unwrap() error {
	return e.Err
}

// WithError 包装一个底层错误。
func (e *BizError) WithError(err error) *BizError {
	e.Err = err
	return e
}

// ErrBadRequest 创建 400 参数错误。
func ErrBadRequest(msg ...string) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, FirstMsg(msg, "Bad Request"))
}

// ErrUnauthorized 创建 401 未登录错误。
func ErrUnauthorized(msg ...string) *HTTPError {
	return NewHTTPError(http.StatusUnauthorized, FirstMsg(msg, "Unauthorized"))
}

// ErrForbidden 创建 403 无权限错误。
func ErrForbidden(msg ...string) *HTTPError {
	return NewHTTPError(http.StatusForbidden, FirstMsg(msg, "Forbidden"))
}

// ErrNotFound 创建 404 资源不存在错误。
func ErrNotFound(msg ...string) *HTTPError {
	return NewHTTPError(http.StatusNotFound, FirstMsg(msg, "Not Found"))
}

// ErrTooManyRequests 创建 429 请求过频错误。
func ErrTooManyRequests(msg ...string) *HTTPError {
	return NewHTTPError(http.StatusTooManyRequests, FirstMsg(msg, "Too Many Requests"))
}

// ErrInternalError 创建 500 服务器内部错误。
func ErrInternalError(msg ...string) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, FirstMsg(msg, "Internal Server Error"))
}

// ErrServiceUnavailable 创建 503 服务不可用错误。
func ErrServiceUnavailable(msg ...string) *HTTPError {
	return NewHTTPError(http.StatusServiceUnavailable, FirstMsg(msg, "Service Unavailable"))
}

// DefaultErrorHandler 是默认的错误处理器。
// BizError 返回 HTTP 200 并携带业务错误码；HTTPError 映射到对应状态码；其他错误回退到 500。
func DefaultErrorHandler(ctx Context, err error) {
	var be *BizError
	if errors.As(err, &be) {
		_ = ctx.JSON(http.StatusOK, NewFailResponse(be.Message))
		return
	}
	var he *HTTPError
	if errors.As(err, &he) {
		_ = ctx.JSON(he.Code, NewFailResponse(he.Message))
	} else {
		_ = ctx.JSON(http.StatusInternalServerError, NewFailResponse(err.Error()))
	}
}
