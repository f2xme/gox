package main

import (
	"fmt"

	"github.com/f2xme/gox/sms"
)

// 注意：这是一个演示示例，实际使用需要配置真实的短信服务商
type MockSMS struct{}

func (m *MockSMS) Send(phone, templateCode, templateParam string) error {
	fmt.Printf("  → 发送短信到: %s\n", phone)
	fmt.Printf("  → 模板代码: %s\n", templateCode)
	fmt.Printf("  → 模板参数: %s\n", templateParam)
	fmt.Println("  ✓ 短信发送成功（模拟）")
	return nil
}

func main() {
	fmt.Println("=== sms 包使用示例 ===")
	fmt.Println("注意：这是模拟示例，实际使用需要配置真实的短信服务商")

	var s sms.SMS = &MockSMS{}

	// 示例 1: 发送验证码
	fmt.Println("\n示例 1: 发送验证码")
	err := s.Send("13800138000", "SMS_123456", `{"code":"1234"}`)
	if err != nil {
		fmt.Printf("发送失败: %v\n", err)
	}

	// 示例 2: 发送订单通知
	fmt.Println("\n示例 2: 发送订单通知")
	err = s.Send("13900139000", "SMS_789012", `{"order_id":"ORDER001","amount":"99.00"}`)
	if err != nil {
		fmt.Printf("发送失败: %v\n", err)
	}

	fmt.Println("\n=== 示例结束 ===")
	fmt.Println("\n实际使用时，请导入并使用真实的适配器：")
	fmt.Println("- github.com/f2xme/gox/sms/adapter/aliyun")
	fmt.Println("- github.com/f2xme/gox/sms/adapter/tencent")
	fmt.Println("- github.com/f2xme/gox/sms/adapter/volcengine")
}
