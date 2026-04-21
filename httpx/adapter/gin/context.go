package gin

import (
	"net/http"

	"github.com/f2xme/gox/httpx"
	ginframework "github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)


type ginContext struct {
	c *ginframework.Context
}

var _ httpx.Context = (*ginContext)(nil)

func (ctx *ginContext) Request() *http.Request       { return ctx.c.Request }
func (ctx *ginContext) Param(key string) httpx.Value  { return httpx.Value(ctx.c.Param(key)) }
func (ctx *ginContext) Query(key string) httpx.Value  { return httpx.Value(ctx.c.Query(key)) }
func (ctx *ginContext) QueryAll(key string) []string  { return ctx.c.QueryArray(key) }
func (ctx *ginContext) Header(key string) httpx.Value { return httpx.Value(ctx.c.GetHeader(key)) }
func (ctx *ginContext) Cookie(name string) (*http.Cookie, error) {
	return ctx.c.Request.Cookie(name)
}
func (ctx *ginContext) ClientIP() string { return ctx.c.ClientIP() }
func (ctx *ginContext) Method() string   { return ctx.c.Request.Method }
func (ctx *ginContext) Path() string     { return ctx.c.Request.URL.Path }

func (ctx *ginContext) Bind(v any) error      { return ctx.c.ShouldBind(v) }
func (ctx *ginContext) BindJSON(v any) error   { return ctx.c.ShouldBindJSON(v) }
func (ctx *ginContext) BindQuery(v any) error  { return ctx.c.ShouldBindQuery(v) }
func (ctx *ginContext) BindForm(v any) error   { return ctx.c.ShouldBindWith(v, binding.Form) }

func (ctx *ginContext) JSON(code int, v any) error {
	ctx.c.JSON(code, v)
	return nil
}

func (ctx *ginContext) String(code int, s string) error {
	ctx.c.String(code, "%s", s)
	return nil
}

func (ctx *ginContext) HTML(code int, html string) error {
	ctx.c.Data(code, "text/html; charset=utf-8", []byte(html))
	return nil
}

func (ctx *ginContext) Blob(code int, contentType string, data []byte) error {
	ctx.c.Data(code, contentType, data)
	return nil
}

func (ctx *ginContext) NoContent(code int) error {
	ctx.c.Status(code)
	return nil
}

func (ctx *ginContext) Redirect(code int, url string) error {
	ctx.c.Redirect(code, url)
	return nil
}

func (ctx *ginContext) SetHeader(key, value string) { ctx.c.Header(key, value) }

func (ctx *ginContext) SetCookie(cookie *http.Cookie) {
	ctx.c.SetCookie(
		cookie.Name, cookie.Value, cookie.MaxAge,
		cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly,
	)
}

func (ctx *ginContext) Status(code int) { ctx.c.Status(code) }

func (ctx *ginContext) Set(key string, value any)    { ctx.c.Set(key, value) }
func (ctx *ginContext) Get(key string) (any, bool)   { return ctx.c.Get(key) }
func (ctx *ginContext) MustGet(key string) any       { return ctx.c.MustGet(key) }
func (ctx *ginContext) ResponseWriter() http.ResponseWriter { return ctx.c.Writer }
func (ctx *ginContext) Raw() any                     { return ctx.c }
