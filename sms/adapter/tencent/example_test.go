package tencent_test

import (
	"fmt"

	"github.com/f2xme/gox/sms/adapter/tencent"
)

func ExampleNew() {
	sms, err := tencent.New(
		tencent.WithSecretID("your-secret-id"),
		tencent.WithSecretKey("your-secret-key"),
		tencent.WithAppID("your-app-id"),
		tencent.WithSignName("测试签名"),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("SMS client created:", fmt.Sprintf("%T", sms))

	// Output:
	// SMS client created: *tencent.tencentSMS
}

func ExampleNew_send() {
	sms, err := tencent.New(
		tencent.WithSecretID("your-secret-id"),
		tencent.WithSecretKey("your-secret-key"),
		tencent.WithAppID("your-app-id"),
		tencent.WithSignName("测试签名"),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	_ = sms

	fmt.Println("发送短信示例")
	fmt.Println(`sms.Send(ctx, sms.Message{Phone: "+8613800138000", TemplateCode: "123456", TemplateParam: []string{"123456"}})`)

	// Output:
	// 发送短信示例
	// sms.Send(ctx, sms.Message{Phone: "+8613800138000", TemplateCode: "123456", TemplateParam: []string{"123456"}})
}
