package gin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/f2xme/gox/httpx"
	goxvalidator "github.com/f2xme/gox/validator"
	ginframework "github.com/gin-gonic/gin"
)

// TestRequest 测试请求结构体,实现了 Validator 接口
type TestRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (r *TestRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.Age < 0 || r.Age > 150 {
		return fmt.Errorf("age must be between 0 and 150")
	}
	return nil
}

// TestRequestNoValidator 不实现 Validator 接口的请求结构体
type TestRequestNoValidator struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type adapterValidateRequest struct {
	Name string `json:"name" validate:"required"`
}

func TestGoxValidatorAdapterValidateStruct(t *testing.T) {
	adapter := &goxValidatorAdapter{validator: newGoxValidator()}
	valid := adapterValidateRequest{Name: "Alice"}
	invalid := adapterValidateRequest{}
	validPtr := &valid

	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name:    "nil",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "non struct",
			input:   "Alice",
			wantErr: false,
		},
		{
			name:    "valid struct",
			input:   valid,
			wantErr: false,
		},
		{
			name:    "invalid struct",
			input:   invalid,
			wantErr: true,
		},
		{
			name:    "valid pointer",
			input:   &valid,
			wantErr: false,
		},
		{
			name:    "nil pointer",
			input:   (*adapterValidateRequest)(nil),
			wantErr: false,
		},
		{
			name:    "valid nested pointer",
			input:   &validPtr,
			wantErr: false,
		},
		{
			name:    "valid slice",
			input:   []adapterValidateRequest{valid},
			wantErr: false,
		},
		{
			name:    "invalid slice element",
			input:   []adapterValidateRequest{valid, invalid},
			wantErr: true,
		},
		{
			name:    "invalid array element",
			input:   [2]adapterValidateRequest{valid, invalid},
			wantErr: true,
		},
		{
			name:    "valid interface",
			input:   any(valid),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ValidateStruct(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGoxValidatorAdapterEngine(t *testing.T) {
	adapter := &goxValidatorAdapter{validator: newGoxValidator()}
	if _, ok := adapter.Engine().(*goxvalidator.Validator); !ok {
		t.Fatalf("Engine() = %T, want *validator.Validator", adapter.Engine())
	}
}

func TestBindJSONWithValidator(t *testing.T) {
	tests := []struct {
		name    string
		body    TestRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid request",
			body:    TestRequest{Name: "Alice", Age: 25},
			wantErr: false,
		},
		{
			name:    "empty name",
			body:    TestRequest{Name: "", Age: 25},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name:    "negative age",
			body:    TestRequest{Name: "Bob", Age: -1},
			wantErr: true,
			errMsg:  "age must be between 0 and 150",
		},
		{
			name:    "age too large",
			body:    TestRequest{Name: "Charlie", Age: 200},
			wantErr: true,
			errMsg:  "age must be between 0 and 150",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin := ginframework.New()
			gin.POST("/test", func(c *ginframework.Context) {
				ctx := &ginContext{c: c}
				var req TestRequest
				err := ctx.BindJSON(&req)
				if (err != nil) != tt.wantErr {
					t.Errorf("BindJSON() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.wantErr && err.Error() != tt.errMsg {
					t.Errorf("BindJSON() error message = %v, want %v", err.Error(), tt.errMsg)
				}
			})

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			gin.ServeHTTP(w, req)
		})
	}
}

func TestBindJSONWithoutValidator(t *testing.T) {
	gin := ginframework.New()
	gin.POST("/test", func(c *ginframework.Context) {
		ctx := &ginContext{c: c}
		var req TestRequestNoValidator
		err := ctx.BindJSON(&req)
		if err != nil {
			t.Errorf("BindJSON() should not return error for struct without Validator, got %v", err)
		}
		if req.Name != "Test" || req.Age != 30 {
			t.Errorf("BindJSON() failed to bind data correctly")
		}
	})

	body, _ := json.Marshal(TestRequestNoValidator{Name: "Test", Age: 30})
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	gin.ServeHTTP(w, req)
}

func TestBindQueryWithValidator(t *testing.T) {
	gin := ginframework.New()
	gin.GET("/test", func(c *ginframework.Context) {
		ctx := &ginContext{c: c}
		var req TestRequest
		err := ctx.BindQuery(&req)
		if err == nil {
			t.Error("BindQuery() should return validation error for empty name")
		}
		if err.Error() != "name is required" {
			t.Errorf("BindQuery() error = %v, want 'name is required'", err)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/test?age=25", nil)
	w := httptest.NewRecorder()
	gin.ServeHTTP(w, req)
}

func TestValidateFunction(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name:    "with validator - valid",
			input:   &TestRequest{Name: "Alice", Age: 25},
			wantErr: false,
		},
		{
			name:    "with validator - invalid",
			input:   &TestRequest{Name: "", Age: 25},
			wantErr: true,
		},
		{
			name:    "without validator",
			input:   &TestRequestNoValidator{Name: "", Age: 25},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if validator, ok := tt.input.(httpx.Validator); ok {
				err = validator.Validate()
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAllBindMethodsWithValidator(t *testing.T) {
	t.Run("Bind", func(t *testing.T) {
		gin := ginframework.New()
		gin.POST("/test", func(c *ginframework.Context) {
			ctx := &ginContext{c: c}
			var req TestRequest
			err := ctx.Bind(&req)
			if err == nil {
				t.Error("Bind() should return validation error")
			}
		})

		body, _ := json.Marshal(TestRequest{Name: "", Age: 25})
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		gin.ServeHTTP(w, req)
	})

	t.Run("BindForm", func(t *testing.T) {
		gin := ginframework.New()
		gin.POST("/test", func(c *ginframework.Context) {
			ctx := &ginContext{c: c}
			var req TestRequest
			err := ctx.BindForm(&req)
			if err == nil {
				t.Error("BindForm() should return validation error")
			}
		})

		req := httptest.NewRequest(http.MethodPost, "/test?name=&age=25", nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		gin.ServeHTTP(w, req)
	})
}

// 确保 validate 函数对 Validator 接口的类型断言正确
func TestValidatorInterfaceAssertion(t *testing.T) {
	var _ httpx.Validator = (*TestRequest)(nil)
}
