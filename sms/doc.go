/*
Package sms 提供统一的短信服务抽象层。

# 概述

sms 包定义了短信服务的标准接口，支持多种短信服务提供商。
通过这些接口，你可以轻松地在不同的短信服务提供商之间切换，而无需修改业务代码。

# 核心接口

	type SMS interface {
		Send(phone, templateCode, templateParam string) error
	}

使用示例：

	// 发送验证码短信
	err := sms.Send(
		"13800138000",
		"SMS_123456789",
		`{"code":"1234"}`,
	)

# 可用适配器

## 阿里云短信

	import "github.com/f2xme/gox/sms/aliyun"

## 腾讯云短信

	import "github.com/f2xme/gox/sms/tencent"

## 火山引擎短信

	import "github.com/f2xme/gox/sms/volcengine"
*/
package sms
