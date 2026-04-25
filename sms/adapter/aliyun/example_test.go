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
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("SMS client created: %T\n", sms)

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
		fmt.Printf("Error: %v\n", err)
		return
	}

	_ = sms

	fmt.Println("发送短信示例")
	fmt.Println("sms.Send(\"13800138000\", \"SMS_123456789\", `{\"code\":\"123456\"}`)")

	// Output:
	// 发送短信示例
	// sms.Send("13800138000", "SMS_123456789", `{"code":"123456"}`)
}
