// Package memory 提供基于内存的短信服务适配器。
//
// memory 包实现 sms.SMS 接口，不会访问真实短信服务。它会把发送成功的短信
// 记录在内存中，适合在单元测试和本地开发中断言业务代码是否发起了正确短信。
//
// # 功能特性
//
//   - 实现 sms.SMS：可替换真实短信服务商适配器
//   - 发送记录：保存手机号、模板代码、模板参数和发送时间
//   - 测试辅助：支持读取全部记录、读取最后一条、统计数量和重置记录
//   - 错误注入：支持配置发送错误，用于测试失败分支
//   - 并发安全：可在并发测试中安全调用
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
//		"github.com/f2xme/gox/sms/adapter/memory"
//	)
//
//	func main() {
//		client, err := memory.New()
//		if err != nil {
//			return
//		}
//
//		_ = client.Send(context.Background(), sms.Message{
//			Phone:         "13800138000",
//			TemplateCode:  "login_code",
//			TemplateParam: map[string]string{"code": "1234"},
//		})
//
//		last, ok := client.LastMessage()
//		_ = last
//		_ = ok
//	}
//
// # 注意事项
//
//   - 仅适用于测试、本地开发和单进程场景
//   - 进程退出后发送记录会丢失
//   - TemplateParam 会复制常见可变类型，未知类型会按原值保存
package memory
