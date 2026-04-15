package httpx

// Response is the unified API response format.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// NewSuccessResponse creates a success response.
func NewSuccessResponse(data any) *Response {
	return &Response{
		Success: true,
		Message: "ok",
		Data:    data,
	}
}

// NewFailResponse creates a failure response.
func NewFailResponse(msg string) *Response {
	return &Response{
		Success: false,
		Message: msg,
	}
}
