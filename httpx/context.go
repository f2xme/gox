package httpx

import "net/http"

// Context defines the unified HTTP request/response context.
type Context interface {
	// Request information
	Request() *http.Request
	Param(key string) string
	Query(key string) string
	QueryDefault(key, def string) string
	Header(key string) string
	Cookie(name string) (*http.Cookie, error)
	ClientIP() string
	Method() string
	Path() string

	// Request body binding (bind + validate)
	Bind(v any) error
	BindJSON(v any) error
	BindQuery(v any) error
	BindForm(v any) error

	// Response output
	JSON(code int, v any) error
	String(code int, s string) error
	HTML(code int, html string) error
	Blob(code int, contentType string, data []byte) error
	NoContent(code int) error
	Redirect(code int, url string) error
	SetHeader(key, value string)
	SetCookie(cookie *http.Cookie)
	Status(code int)

	// Unified response format
	Success(data any) error
	Fail(msg string) error

	// Context value store
	Set(key string, value any)
	Get(key string) (any, bool)
	MustGet(key string) any

	// Underlying access
	ResponseWriter() http.ResponseWriter
	Raw() any
}
