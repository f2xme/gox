package httpx

import "net/http"

// Response 是统一的 API 响应格式。
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// NewSuccessResponse 创建成功响应。
func NewSuccessResponse(data any) *Response {
	return &Response{
		Success: true,
		Message: "ok",
		Data:    data,
	}
}

// NewFailResponse 创建失败响应。
func NewFailResponse(msg string) *Response {
	return &Response{
		Success: false,
		Message: msg,
	}
}

// Success 以统一成功响应格式向客户端输出 HTTP 200 响应。
func Success(c Context, data any) error {
	return c.JSON(http.StatusOK, NewSuccessResponse(data))
}

// Fail 以统一失败响应格式向客户端输出 HTTP 200 响应（业务失败，非 HTTP 错误）。
func Fail(c Context, msg string) error {
	return c.JSON(http.StatusOK, NewFailResponse(msg))
}
