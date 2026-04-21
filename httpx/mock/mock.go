package mock

import (
	"io"
	"net/http"

	"github.com/f2xme/gox/httpx"
)

// MockContext 提供用于测试的 httpx.Context 实现。
// 所有字段均可导出，方便在测试中直接访问和修改。
type MockContext struct {
	// 请求信息
	MethodValue   string
	PathValue     string
	HostValue     string
	ClientIPValue string
	Headers       map[string]string
	QueryParams   map[string][]string
	PathParams    map[string]string
	Cookies       map[string]*http.Cookie
	// BodyValue 是可选的请求体。若为 nil，则 Request().Body 为 nil。
	BodyValue io.ReadCloser

	// 响应信息
	RespCode    int
	RespBody    any
	RespHeaders http.Header

	// 上下文存储
	Store map[string]any

	// 内置的 ResponseWriter，基于 RespHeaders 实时反射响应头。
	writer *mockResponseWriter
}

// NewMockContext 创建一个新的 MockContext 实例。
// 默认 HostValue 为 "example.com"，便于 Request() 构造出可用的 URL。
func NewMockContext(method, path string) *MockContext {
	return &MockContext{
		MethodValue:   method,
		PathValue:     path,
		HostValue:     "example.com",
		ClientIPValue: "127.0.0.1",
		Headers:       make(map[string]string),
		QueryParams:   make(map[string][]string),
		PathParams:    make(map[string]string),
		Cookies:       make(map[string]*http.Cookie),
		RespHeaders:   make(http.Header),
		Store:         make(map[string]any),
	}
}

// Request 基于当前 Method/Path/Host/Body/Headers/QueryParams/Cookies 构造 *http.Request。
// 每次调用都会重新构造一个新的请求对象，避免跨用例状态污染。
func (m *MockContext) Request() *http.Request {
	host := m.HostValue
	if host == "" {
		host = "example.com"
	}
	url := "https://" + host + m.PathValue
	req, err := http.NewRequest(m.MethodValue, url, m.BodyValue)
	if err != nil {
		// 回退：保证返回非 nil，避免调用方空指针。
		req, _ = http.NewRequest(http.MethodGet, "https://"+host+"/", nil)
	}
	req.Host = host
	if len(m.QueryParams) > 0 {
		q := req.URL.Query()
		for k, vs := range m.QueryParams {
			for _, v := range vs {
				q.Add(k, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}
	for k, v := range m.Headers {
		req.Header.Set(k, v)
	}
	for _, c := range m.Cookies {
		req.AddCookie(c)
	}
	return req
}

func (m *MockContext) Param(key string) httpx.Value { return httpx.Value(m.PathParams[key]) }

func (m *MockContext) Query(key string) httpx.Value {
	if vs := m.QueryParams[key]; len(vs) > 0 {
		return httpx.Value(vs[0])
	}
	return ""
}

func (m *MockContext) QueryAll(key string) []string {
	return m.QueryParams[key]
}

func (m *MockContext) Header(key string) httpx.Value { return httpx.Value(m.Headers[key]) }

func (m *MockContext) Cookie(name string) (*http.Cookie, error) {
	if c, ok := m.Cookies[name]; ok {
		return c, nil
	}
	return nil, http.ErrNoCookie
}

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
	m.RespHeaders.Set("Content-Type", contentType)
	return nil
}

func (m *MockContext) NoContent(code int) error {
	m.RespCode = code
	m.RespBody = nil
	return nil
}

func (m *MockContext) Redirect(code int, url string) error {
	m.RespCode = code
	m.RespHeaders.Set("Location", url)
	return nil
}

func (m *MockContext) SetHeader(key, value string) {
	m.RespHeaders.Set(key, value)
}

func (m *MockContext) SetCookie(c *http.Cookie) {
	if m.Cookies == nil {
		m.Cookies = make(map[string]*http.Cookie)
	}
	m.Cookies[c.Name] = c
}

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

// ResponseWriter 返回一个实时映射到 RespHeaders 的 http.ResponseWriter。
// 多次调用返回同一个实例，其 Header() 始终反映最新的 RespHeaders。
func (m *MockContext) ResponseWriter() http.ResponseWriter {
	if m.writer == nil {
		m.writer = &mockResponseWriter{ctx: m}
	}
	return m.writer
}

func (m *MockContext) Raw() any {
	return nil
}

// mockResponseWriter 是一个最小的 http.ResponseWriter 实现，用于中间件测试。
// Header() 动态反映 MockContext.RespHeaders，Write/WriteHeader 会记录到 RespCode。
type mockResponseWriter struct {
	ctx *MockContext
}

func (w *mockResponseWriter) Header() http.Header        { return w.ctx.RespHeaders }
func (w *mockResponseWriter) Write(b []byte) (int, error) { return len(b), nil }
func (w *mockResponseWriter) WriteHeader(code int)        { w.ctx.RespCode = code }

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
	return m.JSON(code, httpx.NewFailResponse(httpx.FirstMsg(msg, defaultMsg)))
}
