package uni_test

import (
	"fmt"

	"github.com/f2xme/gox/sms/adapter/uni"
)

func ExampleNew() {
	sms, err := uni.New(
		uni.WithAccessKeyID("your-access-key-id"),
		uni.WithAccessKeySecret("your-access-key-secret"),
		uni.WithSignName("UniSMS"),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("SMS client created:", fmt.Sprintf("%T", sms))

	// Output:
	// SMS client created: *uni.uniSMS
}

func ExampleNew_simpleAuth() {
	sms, err := uni.New(
		uni.WithAccessKeyID("your-access-key-id"),
		uni.WithSignName("UniSMS"),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("SMS client created:", fmt.Sprintf("%T", sms))

	// Output:
	// SMS client created: *uni.uniSMS
}

func ExampleNew_send() {
	sms, err := uni.New(
		uni.WithAccessKeyID("your-access-key-id"),
		uni.WithSignName("UniSMS"),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	_ = sms

	fmt.Println("发送短信示例")
	fmt.Println(`sms.Send(ctx, sms.Message{Phone: "13800138000", TemplateCode: "login_tmpl", TemplateParam: map[string]string{"code":"6666"}})`)

	// Output:
	// 发送短信示例
	// sms.Send(ctx, sms.Message{Phone: "13800138000", TemplateCode: "login_tmpl", TemplateParam: map[string]string{"code":"6666"}})
}
