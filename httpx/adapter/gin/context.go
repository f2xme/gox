package gin

import (
	"net/http"

	"github.com/f2xme/gox/httpx"
	ginframework "github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func getMsg(msg []string, def string) string {
	if len(msg) > 0 && msg[0] != "" {
		return msg[0]
	}
	return def
}

type ginContext struct {
	c *ginframework.Context
}

var _ httpx.Context = (*ginContext)(nil)

func (ctx *ginContext) Request() *http.Request               { return ctx.c.Request }
func (ctx *ginContext) Param(key string) string              { return ctx.c.Param(key) }
func (ctx *ginContext) Query(key string) string              { return ctx.c.Query(key) }
func (ctx *ginContext) QueryDefault(key, def string) string  { return ctx.c.DefaultQuery(key, def) }
func (ctx *ginContext) Header(key string) string             { return ctx.c.GetHeader(key) }
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

func (ctx *ginContext) Success(data any) error {
	return ctx.JSON(http.StatusOK, httpx.NewSuccessResponse(data))
}

func (ctx *ginContext) Fail(msg string) error {
	return ctx.JSON(http.StatusOK, httpx.NewFailResponse(msg))
}

func (ctx *ginContext) BadRequest(msg ...string) error {
	return ctx.JSON(http.StatusBadRequest, httpx.NewFailResponse(getMsg(msg, "Bad Request")))
}

func (ctx *ginContext) Unauthorized(msg ...string) error {
	return ctx.JSON(http.StatusUnauthorized, httpx.NewFailResponse(getMsg(msg, "Unauthorized")))
}

func (ctx *ginContext) Forbidden(msg ...string) error {
	return ctx.JSON(http.StatusForbidden, httpx.NewFailResponse(getMsg(msg, "Forbidden")))
}

func (ctx *ginContext) NotFound(msg ...string) error {
	return ctx.JSON(http.StatusNotFound, httpx.NewFailResponse(getMsg(msg, "Not Found")))
}

func (ctx *ginContext) TooManyRequests(msg ...string) error {
	return ctx.JSON(http.StatusTooManyRequests, httpx.NewFailResponse(getMsg(msg, "Too Many Requests")))
}

func (ctx *ginContext) InternalError(msg ...string) error {
	return ctx.JSON(http.StatusInternalServerError, httpx.NewFailResponse(getMsg(msg, "Internal Server Error")))
}

func (ctx *ginContext) ServiceUnavailable(msg ...string) error {
	return ctx.JSON(http.StatusServiceUnavailable, httpx.NewFailResponse(getMsg(msg, "Service Unavailable")))
}

func (ctx *ginContext) Set(key string, value any)    { ctx.c.Set(key, value) }
func (ctx *ginContext) Get(key string) (any, bool)   { return ctx.c.Get(key) }
func (ctx *ginContext) MustGet(key string) any       { return ctx.c.MustGet(key) }
func (ctx *ginContext) ResponseWriter() http.ResponseWriter { return ctx.c.Writer }
func (ctx *ginContext) Raw() any                     { return ctx.c }
