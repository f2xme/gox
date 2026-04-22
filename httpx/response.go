package httpx

import "net/http"

// Response 是统一的 API 响应格式。
//
// 语义维度：
//   - Success：业务是否成功（与 HTTP 状态码解耦）
//   - Message：面向用户的提示消息
//   - Data：业务载荷，可为任意 JSON 可序列化值；为空时不输出该字段
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// NewDataResponse 创建带数据的成功响应。可选传入自定义消息，默认为 "ok"。
func NewDataResponse(data any, msg ...string) *Response {
	m := "ok"
	if len(msg) > 0 {
		m = msg[0]
	}
	return &Response{
		Success: true,
		Message: m,
		Data:    data,
	}
}

// NewDoneResponse 创建仅含提示消息的成功响应（无 data 字段）。
// 适用于删除、更新、重置等仅需告知结果的场景。
func NewDoneResponse(msg string) *Response {
	return &Response{
		Success: true,
		Message: msg,
	}
}

// NewFailResponse 创建失败响应。可选传入附加数据（如字段级校验错误详情）。
func NewFailResponse(msg string, data ...any) *Response {
	resp := &Response{
		Success: false,
		Message: msg,
	}
	if len(data) > 0 {
		resp.Data = data[0]
	}
	return resp
}

// Data 以统一成功响应格式向客户端输出 HTTP 200 响应，携带业务数据。
// 可选传入自定义消息覆盖默认的 "ok"。
//
//	httpx.Data(c, user)                 // message="ok"
//	httpx.Data(c, user, "创建成功")       // message="创建成功"
func Data(c Context, data any, msg ...string) error {
	return c.JSON(http.StatusOK, NewDataResponse(data, msg...))
}

// Done 以统一成功响应格式向客户端输出 HTTP 200 响应，仅携带提示消息。
// 响应体不包含 data 字段。
//
//	httpx.Done(c, "删除成功")
func Done(c Context, msg string) error {
	return c.JSON(http.StatusOK, NewDoneResponse(msg))
}

// Fail 以统一失败响应格式向客户端输出 HTTP 200 响应（业务失败，非 HTTP 错误）。
// 可选传入附加数据，用于承载如字段级校验错误等结构化详情。
//
//	httpx.Fail(c, "参数错误")
//	httpx.Fail(c, "表单校验失败", fieldErrors)
func Fail(c Context, msg string, data ...any) error {
	return c.JSON(http.StatusOK, NewFailResponse(msg, data...))
}
