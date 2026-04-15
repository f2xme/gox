package main

import (
	"fmt"
	"log"

	"github.com/f2xme/gox/validator"
	validatorv10 "github.com/go-playground/validator/v10"
)

// User 用户注册信息
type User struct {
	Username string `validate:"required,min=3,max=20"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	Age      int    `validate:"required,gte=18,lte=120"`
	Phone    string `validate:"phone"` // 自定义验证规则
}

func main() {
	fmt.Println("=== validator 包使用示例 ===\n")

	// 注册自定义验证规则：手机号验证
	registerCustomRules()

	// 示例 1: 验证成功的用户
	fmt.Println("示例 1: 验证有效的用户数据")
	validUser := User{
		Username: "zhangsan",
		Email:    "zhangsan@example.com",
		Password: "password123",
		Age:      25,
		Phone:    "13800138000",
	}
	validateUser(validUser)

	// 示例 2: 验证失败 - 缺少必填字段
	fmt.Println("\n示例 2: 缺少必填字段")
	invalidUser1 := User{
		Username: "",
		Email:    "test@example.com",
		Password: "pass",
		Age:      25,
	}
	validateUser(invalidUser1)

	// 示例 3: 验证失败 - 邮箱格式错误
	fmt.Println("\n示例 3: 邮箱格式错误")
	invalidUser2 := User{
		Username: "lisi",
		Email:    "invalid-email",
		Password: "password123",
		Age:      25,
		Phone:    "13800138000",
	}
	validateUser(invalidUser2)

	// 示例 4: 验证失败 - 年龄超出范围
	fmt.Println("\n示例 4: 年龄超出范围")
	invalidUser3 := User{
		Username: "wangwu",
		Email:    "wangwu@example.com",
		Password: "password123",
		Age:      15, // 小于 18 岁
		Phone:    "13800138000",
	}
	validateUser(invalidUser3)

	// 示例 5: 验证失败 - 自定义规则（手机号格式错误）
	fmt.Println("\n示例 5: 手机号格式错误")
	invalidUser4 := User{
		Username: "zhaoliu",
		Email:    "zhaoliu@example.com",
		Password: "password123",
		Age:      30,
		Phone:    "12345", // 无效的手机号
	}
	validateUser(invalidUser4)

	// 示例 6: 使用独立的验证器实例
	fmt.Println("\n示例 6: 使用独立的验证器实例")
	customValidator := validator.New()

	// 为这个实例注册特殊的验证规则
	customValidator.RegisterValidation("strong_password", func(fl validatorv10.FieldLevel) bool {
		password := fl.Field().String()
		// 强密码：至少包含一个大写字母、一个小写字母、一个数字
		hasUpper, hasLower, hasDigit := false, false, false
		for _, r := range password {
			if r >= 'A' && r <= 'Z' {
				hasUpper = true
			} else if r >= 'a' && r <= 'z' {
				hasLower = true
			} else if r >= '0' && r <= '9' {
				hasDigit = true
			}
		}
		return hasUpper && hasLower && hasDigit && len(password) >= 8
	})

	type SecureUser struct {
		Username string `validate:"required"`
		Password string `validate:"strong_password"`
	}

	secureUser := SecureUser{
		Username: "admin",
		Password: "Password123", // 符合强密码规则
	}

	if err := customValidator.Validate(secureUser); err != nil {
		fmt.Printf("❌ 验证失败: %v\n", err)
	} else {
		fmt.Println("✅ 强密码验证成功")
	}

	fmt.Println("\n=== 示例结束 ===")
}

// registerCustomRules 注册自定义验证规则
func registerCustomRules() {
	// 注册手机号验证规则（简化版：11位数字）
	err := validator.RegisterValidation("phone", func(fl validatorv10.FieldLevel) bool {
		phone := fl.Field().String()
		if len(phone) != 11 {
			return false
		}
		// 检查是否全是数字
		for _, r := range phone {
			if r < '0' || r > '9' {
				return false
			}
		}
		// 检查是否以 1 开头
		return phone[0] == '1'
	})

	if err != nil {
		log.Fatalf("注册自定义验证规则失败: %v", err)
	}

	// 注册自定义错误消息
	err = validator.RegisterTranslation("phone", "手机号格式不正确（应为11位数字且以1开头）")
	if err != nil {
		log.Fatalf("注册自定义翻译失败: %v", err)
	}
}

// validateUser 验证用户并打印结果
func validateUser(user User) {
	err := validator.Validate(user)
	if err != nil {
		fmt.Printf("❌ 验证失败: %v\n", err)
	} else {
		fmt.Printf("✅ 验证成功: 用户 %s 的数据有效\n", user.Username)
	}
}
