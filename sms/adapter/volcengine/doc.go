// Package volcengine 提供火山引擎短信适配器的占位实现。
//
// 当前版本不会创建真实的火山引擎短信客户端，也不会发送短信。
// New、NewWithConfig 和 NewWithOptions 在配置校验通过后返回
// ErrNotImplemented，避免调用方误以为该适配器已经可用于生产。
//
// # 功能特性
//
//   - 提供与 sms.SMS 一致的构造函数入口
//   - 复用短信配置校验，提前发现缺失的访问密钥、地域和签名
//   - 明确返回 ErrNotImplemented，避免占位适配器静默发送失败
//
// # 快速开始
//
// 火山引擎短信适配器当前是占位实现，配置校验通过后会返回 ErrNotImplemented：
//
//	client, err := volcengine.New(
//		volcengine.WithAccessKeyID("your-key-id"),
//		volcengine.WithAccessKeySecret("your-key-secret"),
//		volcengine.WithSignName("your-sign-name"),
//	)
//	_ = client
//	_ = err
package volcengine
