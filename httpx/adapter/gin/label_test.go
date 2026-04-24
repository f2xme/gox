package gin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ginframework "github.com/gin-gonic/gin"
)

// TestRequestWithLabel 带 label tag 的测试请求
type TestRequestWithLabel struct {
	Name  string `json:"name" binding:"required" label:"名字"`
	Age   int    `json:"age" binding:"required,min=1,max=150" label:"年龄"`
	Email string `json:"email" binding:"required,email" label:"邮箱"`
}

func TestBindJSONWithLabel(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]any
		wantErr    bool
		errContain string
	}{
		{
			name: "valid request",
			body: map[string]any{
				"name":  "张三",
				"age":   25,
				"email": "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			body: map[string]any{
				"age":   25,
				"email": "test@example.com",
			},
			wantErr:    true,
			errContain: "名字不能为空",
		},
		{
			name: "age too small",
			body: map[string]any{
				"name":  "张三",
				"age":   0,
				"email": "test@example.com",
			},
			wantErr:    true,
			errContain: "年龄不能小于1",
		},
		{
			name: "age too large",
			body: map[string]any{
				"name":  "张三",
				"age":   200,
				"email": "test@example.com",
			},
			wantErr:    true,
			errContain: "年龄不能大于150",
		},
		{
			name: "invalid email",
			body: map[string]any{
				"name":  "张三",
				"age":   25,
				"email": "invalid-email",
			},
			wantErr:    true,
			errContain: "邮箱必须是有效的邮箱地址",
		},
		{
			name: "missing multiple fields",
			body: map[string]any{
				"age": 25,
			},
			wantErr:    true,
			errContain: "不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin := ginframework.New()
			var capturedErr error

			gin.POST("/test", func(c *ginframework.Context) {
				ctx := &ginContext{c: c}
				var req TestRequestWithLabel
				capturedErr = ctx.BindJSON(&req)
				if capturedErr != nil {
					c.JSON(http.StatusBadRequest, map[string]string{"error": capturedErr.Error()})
					return
				}
				c.JSON(http.StatusOK, req)
			})

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			gin.ServeHTTP(w, req)

			if (capturedErr != nil) != tt.wantErr {
				t.Errorf("BindJSON() error = %v, wantErr %v", capturedErr, tt.wantErr)
				return
			}

			if tt.wantErr && capturedErr != nil {
				errMsg := capturedErr.Error()
				if !strings.Contains(errMsg, tt.errContain) {
					t.Errorf("BindJSON() error message = %v, want to contain %v", errMsg, tt.errContain)
				}
			}
		})
	}
}

