// Package tencent 提供基于腾讯云的短信服务实现。
//
// # 功能特性
//
//   - 基于腾讯云短信服务 SDK 封装
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
//		"github.com/f2xme/gox/sms/tencent"
//	)
//
//	func main() {
//		// 创建腾讯云短信客户端
//		client, err := tencent.New(
//			tencent.WithSecretID("your-secret-id"),
//			tencent.WithSecretKey("your-secret-key"),
//			tencent.WithAppID("your-app-id"),
//			tencent.WithSignName("your-sign-name"),
//		)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// 发送短信
//		err = client.Send(
//			"13800138000",
//			"123456",
//			"1234",
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
//		"github.com/f2xme/gox/sms/tencent"
//	)
//
//	func main() {
//		cfg := config.New()
//		client, err := tencent.NewWithConfig(cfg)
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//
// 配置文件示例（YAML）：
//
//	sms:
//	  tencent:
//	    secretID: "your-secret-id"
//	    secretKey: "your-secret-key"
//	    region: "ap-guangzhou"
//	    appID: "your-app-id"
//	    signName: "your-sign-name"
//
// # 注意事项
//
//   - 确保腾讯云账号已开通短信服务
//   - 短信签名和模板需要在腾讯云控制台预先配置
//   - templateParam 参数格式取决于模板定义
package tencent
