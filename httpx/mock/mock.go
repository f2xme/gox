package mock

import (
	"net/http"

	"github.com/f2xme/gox/httpx"
)

// MockContext 提供用于测试的 httpx.Context 实现。
// 所有字段均可导出，方便在测试中直接访问和修改。
type MockContext struct {
	// 请求信息
	MethodValue   string
	PathValue     string
	ClientIPValue string
	Headers       map[string]string
	QueryParams   map[string]string
	PathParams    map[string]string

	// 响应信息
	RespCode    int
	RespBody    any
	RespHeaders map[string]string

	// 上下文存储
	Store map[string]any
}

// NewMockContext 创建一个新的 MockContext 实例。
func NewMockContext(method, path string) *MockContext {
	return &MockContext{
		MethodValue:   method,
		PathValue:     path,
		ClientIPValue: "127.0.0.1",
		Headers:       make(map[string]string),
		QueryParams:   make(map[string]string),
		PathParams:    make(map[string]string),
		RespHeaders:   make(map[string]string),
		Store:         make(map[string]any),
	}
}

func (m *MockContext) Request() *http.Request { return nil }

func (m *MockContext) Param(key string) string { return m.PathParams[key] }

func (m *MockContext) Query(key string) string { return m.QueryParams[key] }

func (m *MockContext) QueryDefault(key, def string) string {
	if v, ok := m.QueryParams[key]; ok {
		return v
	}
	return def
}

func (m *MockContext) Header(key string) string { return m.Headers[key] }

func (m *MockContext) Cookie(string) (*http.Cookie, error) { return nil, http.ErrNoCookie }

func (m *MockContext) ClientIP() string { return m.ClientIPValue }

func (m *MockContext) Method() string { return m.MethodValue }

func (m *MockContext) Path() string { return m.PathValue }

func (m *MockContext) Bind(any) error { return nil }

func (m *MockContext) BindJSON(any) error { return nil }

func (m *MockContext) BindQuery(any) error { return nil }

func (m *MockContext) BindForm(any) error { return nil }

func (m *MockContext) JSON(code int, v any) error {
	m.RespCode = code
	m.RespBody = v
	return nil
}

func (m *MockContext) String(code int, s string) error {
	m.RespCode = code
	m.RespBody = s
	return nil
}

func (m *MockContext) HTML(code int, html string) error {
	m.RespCode = code
	m.RespBody = html
	return nil
}

func (m *MockContext) Blob(code int, contentType string, data []byte) error {
	m.RespCode = code
	m.RespBody = data
	m.RespHeaders["Content-Type"] = contentType
	return nil
}

func (m *MockContext) NoContent(code int) error {
	m.RespCode = code
	m.RespBody = nil
	return nil
}

func (m *MockContext) Redirect(code int, url string) error {
	m.RespCode = code
	m.RespHeaders["Location"] = url
	return nil
}

func (m *MockContext) SetHeader(key, value string) {
	m.RespHeaders[key] = value
}

func (m *MockContext) SetCookie(*http.Cookie) {}

func (m *MockContext) Status(code int) {
	m.RespCode = code
}

func (m *MockContext) Success(data any) error {
	return m.JSON(http.StatusOK, httpx.NewSuccessResponse(data))
}

func (m *MockContext) Fail(msg string) error {
	return m.JSON(http.StatusOK, httpx.NewFailResponse(msg))
}

func (m *MockContext) BadRequest(msg ...string) error {
	return m.error(http.StatusBadRequest, msgBadRequest, msg)
}

func (m *MockContext) Unauthorized(msg ...string) error {
	return m.error(http.StatusUnauthorized, msgUnauthorized, msg)
}

func (m *MockContext) Forbidden(msg ...string) error {
	return m.error(http.StatusForbidden, msgForbidden, msg)
}

func (m *MockContext) NotFound(msg ...string) error {
	return m.error(http.StatusNotFound, msgNotFound, msg)
}

func (m *MockContext) TooManyRequests(msg ...string) error {
	return m.error(http.StatusTooManyRequests, msgTooManyRequests, msg)
}

func (m *MockContext) InternalError(msg ...string) error {
	return m.error(http.StatusInternalServerError, msgInternalError, msg)
}

func (m *MockContext) ServiceUnavailable(msg ...string) error {
	return m.error(http.StatusServiceUnavailable, msgServiceUnavailable, msg)
}

func (m *MockContext) Set(key string, value any) {
	m.Store[key] = value
}

func (m *MockContext) Get(key string) (any, bool) {
	v, ok := m.Store[key]
	return v, ok
}

func (m *MockContext) MustGet(key string) any {
	if v, ok := m.Store[key]; ok {
		return v
	}
	panic("key not found: " + key)
}

func (m *MockContext) ResponseWriter() http.ResponseWriter {
	return nil
}

func (m *MockContext) Raw() any {
	return nil
}

const (
	msgBadRequest         = "Bad Request"
	msgUnauthorized       = "Unauthorized"
	msgForbidden          = "Forbidden"
	msgNotFound           = "Not Found"
	msgTooManyRequests    = "Too Many Requests"
	msgInternalError      = "Internal Server Error"
	msgServiceUnavailable = "Service Unavailable"
)

func (m *MockContext) error(code int, defaultMsg string, msg []string) error {
	return m.JSON(code, httpx.NewFailResponse(getMsg(msg, defaultMsg)))
}

func getMsg(msg []string, def string) string {
	if len(msg) > 0 && msg[0] != "" {
		return msg[0]
	}
	return def
}
