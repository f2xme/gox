package httpx

import "io"

// Renderer defines template rendering.
type Renderer interface {
	Render(w io.Writer, name string, data any) error
}
