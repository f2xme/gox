package main

import (
	"errors"
	"fmt"

	"github.com/f2xme/gox/errorx"
)

func main() {
	fmt.Println("=== Errorx 使用示例 ===")

	// 1. 创建基本错误
	fmt.Println("1. 创建基本错误:")
	err1 := errorx.New("这是一个基本错误")
	fmt.Printf("错误: %v\n", err1)

	// 2. 创建带错误类型的错误
	fmt.Println("\n2. 创建带错误类型的错误:")
	err2 := errorx.NewWithKind(errorx.KindValidation, "用户名不能为空")
	fmt.Printf("错误: %v\n", err2)
	fmt.Printf("错误类型: %s\n", err2.Kind)

	// 3. 创建带错误码的错误
	fmt.Println("\n3. 创建带错误码的错误:")
	err3 := errorx.NewCode("ERR001", "数据库连接失败")
	fmt.Printf("错误: %v\n", err3)

	// 4. 包装现有错误
	fmt.Println("\n4. 包装现有错误:")
	originalErr := errors.New("connection refused")
	wrappedErr := errorx.Wrap(originalErr, "无法连接到数据库")
	fmt.Printf("包装后的错误: %v\n", wrappedErr)

	// 5. 链式设置错误属性
	fmt.Println("\n5. 链式设置错误属性:")
	err4 := errorx.New("用户不存在").
		WithKind(errorx.KindNotFound).
		WithCode("USER_NOT_FOUND").
		WithMetadata("user_id", 12345)
	fmt.Printf("错误: %v\n", err4)
	fmt.Printf("错误类型: %s\n", err4.Kind)
	fmt.Printf("错误码: %s\n", err4.Code)
	fmt.Printf("元数据: %v\n", err4.Metadata)

	// 6. 不同类型的错误
	fmt.Println("\n6. 不同类型的错误示例:")

	// 验证错误
	validationErr := errorx.NewWithKind(errorx.KindValidation, "邮箱格式不正确").
		WithMetadata("field", "email").
		WithMetadata("value", "invalid-email")
	fmt.Printf("验证错误: %v\n", validationErr)

	// 未找到错误
	notFoundErr := errorx.NewWithKind(errorx.KindNotFound, "订单不存在").
		WithCode("ORDER_NOT_FOUND").
		WithMetadata("order_id", "ORD-2024-001")
	fmt.Printf("未找到错误: %v\n", notFoundErr)

	// 冲突错误
	conflictErr := errorx.NewWithKind(errorx.KindConflict, "用户名已存在").
		WithMetadata("username", "zhangsan")
	fmt.Printf("冲突错误: %v\n", conflictErr)

	// 未授权错误
	unauthorizedErr := errorx.NewWithKind(errorx.KindUnauthorized, "未登录或登录已过期")
	fmt.Printf("未授权错误: %v\n", unauthorizedErr)

	// 禁止访问错误
	forbiddenErr := errorx.NewWithKind(errorx.KindForbidden, "没有权限访问此资源")
	fmt.Printf("禁止访问错误: %v\n", forbiddenErr)

	// 7. 错误类型判断
	fmt.Println("\n7. 错误类型判断:")
	fmt.Printf("validationErr 是验证错误: %v\n", validationErr.Kind == errorx.KindValidation)
	fmt.Printf("notFoundErr 是未找到错误: %v\n", notFoundErr.Kind == errorx.KindNotFound)

	// 8. 检查错误是否可重试
	fmt.Println("\n8. 检查错误是否可重试:")
	timeoutErr := errorx.NewWithKind(errorx.KindTimeout, "请求超时")
	retryableErr := errorx.NewWithKind(errorx.KindRetryable, "临时性错误")
	fmt.Printf("超时错误可重试: %v\n", timeoutErr.Kind.IsRetryable())
	fmt.Printf("可重试错误可重试: %v\n", retryableErr.Kind.IsRetryable())
	fmt.Printf("验证错误可重试: %v\n", validationErr.Kind.IsRetryable())

	// 9. 实际应用示例 - 用户注册
	fmt.Println("\n9. 实际应用示例 - 用户注册:")
	if err := registerUser("", "test@example.com"); err != nil {
		if e, ok := err.(*errorx.Error); ok {
			fmt.Printf("注册失败: %v\n", e)
			fmt.Printf("  错误类型: %s\n", e.Kind)
			if e.Metadata != nil {
				fmt.Printf("  错误详情: %v\n", e.Metadata)
			}
		}
	}

	// 10. 错误解包
	fmt.Println("\n10. 错误解包:")
	baseErr := errors.New("底层错误")
	wrapped := errorx.Wrap(baseErr, "包装层 1")
	fmt.Printf("包装的错误: %v\n", wrapped)
	fmt.Printf("解包后的错误: %v\n", errors.Unwrap(wrapped))

	fmt.Println("\n错误处理示例完成")
}

// registerUser 模拟用户注册函数
func registerUser(username, email string) error {
	if username == "" {
		return errorx.NewWithKind(errorx.KindValidation, "用户名不能为空").
			WithMetadata("field", "username")
	}

	if email == "" {
		return errorx.NewWithKind(errorx.KindValidation, "邮箱不能为空").
			WithMetadata("field", "email")
	}

	// 模拟用户名已存在
	if username == "admin" {
		return errorx.NewWithKind(errorx.KindConflict, "用户名已被占用").
			WithCode("USERNAME_EXISTS").
			WithMetadata("username", username)
	}

	return nil
}
