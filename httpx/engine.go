package httpx

import "context"

// Engine 是 HTTP 服务器的顶层入口。
type Engine interface {
	Router

	Start(addr string) error
	Shutdown(ctx context.Context) error

	SetErrorHandler(h ErrorHandler)
	SetNotFoundHandler(h Handler)
	SetRenderer(r Renderer)

	Raw() any
}
