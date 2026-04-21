package httpx

// Handler 是统一的请求处理函数类型。
type Handler func(ctx Context) error

// Middleware 包装 Handler，以洋葱模型添加横切关注点。
type Middleware func(next Handler) Handler

// ErrorHandler 处理 Handler 返回的错误，负责根据错误类型写入 HTTP 响应。
type ErrorHandler func(ctx Context, err error)
