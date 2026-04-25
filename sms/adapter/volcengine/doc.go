// Package volcengine 提供基于火山引擎的短信服务实现。
//
// # 功能特性
//
//   - 基于火山引擎短信服务 SDK 封装
//   - 支持模板短信发送
//   - 统一的错误处理
//   - 支持配置文件初始化
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"log"
//
//		"github.com/f2xme/gox/sms/volcengine"
//	)
//
//	func main() {
//		// 创建火山引擎短信客户端
//		client, err := volcengine.New(
//			volcengine.WithAccessKeyID("your-access-key-id"),
//			volcengine.WithAccessKeySecret("your-access-key-secret"),
//			volcengine.WithSignName("your-sign-name"),
//		)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// 发送短信
//		err = client.Send(
//			"13800138000",
//			"SMS_123456789",
//			`{"code":"1234"}`,
//		)
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//
// # 使用配置文件
//
// 从配置文件初始化：
//
//	import (
//		"github.com/f2xme/gox/config"
//		"github.com/f2xme/gox/sms/volcengine"
//	)
//
//	func main() {
//		cfg := config.New()
//		client, err := volcengine.NewWithConfig(cfg)
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//
// 配置文件示例（YAML）：
//
//	sms:
//	  volcengine:
//	    accessKeyID: "your-access-key-id"
//	    accessKeySecret: "your-access-key-secret"
//	    region: "cn-north-1"
//	    signName: "your-sign-name"
//
// # 注意事项
//
//   - 确保火山引擎账号已开通短信服务
//   - 短信签名和模板需要在火山引擎控制台预先配置
//   - templateParam 参数必须是有效的 JSON 字符串
//   - 当前实现为占位符，实际发送功能待实现
package volcengine
