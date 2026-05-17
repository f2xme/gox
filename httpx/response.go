package httpx

import "net/http"

// Response 是可选的统一业务 API 响应格式。
// httpx 默认错误处理器不依赖该格式；只有显式调用 Data、Done 或 Fail 时才会使用。
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

// Data 以统一成功响应格式向客户端输出 HTTP 200 响应，携带业务数据。
// 可选传入自定义消息覆盖默认的 "ok"。
//
//	httpx.Data(c, user)                 // message="ok"
//	httpx.Data(c, user, "创建成功")       // message="创建成功"
func Data(c Context, data any, msg ...string) error {
	m := "ok"
	if len(msg) > 0 {
		m = msg[0]
	}
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Message: m,
		Data:    data,
	})
}

// Done 以统一成功响应格式向客户端输出 HTTP 200 响应，仅携带提示消息。
// 响应体不包含 data 字段。
//
//	httpx.Done(c, "删除成功")
func Done(c Context, msg string) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Message: msg,
	})
}

// Fail 以可选统一失败响应格式向客户端输出 HTTP 200 响应（业务失败，非 HTTP 错误）。
// 可选传入附加数据，用于承载如字段级校验错误等结构化详情。
//
//	httpx.Fail(c, "参数错误")
//	httpx.Fail(c, "表单校验失败", fieldErrors)
func Fail(c Context, msg string, data ...any) error {
	resp := Response{
		Success: false,
		Message: msg,
	}
	if len(data) > 0 {
		resp.Data = data[0]
	}
	return c.JSON(http.StatusOK, resp)
}
