package validator

import (
	"strings"
	"sync"
	"testing"

	"github.com/go-playground/validator/v10"
)

// 测试 1: 创建默认验证器实例成功
func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New() returned nil")
	}
	if v.validate == nil {
		t.Fatal("validator instance is nil")
	}
}

// 测试 New 创建的独立实例包含内置中国本地化验证规则
func TestNew_RegistersBuiltinValidations(t *testing.T) {
	type User struct {
		Phone string `validate:"phone"`
	}

	v := New()
	err := v.Validate(User{Phone: "13800138000"})
	if err != nil {
		t.Errorf("expected no error for valid phone, got: %v", err)
	}

	err = v.Validate(User{Phone: "123"})
	if err == nil {
		t.Fatal("expected error for invalid phone, got nil")
	}
	if !strings.Contains(err.Error(), "手机号") {
		t.Errorf("expected Chinese error message containing '手机号', got: %v", err)
	}
}

// 测试默认使用 label 标签作为错误消息字段名
func TestNew_UsesDefaultLabelTag(t *testing.T) {
	type User struct {
		Name string `validate:"required" label:"姓名"`
	}

	v := New()
	err := v.Validate(User{})
	if err == nil {
		t.Fatal("expected error for missing required field, got nil")
	}
	if !strings.Contains(err.Error(), "姓名") {
		t.Errorf("error should mention label field name, got: %v", err)
	}
}

// 测试可通过 Options 自定义错误消息字段名标签
func TestNew_WithFieldNameTag(t *testing.T) {
	type User struct {
		Name string `json:"name,omitempty" validate:"required"`
	}

	v := New(WithFieldNameTag("json"))
	err := v.Validate(User{})
	if err == nil {
		t.Fatal("expected error for missing required field, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention json field name, got: %v", err)
	}
}

// 测试空字段名标签配置回退到结构体字段名
func TestNew_WithEmptyFieldNameTag(t *testing.T) {
	type User struct {
		Name string `json:"name,omitempty" validate:"required"`
	}

	v := New(WithFieldNameTag(""))
	err := v.Validate(User{})
	if err == nil {
		t.Fatal("expected error for missing required field, got nil")
	}
	if !strings.Contains(err.Error(), "Name") {
		t.Errorf("error should mention struct field name, got: %v", err)
	}
}

// 测试 2: 验证带 required 标签的结构体，缺失字段返回错误
func TestValidate_Required(t *testing.T) {
	type User struct {
		Name string `validate:"required"`
	}

	v := New()
	user := User{Name: ""}
	err := v.Validate(user)
	if err == nil {
		t.Fatal("expected error for missing required field, got nil")
	}
	if !strings.Contains(err.Error(), "Name") {
		t.Errorf("error should mention field Name, got: %v", err)
	}
}

// 测试 3: 验证带 email 标签的字段，无效邮箱返回错误
func TestValidate_Email(t *testing.T) {
	type User struct {
		Email string `validate:"email"`
	}

	v := New()
	user := User{Email: "invalid-email"}
	err := v.Validate(user)
	if err == nil {
		t.Fatal("expected error for invalid email, got nil")
	}
	if !strings.Contains(err.Error(), "Email") {
		t.Errorf("error should mention field Email, got: %v", err)
	}
}

// 测试 4: 验证带 min/max 标签的字段，超出范围返回错误
func TestValidate_MinMax(t *testing.T) {
	type User struct {
		Age int `validate:"min=18,max=100"`
	}

	tests := []struct {
		name    string
		age     int
		wantErr bool
	}{
		{"age too low", 10, true},
		{"age too high", 150, true},
		{"age valid", 25, false},
	}

	v := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{Age: tt.age}
			err := v.Validate(user)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// 测试 5: 注册自定义验证规则并成功验证
func TestRegisterValidation(t *testing.T) {
	type User struct {
		Username string `validate:"custom_username"`
	}

	v := New()
	// 注册自定义规则：用户名必须以字母开头
	err := v.RegisterValidation("custom_username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		if len(username) == 0 {
			return false
		}
		first := username[0]
		return (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')
	})
	if err != nil {
		t.Fatalf("RegisterValidation() error = %v", err)
	}

	// 测试无效用户名
	user1 := User{Username: "123abc"}
	err = v.Validate(user1)
	if err == nil {
		t.Error("expected error for username starting with number, got nil")
	}

	// 测试有效用户名
	user2 := User{Username: "abc123"}
	err = v.Validate(user2)
	if err != nil {
		t.Errorf("expected no error for valid username, got: %v", err)
	}
}

