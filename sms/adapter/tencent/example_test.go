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
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("SMS client created: %T\n", sms)

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
		fmt.Printf("Error: %v\n", err)
		return
	}

	_ = sms

	fmt.Println("发送短信示例")
	fmt.Println("sms.Send(\"13800138000\", \"123456\", \"123456\")")

	// Output:
	// 发送短信示例
	// sms.Send("13800138000", "123456", "123456")
}
