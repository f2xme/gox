// Package uni 提供基于 UniSMS 的短信服务适配器。
//
// 该包实现了 sms.SMS 接口，封装 UniSMS 官方 Go SDK，
// 用于通过统一的 gox 短信接口发送模板短信。
//
// # 功能特性
//
//   - 支持 UniSMS 模板短信发送
//   - 支持 AccessKeySecret 完整验签模式
//   - 支持 AccessKeySecret 为空的简易验签模式
//   - 支持 Options 模式配置客户端
//   - 支持通过 config.Config 读取配置
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"context"
//
//		"github.com/f2xme/gox/sms"
//		"github.com/f2xme/gox/sms/adapter/uni"
//	)
//
//	func main() {
//		client, err := uni.New(
//			uni.WithAccessKeyID("your-access-key-id"),
//			uni.WithAccessKeySecret("your-access-key-secret"),
//			uni.WithSignName("UniSMS"),
//		)
//		if err != nil {
//			return
//		}
//
//		err = client.Send(context.Background(), sms.Message{
//			Phone:        "13800138000",
//			TemplateCode: "login_tmpl",
//			TemplateParam: map[string]string{
//				"code": "6666",
//			},
//		})
//		if err != nil {
//			return
//		}
//	}
//
// # 简易验签模式
//
// 如果 UniSMS 账号使用简易验签模式，可以不设置 AccessKeySecret：
//
//	client, err := uni.New(
//		uni.WithAccessKeyID("your-access-key-id"),
//		uni.WithSignName("UniSMS"),
//	)
//
// # 注意事项
//
//   - sms.Message.TemplateCode 对应 UniSMS 的模板 ID
//   - sms.Message.TemplateParam 支持 map[string]string 和 map[string]any
//   - UniSMS 官方 SDK 的发送方法不接收 context，发送前会检查 context 状态
package uni
