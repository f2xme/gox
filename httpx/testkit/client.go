package testkit

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/f2xme/gox/httpx"
)

// Client 是面向 httpx.Engine 的集成测试客户端。
type Client struct {
	t              testing.TB
	server         *httptest.Server
	httpClient     *http.Client
	baseURL        string
	defaultHeaders http.Header
}

// New 使用 httpx.Engine 创建测试客户端。
//
// engine.Raw() 必须实现 http.Handler。Gin adapter 的 Raw() 返回 *gin.Engine，
// 因此可直接用于 testkit。
func New(t testing.TB, engine httpx.Engine, opts ...Option) *Client {
	t.Helper()
	if engine == nil {
		t.Fatal("testkit: engine is nil")
	}

	handler, ok := engine.Raw().(http.Handler)
	if !ok {
		t.Fatalf("testkit: engine raw type %T does not implement http.Handler", engine.Raw())
	}

	return NewWithHandler(t, handler, opts...)
}

// NewWithHandler 使用标准 http.Handler 创建测试客户端。
func NewWithHandler(t testing.TB, handler http.Handler, opts ...Option) *Client {
	t.Helper()
	if handler == nil {
		t.Fatal("testkit: handler is nil")
	}

	server := httptest.NewServer(handler)
	c := &Client{
		t:              t,
		server:         server,
		httpClient:     server.Client(),
		baseURL:        server.URL,
		defaultHeaders: make(http.Header),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Close 关闭测试服务。
func (c *Client) Close() {
	if c != nil && c.server != nil {
		c.server.Close()
	}
}

// URL 返回测试服务的完整 URL。
func (c *Client) URL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.baseURL + path
}

// GET 发送 GET 请求。
func (c *Client) GET(path string, opts ...RequestOption) *Response {
	return c.Do(http.MethodGet, path, nil, opts...)
}

// HEAD 发送 HEAD 请求。
func (c *Client) HEAD(path string, opts ...RequestOption) *Response {
	return c.Do(http.MethodHead, path, nil, opts...)
}

// OPTIONS 发送 OPTIONS 请求。
func (c *Client) OPTIONS(path string, opts ...RequestOption) *Response {
	return c.Do(http.MethodOptions, path, nil, opts...)
}

// DELETE 发送 DELETE 请求。
func (c *Client) DELETE(path string, opts ...RequestOption) *Response {
	return c.Do(http.MethodDelete, path, nil, opts...)
}

// POST 发送 POST 请求。
func (c *Client) POST(path string, body io.Reader, opts ...RequestOption) *Response {
	return c.Do(http.MethodPost, path, body, opts...)
}

// PUT 发送 PUT 请求。
func (c *Client) PUT(path string, body io.Reader, opts ...RequestOption) *Response {
	return c.Do(http.MethodPut, path, body, opts...)
}

// PATCH 发送 PATCH 请求。
func (c *Client) PATCH(path string, body io.Reader, opts ...RequestOption) *Response {
	return c.Do(http.MethodPatch, path, body, opts...)
}

// POSTJSON 发送 JSON POST 请求。
func (c *Client) POSTJSON(path string, body any, opts ...RequestOption) *Response {
	return c.JSON(http.MethodPost, path, body, opts...)
}

// PUTJSON 发送 JSON PUT 请求。
func (c *Client) PUTJSON(path string, body any, opts ...RequestOption) *Response {
	return c.JSON(http.MethodPut, path, body, opts...)
}

// PATCHJSON 发送 JSON PATCH 请求。
func (c *Client) PATCHJSON(path string, body any, opts ...RequestOption) *Response {
	return c.JSON(http.MethodPatch, path, body, opts...)
}

// JSON 发送 JSON 请求。
func (c *Client) JSON(method, path string, body any, opts ...RequestOption) *Response {
	c.t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			c.t.Fatalf("testkit: encode json body: %v", err)
		}
	}
	opts = append([]RequestOption{WithContentType("application/json")}, opts...)
	return c.Do(method, path, &buf, opts...)
}

// POSTForm 发送 application/x-www-form-urlencoded POST 请求。
func (c *Client) POSTForm(path string, values url.Values, opts ...RequestOption) *Response {
	opts = append([]RequestOption{WithContentType("application/x-www-form-urlencoded")}, opts...)
	return c.Do(http.MethodPost, path, strings.NewReader(values.Encode()), opts...)
}

// File 表示 multipart/form-data 中的上传文件。
type File struct {
	// FieldName 是表单字段名。
	FieldName string
	// FileName 是上传文件名。
	FileName string
	// ContentType 是可选的文件 Content-Type。
	ContentType string
	// Reader 是文件内容。
	Reader io.Reader
}

// POSTMultipart 发送 multipart/form-data POST 请求。
func (c *Client) POSTMultipart(path string, fields map[string]string, files []File, opts ...RequestOption) *Response {
	c.t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			c.t.Fatalf("testkit: write multipart field %q: %v", key, err)
		}
	}
	for _, file := range files {
		if file.FieldName == "" {
			c.t.Fatal("testkit: multipart file field name is empty")
		}
		if file.FileName == "" {
			c.t.Fatal("testkit: multipart file name is empty")
		}
		if file.Reader == nil {
			c.t.Fatalf("testkit: multipart file %q reader is nil", file.FileName)
		}
		part, err := createMultipartFile(writer, file)
		if err != nil {
			c.t.Fatalf("testkit: create multipart file %q: %v", file.FileName, err)
		}
		if _, err := io.Copy(part, file.Reader); err != nil {
			c.t.Fatalf("testkit: copy multipart file %q: %v", file.FileName, err)
		}
	}
	if err := writer.Close(); err != nil {
		c.t.Fatalf("testkit: close multipart writer: %v", err)
	}

	opts = append([]RequestOption{WithContentType(writer.FormDataContentType())}, opts...)
	return c.Do(http.MethodPost, path, &buf, opts...)
}

// Do 发送 HTTP 请求并读取完整响应体。
func (c *Client) Do(method, path string, body io.Reader, opts ...RequestOption) *Response {
	c.t.Helper()

	cfg := requestConfig{
		headers: make(http.Header),
		query:   make(url.Values),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, c.URL(path), body)
	if err != nil {
		c.t.Fatalf("testkit: create request: %v", err)
	}

	for key, values := range c.defaultHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	for key, values := range cfg.headers {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	if cfg.contentType != "" {
		req.Header.Set("Content-Type", cfg.contentType)
	}
	for _, cookie := range cfg.cookies {
		req.AddCookie(cookie)
	}
	if len(cfg.query) > 0 {
		q := req.URL.Query()
		for key, values := range cfg.query {
			for _, value := range values {
				q.Add(key, value)
			}
		}
		req.URL.RawQuery = q.Encode()
	}
	if cfg.basicAuth != nil {
		req.SetBasicAuth(cfg.basicAuth.username, cfg.basicAuth.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.t.Fatalf("testkit: send request %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatalf("testkit: read response body: %v", err)
	}

	return &Response{
		t:    c.t,
		Raw:  resp,
		Body: data,
	}
}

func createMultipartFile(writer *multipart.Writer, file File) (io.Writer, error) {
	if file.ContentType == "" {
		return writer.CreateFormFile(file.FieldName, file.FileName)
	}
	header := make(textprotoMIMEHeader)
	header.Set("Content-Disposition", multipartFileContentDisposition(file.FieldName, file.FileName))
	header.Set("Content-Type", file.ContentType)
	return writer.CreatePart(header)
}