// 测试 6: 验证成功返回 nil
func TestValidate_Success(t *testing.T) {
	type User struct {
		Name  string `validate:"required"`
		Email string `validate:"email"`
		Age   int    `validate:"min=18,max=100"`
	}

	v := New()
	user := User{
		Name:  "张三",
		Email: "zhangsan@example.com",
		Age:   25,
	}
	err := v.Validate(user)
	if err != nil {
		t.Errorf("expected no error for valid user, got: %v", err)
	}
}

// 测试 7: 错误消息支持中文翻译
func TestValidate_ChineseError(t *testing.T) {
	type User struct {
		Name string `validate:"required"`
	}

	v := New()
	user := User{Name: ""}
	err := v.Validate(user)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errMsg := err.Error()
	// 检查是否包含中文字符（中文翻译的标志）
	hasChinese := false
	for _, r := range errMsg {
		if r >= 0x4e00 && r <= 0x9fff {
			hasChinese = true
			break
		}
	}
	if !hasChinese {
		t.Errorf("expected Chinese error message, got: %v", errMsg)
	}
}

// 测试全局快捷函数
func TestValidate_GlobalFunction(t *testing.T) {
	type User struct {
		Name string `validate:"required"`
	}

	user := User{Name: "测试"}
	err := Validate(user)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	user2 := User{Name: ""}
	err = Validate(user2)
	if err == nil {
		t.Error("expected error for missing required field, got nil")
	}
}

// 测试全局注册自定义验证
func TestRegisterValidation_Global(t *testing.T) {
	type Product struct {
		SKU string `validate:"test_sku"`
	}

	// 注册全局自定义规则
	err := RegisterValidation("test_sku", func(fl validator.FieldLevel) bool {
		sku := fl.Field().String()
		return len(sku) >= 3
	})
	if err != nil {
		t.Fatalf("RegisterValidation() error = %v", err)
	}

	product := Product{SKU: "AB"}
	err = Validate(product)
	if err == nil {
		t.Error("expected error for short SKU, got nil")
	}

	product2 := Product{SKU: "ABC123"}
	err = Validate(product2)
	if err != nil {
		t.Errorf("expected no error for valid SKU, got: %v", err)
	}
}

// 测试验证和注册操作并发使用时不会产生数据竞态
func TestValidator_ConcurrentValidateAndRegister(t *testing.T) {
	type User struct {
		Name string `validate:"required"`
	}

	v := New()
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = v.Validate(User{Name: "张三"})
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := v.RegisterValidation("concurrent_rule", func(fl validator.FieldLevel) bool {
			return true
		})
		if err != nil {
			t.Errorf("RegisterValidation() error = %v", err)
		}
	}()

	wg.Wait()
}

// 测试手机号验证器
func TestValidate_Phone(t *testing.T) {
	type User struct {
		Phone string `validate:"phone"`
	}

	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{"valid phone 138", "13800138000", false},
		{"valid phone 186", "18612345678", false},
		{"invalid 10 digits", "1234567890", true},
		{"invalid not start with 1", "28612345678", true},
		{"invalid contains letter", "1861234567a", true},
		{"invalid empty", "", true},
		{"invalid 12 digits", "138001380001", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{Phone: tt.phone}
			err := Validate(user)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "手机号") {
				t.Errorf("expected Chinese error message containing '手机号', got: %v", err)
			}
		})
	}
}

// 测试身份证号验证器
func TestValidate_IDCard(t *testing.T) {
	type User struct {
		IDCard string `validate:"id_card"`
	}

	tests := []struct {
		name    string
		idCard  string
		wantErr bool
	}{
		{"valid id card with X", "11010519491231002X", false},
		{"valid id card with digit", "110105194912310038", false},
		{"invalid 17 digits", "12345678901234567", true},
		{"invalid checksum", "110105194912310028", true},
		{"invalid contains letter in body", "1101051949123a002X", true},
		{"invalid empty", "", true},
		{"invalid 19 digits", "1101051949123100281", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{IDCard: tt.idCard}
			err := Validate(user)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "身份证") {
				t.Errorf("expected Chinese error message containing '身份证', got: %v", err)
			}
		})
	}
}

// 测试银行卡号验证器
func TestValidate_BankCard(t *testing.T) {
	type Payment struct {
		CardNumber string `validate:"bank_card"`
	}

	tests := []struct {
		name       string
		cardNumber string
		wantErr    bool
	}{
		{"valid card 19 digits", "6222021001234567896", false},
		{"valid card 16 digits", "6222021001234561", false},
		{"invalid luhn", "6222021001234567891", true},
		{"invalid contains letter", "622202100123456789a", true},
		{"invalid empty", "", true},
		{"invalid 12 digits", "622202100123", true},
		{"invalid 20 digits", "62220210012345678901", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := Payment{CardNumber: tt.cardNumber}
			err := Validate(payment)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "银行卡") {
				t.Errorf("expected Chinese error message containing '银行卡', got: %v", err)
			}
		})
	}
}
