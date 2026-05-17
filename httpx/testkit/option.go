package testkit

import (
	"net/http"
	"net/url"
)

// Option 配置测试客户端。
type Option func(*Client)

// WithHTTPClient 设置用于发送请求的 HTTP 客户端。
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// WithDefaultHeader 设置所有请求都会携带的默认请求头。
func WithDefaultHeader(key, value string) Option {
	return func(c *Client) {
		c.defaultHeaders.Set(key, value)
	}
}

// WithDefaultHeaders 设置所有请求都会携带的默认请求头。
func WithDefaultHeaders(headers http.Header) Option {
	return func(c *Client) {
		for key, values := range headers {
			for _, value := range values {
				c.defaultHeaders.Add(key, value)
			}
		}
	}
}

type requestConfig struct {
	headers     http.Header
	cookies     []*http.Cookie
	query       url.Values
	contentType string
	basicAuth   *basicAuth
}

type basicAuth struct {
	username string
	password string
}

// RequestOption 配置单次 HTTP 请求。
type RequestOption func(*requestConfig)

// WithHeader 设置单次请求头。
func WithHeader(key, value string) RequestOption {
	return func(c *requestConfig) {
		c.headers.Set(key, value)
	}
}

// WithHeaders 设置单次请求头。
func WithHeaders(headers http.Header) RequestOption {
	return func(c *requestConfig) {
		for key, values := range headers {
			for _, value := range values {
				c.headers.Add(key, value)
			}
		}
	}
}

// WithCookie 添加单次请求 Cookie。
func WithCookie(cookie *http.Cookie) RequestOption {
	return func(c *requestConfig) {
		if cookie != nil {
			c.cookies = append(c.cookies, cookie)
		}
	}
}

// WithQuery 添加单次请求 Query 参数。
func WithQuery(key, value string) RequestOption {
	return func(c *requestConfig) {
		c.query.Add(key, value)
	}
}

// WithQueries 添加单次请求 Query 参数。
func WithQueries(values url.Values) RequestOption {
	return func(c *requestConfig) {
		for key, vals := range values {
			for _, value := range vals {
				c.query.Add(key, value)
			}
		}
	}
}

// WithContentType 设置单次请求的 Content-Type。
func WithContentType(contentType string) RequestOption {
	return func(c *requestConfig) {
		c.contentType = contentType
	}
}

// WithBasicAuth 设置单次请求的 Basic Auth。
func WithBasicAuth(username, password string) RequestOption {
	return func(c *requestConfig) {
		c.basicAuth = &basicAuth{username: username, password: password}
	}
}

// WithBearerToken 设置单次请求的 Bearer Token。
func WithBearerToken(token string) RequestOption {
	return WithHeader("Authorization", "Bearer "+token)
}
