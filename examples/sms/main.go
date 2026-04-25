package main

import (
	"context"
	"fmt"

	"github.com/f2xme/gox/sms"
)

// MockSMS 是用于演示的短信服务实现。
//
// 注意：这是一个演示示例，实际使用需要配置真实的短信服务商。
type MockSMS struct{}

// Send 模拟发送短信消息。
func (m *MockSMS) Send(ctx context.Context, message sms.Message) error {
	fmt.Println("  → 发送短信到:", message.Phone)
	fmt.Println("  → 模板代码:", message.TemplateCode)
	fmt.Println("  → 模板参数:", message.TemplateParam)
	fmt.Println("  ✓ 短信发送成功（模拟）")
	return nil
}

func main() {
	fmt.Println("=== sms 包使用示例 ===")
	fmt.Println("注意：这是模拟示例，实际使用需要配置真实的短信服务商")

	var s sms.SMS = &MockSMS{}

	// 示例 1: 发送验证码
	fmt.Println()
	fmt.Println("示例 1: 发送验证码")
	err := s.Send(context.Background(), sms.Message{
		Phone:         "13800138000",
		TemplateCode:  "SMS_123456",
		TemplateParam: map[string]string{"code": "1234"},
	})
	if err != nil {
		fmt.Println("发送失败:", err)
	}

	// 示例 2: 发送订单通知
	fmt.Println()
	fmt.Println("示例 2: 发送订单通知")
	err = s.Send(context.Background(), sms.Message{
		Phone:        "13900139000",
		TemplateCode: "SMS_789012",
		TemplateParam: map[string]string{
			"order_id": "ORDER001",
			"amount":   "99.00",
		},
	})
	if err != nil {
		fmt.Println("发送失败:", err)
	}

	fmt.Println()
	fmt.Println("=== 示例结束 ===")
	fmt.Println()
	fmt.Println("实际使用时，请导入并使用真实的适配器：")
	fmt.Println("- github.com/f2xme/gox/sms/adapter/aliyun")
	fmt.Println("- github.com/f2xme/gox/sms/adapter/tencent")
	fmt.Println("- github.com/f2xme/gox/sms/adapter/volcengine")
}
