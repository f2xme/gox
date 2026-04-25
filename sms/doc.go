// Package sms 提供统一的短信服务抽象层。
//
// sms 包定义短信发送的标准接口和消息结构，让业务代码可以通过同一套 API
// 使用不同短信服务商。具体服务商实现位于独立 adapter module 中。
//
// # 功能特性
//
//   - 统一接口：通过 SMS 接口屏蔽不同服务商 SDK 差异
//   - 结构化消息：使用 Message 描述手机号、模板和模板参数
//   - 上下文控制：Send 方法接收 context.Context，支持取消和超时
//   - 独立适配器：阿里云、腾讯云、火山引擎适配器可按需引入
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
//		client, err := aliyun.New(
//			aliyun.WithAccessKeyID("your-key-id"),
//			aliyun.WithAccessKeySecret("your-key-secret"),
//			aliyun.WithSignName("your-sign-name"),
//		)
//		if err != nil {
//			log.Fatal(err)
//		}
//
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
// # 可用适配器
//
//   - 阿里云短信：github.com/f2xme/gox/sms/adapter/aliyun
//   - 腾讯云短信：github.com/f2xme/gox/sms/adapter/tencent
//   - 火山引擎短信：github.com/f2xme/gox/sms/adapter/volcengine
package sms
