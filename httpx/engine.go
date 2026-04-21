package httpx

import "context"

// Engine is the top-level HTTP server entry point.
type Engine interface {
	Router

	Start(addr string) error
	Shutdown(ctx context.Context) error

	SetErrorHandler(h ErrorHandler)
	SetNotFoundHandler(h Handler)
	SetRenderer(r Renderer)

	Raw() any
}
