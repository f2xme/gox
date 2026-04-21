package httpx

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
