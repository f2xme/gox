package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/f2xme/gox/httpx"
)

func TestValidationResponses(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	app.POST("/missing-key", func(c httpx.Context) error {
		var req struct {
			Other string `validate:"required"`
		}
		return c.BindJSON(&req)
	})
	app.GET("/internal", func(c httpx.Context) error {
		return httpx.ErrInternalError("private detail")
	})
	handler := app.Raw().(http.Handler)
	for _, tt := range []struct {
		name, method, path, accept, body string
		status                           int
		lang, message                    string
	}{
		{"中文", "POST", "/users", "zh-CN", `{}`, 400, "zh", "名称为必填字段"},
		{"英文", "POST", "/users", "en-US", `{}`, 400, "en", "Name is required"},
		{"权重", "POST", "/users", "zh-CN;q=0.1,en;q=0.9", `{}`, 400, "en", "Name is required"},
		{"默认语言", "POST", "/users", "", `{}`, 400, "zh", "名称为必填字段"},
		{"不支持的语言", "POST", "/users", "fr", `{}`, 400, "zh", "名称为必填字段"},
		{"规则参数", "POST", "/users", "en", `{"name":"A"}`, 400, "en", "Name must contain at least 2 characters"},
		{"缺少翻译键", "POST", "/missing-key", "en", `{}`, 400, "en", "Invalid request parameters"},
		{"非法JSON", "POST", "/users", "en", `{`, 400, "", "请求参数无效"},
		{"服务错误", "GET", "/internal", "en", "", 500, "", "Internal Server Error"},
		{"通过校验", "POST", "/users", "zh-CN", `{"name":"张三"}`, 204, "", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept-Language", tt.accept)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.status || rec.Header().Get("Content-Language") != tt.lang {
				t.Fatalf("status=%d, language=%q, body=%s", rec.Code, rec.Header().Get("Content-Language"), rec.Body)
			}
			if tt.status == http.StatusNoContent {
				if rec.Body.Len() != 0 {
					t.Fatalf("意外响应体: %s", rec.Body)
				}
				return
			}
			var result struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Message != tt.message {
				t.Fatalf("message=%q, want %q", result.Message, tt.message)
			}
			if tt.lang != "" && rec.Header().Get("Vary") != "Accept-Language" {
				t.Fatalf("缺少 Vary: Accept-Language")
			}
		})
	}
}
