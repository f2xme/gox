// Package email 提供通过 SMTP 发送邮件的客户端。
//
// 本包封装了 gomail，提供简单的邮件发送接口，支持 HTML 内容、附件和多个收件人。
//
// # 功能特性
//
//   - 支持纯文本和 HTML 邮件发送
//   - 支持 SSL/TLS 加密连接
//   - 灵活的 Options 模式配置
//   - 支持从配置文件创建客户端
//   - 支持自定义发件人名称
//   - 简洁的 API 设计
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
//		"github.com/f2xme/gox/email"
//	)
//
//	func main() {
//		// 推荐：使用 Options 模式
//		client, err := email.NewWithOptions(
//			email.WithHost("smtp.gmail.com"),
//			email.WithPort(587),
//			email.WithUsername("user@gmail.com"),
//			email.WithPassword("password"),
//			email.WithName("My App"),
//			email.WithSSL(false),
//		)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// 发送纯文本邮件
//		err = client.SendText("recipient@example.com", "Hello", "Hello, World!")
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// 发送 HTML 邮件
//		html := "<html><body><h1>Hello</h1><p>Welcome!</p></body></html>"
//		err = client.SendHTML("recipient@example.com", "Welcome", html)
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//
// # 从配置文件创建
//
// 支持从配置文件读取邮件服务器配置：
//
//	cfg := config.New("config.yaml")
//	client, err := email.NewWithConfig(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// 配置文件示例（YAML）：
//
//	email:
//	  host: smtp.gmail.com
//	  port: 587
//	  username: user@gmail.com
//	  password: your-password
//	  name: My App
//	  ssl: false
//
// # 向后兼容
//
// 兼容旧版本 API（仍然支持）：
//
//	client, err := email.New("smtp.gmail.com", 587, "user@gmail.com", "password",
//		email.WithName("My App"),
//		email.WithSSL(false),
//	)
package email
