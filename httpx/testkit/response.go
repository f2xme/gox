package testkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// Response 表示一次测试请求的完整响应。
type Response struct {
	t testing.TB
	// Raw 是原始 HTTP 响应。
	Raw *http.Response
	// Body 是已读取的完整响应体。
	Body []byte
}

// ExpectStatus 断言响应状态码。
func (r *Response) ExpectStatus(code int) *Response {
	r.t.Helper()
	if r.Raw.StatusCode != code {
		r.t.Fatalf("status = %d, want %d, body = %s", r.Raw.StatusCode, code, string(r.Body))
	}
	return r
}

// ExpectHeader 断言响应头等于指定值。
func (r *Response) ExpectHeader(key, value string) *Response {
	r.t.Helper()
	if got := r.Raw.Header.Get(key); got != value {
		r.t.Fatalf("header %q = %q, want %q", key, got, value)
	}
	return r
}

// ExpectHeaderContains 断言响应头包含指定片段。
func (r *Response) ExpectHeaderContains(key, value string) *Response {
	r.t.Helper()
	if got := r.Raw.Header.Get(key); !strings.Contains(got, value) {
		r.t.Fatalf("header %q = %q, want contains %q", key, got, value)
	}
	return r
}

// ExpectCookie 断言响应 Set-Cookie 中存在指定 Cookie 值。
func (r *Response) ExpectCookie(name, value string) *Response {
	r.t.Helper()
	for _, cookie := range r.Raw.Cookies() {
		if cookie.Name == name {
			if cookie.Value != value {
				r.t.Fatalf("cookie %q = %q, want %q", name, cookie.Value, value)
			}
			return r
		}
	}
	r.t.Fatalf("cookie %q not found", name)
	return r
}

// ExpectBody 断言响应体等于指定字符串。
func (r *Response) ExpectBody(body string) *Response {
	r.t.Helper()
	if got := string(r.Body); got != body {
		r.t.Fatalf("body = %q, want %q", got, body)
	}
	return r
}

// ExpectBodyContains 断言响应体包含指定片段。
func (r *Response) ExpectBodyContains(value string) *Response {
	r.t.Helper()
	if !bytes.Contains(r.Body, []byte(value)) {
		r.t.Fatalf("body = %q, want contains %q", string(r.Body), value)
	}
	return r
}

// DecodeJSON 将响应体解码到 v。
func (r *Response) DecodeJSON(v any) *Response {
	r.t.Helper()
	if err := json.Unmarshal(r.Body, v); err != nil {
		r.t.Fatalf("decode json response: %v, body = %s", err, string(r.Body))
	}
	return r
}

// ExpectJSONValue 断言 JSON 响应中指定路径的值。
//
// path 使用点号和数组下标表示，例如 "success"、"data.id"、"items[0].name"。
func (r *Response) ExpectJSONValue(path string, want any) *Response {
	r.t.Helper()

	var root any
	if err := json.Unmarshal(r.Body, &root); err != nil {
		r.t.Fatalf("decode json response: %v, body = %s", err, string(r.Body))
	}

	got, ok := jsonPath(root, path)
	if !ok {
		r.t.Fatalf("json path %q not found, body = %s", path, string(r.Body))
	}
	if !valuesEqual(got, want) {
		r.t.Fatalf("json path %q = %#v, want %#v", path, got, want)
	}
	return r
}

func jsonPath(root any, path string) (any, bool) {
	if path == "" {
		return root, true
	}
	current := root
	for _, part := range strings.Split(path, ".") {
		if part == "" {
			return nil, false
		}
		name, indexes, ok := parsePathPart(part)
		if !ok {
			return nil, false
		}
		if name != "" {
			obj, ok := current.(map[string]any)
			if !ok {
				return nil, false
			}
			current, ok = obj[name]
			if !ok {
				return nil, false
			}
		}
		for _, idx := range indexes {
			arr, ok := current.([]any)
			if !ok || idx < 0 || idx >= len(arr) {
				return nil, false
			}
			current = arr[idx]
		}
	}
	return current, true
}

func parsePathPart(part string) (string, []int, bool) {
	nameEnd := strings.Index(part, "[")
	if nameEnd == -1 {
		return part, nil, true
	}
	name := part[:nameEnd]
	rest := part[nameEnd:]
	var indexes []int
	for rest != "" {
		if !strings.HasPrefix(rest, "[") {
			return "", nil, false
		}
		end := strings.Index(rest, "]")
		if end < 0 {
			return "", nil, false
		}
		idx, err := strconv.Atoi(rest[1:end])
		if err != nil {
			return "", nil, false
		}
		indexes = append(indexes, idx)
		rest = rest[end+1:]
	}
	return name, indexes, true
}

func valuesEqual(got, want any) bool {
	if reflect.DeepEqual(got, want) {
		return true
	}
	g, gok := toFloat64(got)
	w, wok := toFloat64(want)
	if gok && wok {
		return math.Abs(g-w) < 1e-9
	}
	return fmt.Sprint(got) == fmt.Sprint(want)
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}
