package validator_test

import (
	"fmt"

	"github.com/f2xme/gox/validator"
	validatorv10 "github.com/go-playground/validator/v10"
)

// ExampleValidate 演示基本验证用法
func ExampleValidate() {
	type User struct {
		Name  string `validate:"required"`
		Email string `validate:"email"`
		Age   int    `validate:"min=18,max=100"`
	}

	// 验证成功的例子
	user := User{
		Name:  "张三",
		Email: "zhangsan@example.com",
		Age:   25,
	}

	if err := validator.Validate(user); err != nil {
		fmt.Println("验证失败:", err)
	} else {
		fmt.Println("验证成功")
	}

	// Output:
	// 验证成功
}

// ExampleValidator_RegisterValidation 演示注册自定义验证规则
func ExampleValidator_RegisterValidation() {
	type Product struct {
		SKU string `validate:"custom_sku"`
	}

	v := validator.New()

	// 注册自定义规则：SKU 必须是 6 位字母数字组合
	v.RegisterValidation("custom_sku", func(fl validatorv10.FieldLevel) bool {
		sku := fl.Field().String()
		if len(sku) != 6 {
			return false
		}
		for _, r := range sku {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
				return false
			}
		}
		return true
	})

	// 验证有效的 SKU
	product := Product{SKU: "ABC123"}
	if err := v.Validate(product); err != nil {
		fmt.Println("验证失败:", err)
	} else {
		fmt.Println("SKU 验证成功")
	}

	// Output:
	// SKU 验证成功
}

// ExampleNew 演示创建验证器实例
func ExampleNew() {
	type Config struct {
		Host string `validate:"required,hostname"`
		Port int    `validate:"required,min=1,max=65535"`
	}

	v := validator.New()

	config := Config{
		Host: "localhost",
		Port: 8080,
	}

	if err := v.Validate(config); err != nil {
		fmt.Println("配置验证失败:", err)
	} else {
		fmt.Println("配置验证成功")
	}

	// Output:
	// 配置验证成功
}

// ExampleRegisterValidation 演示全局注册自定义验证规则
func ExampleRegisterValidation() {
	type User struct {
		Username string `validate:"username_rule"`
	}

	// 全局注册自定义规则：用户名必须以字母开头
	validator.RegisterValidation("username_rule", func(fl validatorv10.FieldLevel) bool {
		username := fl.Field().String()
		if len(username) == 0 {
			return false
		}
		first := username[0]
		return (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')
	})

	user := User{Username: "alice"}
	if err := validator.Validate(user); err != nil {
		fmt.Println("验证失败:", err)
	} else {
		fmt.Println("用户名验证成功")
	}

	// Output:
	// 用户名验证成功
}
