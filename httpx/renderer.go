package httpx

import "io"

// Renderer 定义模板渲染接口。
type Renderer interface {
	Render(w io.Writer, name string, data any) error
}
