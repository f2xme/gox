package sms

// SMS 定义短信服务提供商接口
type SMS interface {
	// Send 发送短信消息
	//
	// 参数：
	//   - phone: 手机号码
	//   - templateCode: 短信模板代码
	//   - templateParam: 模板参数（JSON 格式），例如 {"code":"1234"}
	Send(phone, templateCode, templateParam string) error
}
