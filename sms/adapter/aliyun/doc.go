// Package aliyun 提供基于阿里云的短信服务实现。
//
// # 功能特性
//
//   - 基于阿里云短信服务 SDK 封装
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
//		"github.com/f2xme/gox/sms/adapter/aliyun"
//	)
//
//	func main() {
//		// 创建阿里云短信客户端
//		client, err := aliyun.New(
//			aliyun.WithAccessKeyID("your-access-key-id"),
//			aliyun.WithAccessKeySecret("your-access-key-secret"),
//			aliyun.WithSignName("your-sign-name"),
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
//		"github.com/f2xme/gox/sms/adapter/aliyun"
//	)
//
//	func main() {
//		cfg := config.New()
//		client, err := aliyun.NewWithConfig(cfg)
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//
// 配置文件示例（YAML）：
//
//	sms:
//	  aliyun:
//	    accessKeyID: "your-access-key-id"
//	    accessKeySecret: "your-access-key-secret"
//	    endpoint: "dysmsapi.aliyuncs.com"
//	    signName: "your-sign-name"
//
// # 注意事项
//
//   - 确保阿里云账号已开通短信服务
//   - 短信签名和模板需要在阿里云控制台预先配置
//   - templateParam 参数必须是有效的 JSON 字符串
package aliyun
