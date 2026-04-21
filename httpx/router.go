package httpx

// Router 定义路由注册能力。
type Router interface {
	GET(path string, h Handler, mw ...Middleware)
	POST(path string, h Handler, mw ...Middleware)
	PUT(path string, h Handler, mw ...Middleware)
	DELETE(path string, h Handler, mw ...Middleware)
	PATCH(path string, h Handler, mw ...Middleware)
	HEAD(path string, h Handler, mw ...Middleware)
	OPTIONS(path string, h Handler, mw ...Middleware)
	Any(path string, h Handler, mw ...Middleware)

	Group(prefix string, mw ...Middleware) Router
	Use(mw ...Middleware)

	Static(prefix, root string)
	StaticFile(path, file string)
}
