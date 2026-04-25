package sms

import "context"

// Message 定义短信发送请求。
type Message struct {
	// Phone 接收短信的手机号码。
	Phone string
	// TemplateCode 短信模板代码。
	TemplateCode string
	// TemplateParam 短信模板参数。
	//
	// 不同服务商对模板参数的格式要求不同：
	//   - 阿里云适合传入 map、struct、JSON 字符串或 []byte；
	//   - 腾讯云适合传入 []string，也兼容单个 string。
	TemplateParam any
}

// SMS 定义短信服务提供商接口
type SMS interface {
	// Send 发送短信消息
	Send(ctx context.Context, message Message) error
}