func TestBindQueryWithLabel(t *testing.T) {
	gin := ginframework.New()
	var capturedErr error

	gin.GET("/test", func(c *ginframework.Context) {
		ctx := &ginContext{c: c}
		var req TestRequestWithLabel
		capturedErr = ctx.BindQuery(&req)
		if capturedErr != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": capturedErr.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	req := httptest.NewRequest(http.MethodGet, "/test?age=25&email=test@example.com", nil)
	w := httptest.NewRecorder()
	gin.ServeHTTP(w, req)

	if capturedErr == nil {
		t.Error("BindQuery() should return validation error for missing name")
	}

	if !strings.Contains(capturedErr.Error(), "名字不能为空") {
		t.Errorf("BindQuery() error = %v, want to contain '名字不能为空'", capturedErr)
	}
}

func TestTranslateValidationTags(t *testing.T) {
	type TestStruct struct {
		MinStr    string `binding:"min=5" label:"最小字符串"`
		MaxStr    string `binding:"max=10" label:"最大字符串"`
		MinNum    int    `binding:"min=1" label:"最小数字"`
		MaxNum    int    `binding:"max=100" label:"最大数字"`
		Email     string `binding:"email" label:"电子邮件"`
		URL       string `binding:"url" label:"网址"`
		Len       string `binding:"len=6" label:"固定长度"`
		OneOf     string `binding:"oneof=red green blue" label:"颜色"`
		Alpha     string `binding:"alpha" label:"字母"`
		Alphanum  string `binding:"alphanum" label:"字母数字"`
		Numeric   string `binding:"numeric" label:"纯数字"`
		UUID      string `binding:"uuid" label:"UUID"`
		IP        string `binding:"ip" label:"IP地址"`
		IPv4      string `binding:"ipv4" label:"IPv4地址"`
		Contains  string `binding:"contains=test" label:"包含字符"`
		StartWith string `binding:"startswith=pre" label:"前缀"`
		EndWith   string `binding:"endswith=suf" label:"后缀"`
	}

	tests := []struct {
		name       string
		body       map[string]any
		errContain string
	}{
		{
			name:       "min string",
			body:       map[string]any{"MinStr": "abc"},
			errContain: "最小字符串长度不能小于5",
		},
		{
			name:       "max string",
			body:       map[string]any{"MaxStr": "12345678901"},
			errContain: "最大字符串长度不能大于10",
		},
		{
			name:       "min number",
			body:       map[string]any{"MinNum": 0},
			errContain: "最小数字不能小于1",
		},
		{
			name:       "max number",
			body:       map[string]any{"MaxNum": 101},
			errContain: "最大数字不能大于100",
		},
		{
			name:       "invalid email",
			body:       map[string]any{"Email": "not-email"},
			errContain: "电子邮件必须是有效的邮箱地址",
		},
		{
			name:       "invalid url",
			body:       map[string]any{"URL": "not-url"},
			errContain: "网址必须是有效的URL",
		},
		{
			name:       "invalid len",
			body:       map[string]any{"Len": "12345"},
			errContain: "固定长度长度必须为6",
		},
		{
			name:       "invalid oneof",
			body:       map[string]any{"OneOf": "yellow"},
			errContain: "颜色必须是以下值之一",
		},
		{
			name:       "invalid alpha",
			body:       map[string]any{"Alpha": "abc123"},
			errContain: "字母只能包含字母",
		},
		{
			name:       "invalid alphanum",
			body:       map[string]any{"Alphanum": "abc-123"},
			errContain: "字母数字只能包含字母和数字",
		},
		{
			name:       "invalid numeric",
			body:       map[string]any{"Numeric": "12a34"},
			errContain: "纯数字必须是数字",
		},
		{
			name:       "invalid uuid",
			body:       map[string]any{"UUID": "not-uuid"},
			errContain: "UUID必须是有效的UUID",
		},
		{
			name:       "invalid ip",
			body:       map[string]any{"IP": "999.999.999.999"},
			errContain: "IP地址必须是有效的IP地址",
		},
		{
			name:       "invalid ipv4",
			body:       map[string]any{"IPv4": "not-ip"},
			errContain: "IPv4地址必须是有效的IPv4地址",
		},
		{
			name:       "invalid contains",
			body:       map[string]any{"Contains": "hello"},
			errContain: "包含字符必须包含'test'",
		},
		{
			name:       "invalid startswith",
			body:       map[string]any{"StartWith": "hello"},
			errContain: "前缀必须以'pre'开头",
		},
		{
			name:       "invalid endswith",
			body:       map[string]any{"EndWith": "hello"},
			errContain: "后缀必须以'suf'结尾",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin := ginframework.New()
			var capturedErr error

			gin.POST("/test", func(c *ginframework.Context) {
				ctx := &ginContext{c: c}
				var req TestStruct
				capturedErr = ctx.BindJSON(&req)
			})

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			gin.ServeHTTP(w, req)

			if capturedErr == nil {
				t.Errorf("Expected validation error for %s", tt.name)
				return
			}

			errMsg := capturedErr.Error()
			if !strings.Contains(errMsg, tt.errContain) {
				t.Errorf("Error message = %v, want to contain %v", errMsg, tt.errContain)
			}
		})
	}
}
