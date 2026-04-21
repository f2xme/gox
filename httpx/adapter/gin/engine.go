package gin

import (
	"context"
	"net/http"

	"github.com/f2xme/gox/httpx"
	ginframework "github.com/gin-gonic/gin"
)

type ginEngine struct {
	engine       *ginframework.Engine
	server       *http.Server
	errorHandler httpx.ErrorHandler
	// TODO: renderer support not yet implemented
	// renderer     httpx.Renderer
	mw []httpx.Middleware
}

var _ httpx.Engine = (*ginEngine)(nil)

func (e *ginEngine) wrapHandler(h httpx.Handler, mw []httpx.Middleware) ginframework.HandlerFunc {
	handler := h
	for i := len(mw) - 1; i >= 0; i-- {
		handler = mw[i](handler)
	}
	return func(c *ginframework.Context) {
		ctx := &ginContext{c: c}
		if err := handler(ctx); err != nil {
			e.errorHandler(ctx, err)
		}
	}
}

func (e *ginEngine) combineMw(routeMw []httpx.Middleware) []httpx.Middleware {
	if len(e.mw) == 0 {
		return routeMw
	}
	combined := make([]httpx.Middleware, 0, len(e.mw)+len(routeMw))
	combined = append(combined, e.mw...)
	combined = append(combined, routeMw...)
	return combined
}

func (e *ginEngine) Use(mw ...httpx.Middleware)    { e.mw = append(e.mw, mw...) }
func (e *ginEngine) GET(path string, h httpx.Handler, mw ...httpx.Middleware) {
	e.engine.GET(path, e.wrapHandler(h, e.combineMw(mw)))
}
func (e *ginEngine) POST(path string, h httpx.Handler, mw ...httpx.Middleware) {
	e.engine.POST(path, e.wrapHandler(h, e.combineMw(mw)))
}
func (e *ginEngine) PUT(path string, h httpx.Handler, mw ...httpx.Middleware) {
	e.engine.PUT(path, e.wrapHandler(h, e.combineMw(mw)))
}
func (e *ginEngine) DELETE(path string, h httpx.Handler, mw ...httpx.Middleware) {
	e.engine.DELETE(path, e.wrapHandler(h, e.combineMw(mw)))
}
func (e *ginEngine) PATCH(path string, h httpx.Handler, mw ...httpx.Middleware) {
	e.engine.PATCH(path, e.wrapHandler(h, e.combineMw(mw)))
}
func (e *ginEngine) HEAD(path string, h httpx.Handler, mw ...httpx.Middleware) {
	e.engine.HEAD(path, e.wrapHandler(h, e.combineMw(mw)))
}
func (e *ginEngine) OPTIONS(path string, h httpx.Handler, mw ...httpx.Middleware) {
	e.engine.OPTIONS(path, e.wrapHandler(h, e.combineMw(mw)))
}
func (e *ginEngine) Any(path string, h httpx.Handler, mw ...httpx.Middleware) {
	e.engine.Any(path, e.wrapHandler(h, e.combineMw(mw)))
}

func (e *ginEngine) Group(prefix string, mw ...httpx.Middleware) httpx.Router {
	combined := make([]httpx.Middleware, len(e.mw), len(e.mw)+len(mw))
	copy(combined, e.mw)
	combined = append(combined, mw...)
	return &ginGroup{
		group:  e.engine.Group(prefix),
		engine: e,
		mw:     combined,
	}
}

func (e *ginEngine) Static(prefix, root string)              { e.engine.Static(prefix, root) }
func (e *ginEngine) StaticFile(path, file string)            { e.engine.StaticFile(path, file) }
func (e *ginEngine) SetErrorHandler(h httpx.ErrorHandler) { e.errorHandler = h }
func (e *ginEngine) SetNotFoundHandler(h httpx.Handler) {
	e.engine.NoRoute(e.wrapHandler(h, e.mw))
}
func (e *ginEngine) SetRenderer(r httpx.Renderer) { /* TODO: not yet implemented */ }
func (e *ginEngine) Raw() any                                { return e.engine }

func (e *ginEngine) Start(addr string) error {
	e.server = &http.Server{
		Addr:    addr,
		Handler: e.engine,
	}
	return e.server.ListenAndServe()
}

func (e *ginEngine) Shutdown(ctx context.Context) error {
	if e.server != nil {
		return e.server.Shutdown(ctx)
	}
	return nil
}

type ginGroup struct {
	group  *ginframework.RouterGroup
	engine *ginEngine
	mw     []httpx.Middleware
}

var _ httpx.Router = (*ginGroup)(nil)

func (g *ginGroup) combineMw(routeMw []httpx.Middleware) []httpx.Middleware {
	if len(g.mw) == 0 {
		return routeMw
	}
	combined := make([]httpx.Middleware, 0, len(g.mw)+len(routeMw))
	combined = append(combined, g.mw...)
	combined = append(combined, routeMw...)
	return combined
}

func (g *ginGroup) Use(mw ...httpx.Middleware)    { g.mw = append(g.mw, mw...) }
func (g *ginGroup) GET(path string, h httpx.Handler, mw ...httpx.Middleware) {
	g.group.GET(path, g.engine.wrapHandler(h, g.combineMw(mw)))
}
func (g *ginGroup) POST(path string, h httpx.Handler, mw ...httpx.Middleware) {
	g.group.POST(path, g.engine.wrapHandler(h, g.combineMw(mw)))
}
func (g *ginGroup) PUT(path string, h httpx.Handler, mw ...httpx.Middleware) {
	g.group.PUT(path, g.engine.wrapHandler(h, g.combineMw(mw)))
}
func (g *ginGroup) DELETE(path string, h httpx.Handler, mw ...httpx.Middleware) {
	g.group.DELETE(path, g.engine.wrapHandler(h, g.combineMw(mw)))
}
func (g *ginGroup) PATCH(path string, h httpx.Handler, mw ...httpx.Middleware) {
	g.group.PATCH(path, g.engine.wrapHandler(h, g.combineMw(mw)))
}
func (g *ginGroup) HEAD(path string, h httpx.Handler, mw ...httpx.Middleware) {
	g.group.HEAD(path, g.engine.wrapHandler(h, g.combineMw(mw)))
}
func (g *ginGroup) OPTIONS(path string, h httpx.Handler, mw ...httpx.Middleware) {
	g.group.OPTIONS(path, g.engine.wrapHandler(h, g.combineMw(mw)))
}
func (g *ginGroup) Any(path string, h httpx.Handler, mw ...httpx.Middleware) {
	g.group.Any(path, g.engine.wrapHandler(h, g.combineMw(mw)))
}

func (g *ginGroup) Group(prefix string, mw ...httpx.Middleware) httpx.Router {
	combined := make([]httpx.Middleware, len(g.mw), len(g.mw)+len(mw))
	copy(combined, g.mw)
	combined = append(combined, mw...)
	return &ginGroup{
		group:  g.group.Group(prefix),
		engine: g.engine,
		mw:     combined,
	}
}

func (g *ginGroup) Static(prefix, root string)   { g.group.Static(prefix, root) }
func (g *ginGroup) StaticFile(path, file string)  { g.group.StaticFile(path, file) }
