// Package aliyun 提供基于阿里云的短信服务实现。
//
// # 功能特性
//
//   - 基于阿里云短信服务 SDK 封装
//   - 支持模板短信发送
//   - 统一的错误处理
//   - 支持阿里云默认凭据链和显式 AccessKey
//   - 支持配置文件初始化
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"context"
//		"log"
//
//		"github.com/f2xme/gox/sms"
//		"github.com/f2xme/gox/sms/adapter/aliyun"
//	)
//
//	func main() {
//		// 创建阿里云短信客户端。默认使用阿里云凭据链读取环境变量、
//		// 配置文件、RAM 角色等凭据。
//		client, err := aliyun.New(
//			aliyun.WithSignName("your-sign-name"),
//		)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// 发送短信
//		err = client.Send(context.Background(), sms.Message{
//			Phone:         "13800138000",
//			TemplateCode:  "SMS_123456789",
//			TemplateParam: map[string]string{"code": "1234"},
//		})
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
//	    # accessKeyID 和 accessKeySecret 可选；未配置时使用阿里云默认凭据链
//	    accessKeyID: "your-access-key-id"
//	    accessKeySecret: "your-access-key-secret"
//	    endpoint: "dysmsapi.aliyuncs.com"
//	    signName: "your-sign-name"
//
// # 注意事项
//
//   - 确保阿里云账号已开通短信服务
//   - 推荐使用环境变量、配置文件、RAM 角色等方式配置阿里云默认凭据链
//   - 短信签名和模板需要在阿里云控制台预先配置
//   - templateParam 可传入 map、struct、JSON 字符串或 []byte
package aliyun
