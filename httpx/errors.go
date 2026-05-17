package httpx

import (
	"errors"
	"fmt"
	"net/http"
)

// StatusError 表示带有 HTTP 状态码的错误。
type StatusError struct {
	Status  int
	Message string
	Err     error
}

// NewStatusError 创建一个新的 StatusError。
func NewStatusError(status int, message string) *StatusError {
	return &StatusError{Status: status, Message: message}
}

// NewStatusErrorCause 创建一个带底层原因的 StatusError。
func NewStatusErrorCause(status int, message string, err error) *StatusError {
	return &StatusError{Status: status, Message: message, Err: err}
}

// Error 实现 error 接口。
func (e *StatusError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("status=%d, message=%s, err=%v", e.Status, e.Message, e.Err)
	}
	return fmt.Sprintf("status=%d, message=%s", e.Status, e.Message)
}

// Unwrap 返回被包装的原始错误。
func (e *StatusError) Unwrap() error {
	return e.Err
}

// WithError 包装一个底层错误。
func (e *StatusError) WithError(err error) *StatusError {
	if e == nil {
		return &StatusError{Err: err}
	}
	clone := *e
	clone.Err = err
	return &clone
}

func (e *StatusError) withErr(err error) *StatusError {
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

// ErrBadRequest 创建 400 参数错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrBadRequest(args ...any) *StatusError {
	msg, err := resolveArgs("Bad Request", args)
	return NewStatusError(http.StatusBadRequest, msg).withErr(err)
}

// ErrUnauthorized 创建 401 未登录错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrUnauthorized(args ...any) *StatusError {
	msg, err := resolveArgs("Unauthorized", args)
	return NewStatusError(http.StatusUnauthorized, msg).withErr(err)
}

// ErrForbidden 创建 403 无权限错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrForbidden(args ...any) *StatusError {
	msg, err := resolveArgs("Forbidden", args)
	return NewStatusError(http.StatusForbidden, msg).withErr(err)
}

// ErrNotFound 创建 404 资源不存在错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrNotFound(args ...any) *StatusError {
	msg, err := resolveArgs("Not Found", args)
	return NewStatusError(http.StatusNotFound, msg).withErr(err)
}

// ErrTooManyRequests 创建 429 请求过频错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrTooManyRequests(args ...any) *StatusError {
	msg, err := resolveArgs("Too Many Requests", args)
	return NewStatusError(http.StatusTooManyRequests, msg).withErr(err)
}

// ErrInternalError 创建 500 服务器内部错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrInternalError(args ...any) *StatusError {
	msg, err := resolveArgs("Internal Server Error", args)
	return NewStatusError(http.StatusInternalServerError, msg).withErr(err)
}

// ErrServiceUnavailable 创建 503 服务不可用错误。args 支持 string、error 或两者混合，规则同 resolveArgs。
func ErrServiceUnavailable(args ...any) *StatusError {
	msg, err := resolveArgs("Service Unavailable", args)
	return NewStatusError(http.StatusServiceUnavailable, msg).withErr(err)
}

// DefaultErrorHandler 是默认的错误处理器。
// StatusError 映射到对应状态码；其他错误回退到 500。
// 5xx 响应只返回标准 HTTP 状态文本，避免向客户端暴露底层错误细节。
func DefaultErrorHandler(ctx Context, err error) {
	if err == nil {
		return
	}

	var se *StatusError
	if errors.As(err, &se) {
		status := normalizeHTTPStatus(se.Status)
		writeError(ctx, status, responseMessage(status, se.Message))
	} else {
		writeError(ctx, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

type errorResponse struct {
	Message string `json:"message"`
}

func writeError(ctx Context, code int, message string) {
	_ = ctx.JSON(code, errorResponse{Message: message})
}

func normalizeHTTPStatus(code int) int {
	if code < 100 || code > 599 {
		return http.StatusInternalServerError
	}
	return code
}

func responseMessage(code int, message string) string {
	if code >= 500 {
		if text := http.StatusText(code); text != "" {
			return text
		}
		return http.StatusText(http.StatusInternalServerError)
	}
	if message != "" {
		return message
	}
	if text := http.StatusText(code); text != "" {
		return text
	}
	return "error"
}
