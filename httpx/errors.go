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

func (e *HTTPError) withErr(err error) *HTTPError {
	if err != nil {
		e.Err = err
	}
	return e
}

// resolveArgs 从 any 变参中提取消息字符串和底层错误。
// 规则：string 覆盖默认消息；error 存入 Err 字段，若消息尚未被 string 覆盖则同时以 err.Error() 作为消息；
// fmt.Stringer 等效于 string；nil 忽略；其余类型忽略。
func resolveArgs(def string, args []any) (msg string, err error) {
	msg = def
	for _, a := range args {
		switch v := a.(type) {
		case nil:
		case string:
			if v != "" {
				msg = v
			}
		case error:
			err = v
			if msg == def {
				msg = v.Error()
			}
		case fmt.Stringer:
			if s := v.String(); s != "" {
				msg = s
			}
		}
	}
	return
}

// BizError 表示业务层错误，HTTP 状态码固定为 200，通过 success:false 标识失败。
type BizError struct {
	Message string
	Err     error
}

// NewBizError 创建一个新的 BizError。args 支持 string、error 或两者混合，规则同 resolveArgs。
func NewBizError(args ...any) *BizError {
	msg, err := resolveArgs("", args)
	return &BizError{Message: msg, Err: err}
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

// ErrBadRequest 创建 400 参数错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrBadRequest(args ...any) *HTTPError {
	msg, err := resolveArgs("Bad Request", args)
	return NewHTTPError(http.StatusBadRequest, msg).withErr(err)
}

// ErrUnauthorized 创建 401 未登录错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrUnauthorized(args ...any) *HTTPError {
	msg, err := resolveArgs("Unauthorized", args)
	return NewHTTPError(http.StatusUnauthorized, msg).withErr(err)
}

// ErrForbidden 创建 403 无权限错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrForbidden(args ...any) *HTTPError {
	msg, err := resolveArgs("Forbidden", args)
	return NewHTTPError(http.StatusForbidden, msg).withErr(err)
}

// ErrNotFound 创建 404 资源不存在错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrNotFound(args ...any) *HTTPError {
	msg, err := resolveArgs("Not Found", args)
	return NewHTTPError(http.StatusNotFound, msg).withErr(err)
}

// ErrTooManyRequests 创建 429 请求过频错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrTooManyRequests(args ...any) *HTTPError {
	msg, err := resolveArgs("Too Many Requests", args)
	return NewHTTPError(http.StatusTooManyRequests, msg).withErr(err)
}

// ErrInternalError 创建 500 服务器内部错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrInternalError(args ...any) *HTTPError {
	msg, err := resolveArgs("Internal Server Error", args)
	return NewHTTPError(http.StatusInternalServerError, msg).withErr(err)
}

// ErrServiceUnavailable 创建 503 服务不可用错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrServiceUnavailable(args ...any) *HTTPError {
	msg, err := resolveArgs("Service Unavailable", args)
	return NewHTTPError(http.StatusServiceUnavailable, msg).withErr(err)
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
