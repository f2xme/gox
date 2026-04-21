package httpx

import "net/http"

// Context 定义统一的 HTTP 请求/响应上下文。
//
// 设计原则：
//   - 取参接口（Param / Query / Header）统一返回 Value，便于链式类型转换与默认值回退
//   - Value 的底层类型为 string，可直接与字符串字面量比较
//   - 多值 Query 参数通过 QueryAll 获取
type Context interface {
	// Request 返回底层 *http.Request。
	Request() *http.Request

	// Param 返回 URI 路径参数（如 "/users/:id" 中的 id）。
	//
	// 示例：
	//
	//	id, err := c.Param("id").Int64()
	//	name := c.Param("name").Or("guest")
	Param(key string) Value

	// Query 返回 URL Query 参数（取第一个值）。
	//
	// 示例：
	//
	//	page := c.Query("page").IntOr(1)
	Query(key string) Value

	// QueryAll 返回同名 Query 参数的所有值（如 ?tag=a&tag=b）。
	QueryAll(key string) []string

	// Header 返回请求头单值。
	Header(key string) Value

	// Cookie 返回指定名称的 Cookie。
	Cookie(name string) (*http.Cookie, error)

	// ClientIP 返回客户端 IP。
	ClientIP() string

	// Method 返回 HTTP 方法。
	Method() string

	// Path 返回请求路径。
	Path() string

	// Bind 根据 Content-Type 自动绑定并校验请求体。
	Bind(v any) error
	// BindJSON 绑定 JSON 请求体。
	BindJSON(v any) error
	// BindQuery 绑定 Query 参数到结构体。
	BindQuery(v any) error
	// BindForm 绑定表单数据。
	BindForm(v any) error

	// JSON 输出 JSON 响应。
	JSON(code int, v any) error
	// String 输出纯文本响应。
	String(code int, s string) error
	// HTML 输出 HTML 响应。
	HTML(code int, html string) error
	// Blob 输出指定 Content-Type 的二进制响应。
	Blob(code int, contentType string, data []byte) error
	// NoContent 输出仅有状态码的空响应。
	NoContent(code int) error
	// Redirect 输出重定向响应。
	Redirect(code int, url string) error
	// SetHeader 设置响应头。
	SetHeader(key, value string)
	// SetCookie 设置响应 Cookie。
	SetCookie(cookie *http.Cookie)
	// Status 设置响应状态码（不写入响应体）。
	Status(code int)

	// Set 在上下文中保存键值。
	Set(key string, value any)
	// Get 从上下文中读取键值，不存在时 ok 为 false。
	Get(key string) (any, bool)
	// MustGet 从上下文中读取键值，不存在时 panic。
	MustGet(key string) any

	// ResponseWriter 返回底层 http.ResponseWriter。
	ResponseWriter() http.ResponseWriter
	// Raw 返回底层框架原生 Context 对象（如 *gin.Context）。
	Raw() any
}
