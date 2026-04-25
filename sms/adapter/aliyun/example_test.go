package aliyun_test

import (
	"fmt"

	"github.com/f2xme/gox/sms/adapter/aliyun"
)

func ExampleNew() {
	sms, err := aliyun.New(
		aliyun.WithAccessKeyID("your-access-key-id"),
		aliyun.WithAccessKeySecret("your-access-key-secret"),
		aliyun.WithSignName("测试签名"),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("SMS client created:", fmt.Sprintf("%T", sms))

	// Output:
	// SMS client created: *aliyun.aliyunSMS
}

func ExampleNew_send() {
	sms, err := aliyun.New(
		aliyun.WithAccessKeyID("your-access-key-id"),
		aliyun.WithAccessKeySecret("your-access-key-secret"),
		aliyun.WithSignName("测试签名"),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	_ = sms

	fmt.Println("发送短信示例")
	fmt.Println(`sms.Send(ctx, sms.Message{Phone: "13800138000", TemplateCode: "SMS_123456789", TemplateParam: map[string]string{"code":"123456"}})`)

	// Output:
	// 发送短信示例
	// sms.Send(ctx, sms.Message{Phone: "13800138000", TemplateCode: "SMS_123456789", TemplateParam: map[string]string{"code":"123456"}})
}
