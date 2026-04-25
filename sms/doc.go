/*
Package sms 提供统一的短信服务抽象层。

# 概述

sms 包定义了短信服务的标准接口，支持多种短信服务提供商。
通过这些接口，你可以轻松地在不同的短信服务提供商之间切换，而无需修改业务代码。

# 核心接口

	type SMS interface {
		Send(phone, templateCode, templateParam string) error
	}

# 可用适配器（独立 module）

## 阿里云短信

	import "github.com/f2xme/gox/sms/adapter/aliyun"

## 腾讯云短信

	import "github.com/f2xme/gox/sms/adapter/tencent"

## 火山引擎短信

	import "github.com/f2xme/gox/sms/adapter/volcengine"

# 使用示例

	import (
		"github.com/f2xme/gox/sms"
		"github.com/f2xme/gox/sms/adapter/aliyun"
	)

	client, err := aliyun.New(
		aliyun.WithAccessKeyID("your-key-id"),
		aliyun.WithAccessKeySecret("your-key-secret"),
		aliyun.WithSignName("your-sign-name"),
	)
	if err != nil {
		// 处理错误
	}

	err = client.Send("13800138000", "SMS_123456789", `{"code":"1234"}`)
*/
package sms
