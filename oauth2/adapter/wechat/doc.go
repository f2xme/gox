/*
Package wechat 提供微信开放平台网站应用登录适配器。

# 功能特性

  - 生成微信扫码登录授权地址
  - 使用授权码换取 access_token
  - 使用 refresh_token 续期 access_token
  - 获取微信用户基础信息
  - 映射 openid、unionid、昵称、头像、地区等通用字段

# 快速开始

基本使用：

	package main

	import (
		"context"

		"github.com/f2xme/gox/oauth2"
		"github.com/f2xme/gox/oauth2/adapter/wechat"
	)

	func main() {
		provider := wechat.New(
			wechat.WithClientID("appid"),
			wechat.WithClientSecret("secret"),
			wechat.WithRedirectURL("https://example.com/auth/wechat/callback"),
		)

		_ = provider.AuthCodeURL("state", oauth2.WithScopes("snsapi_login"))

		token, _ := provider.Exchange(context.Background(), "code")
		user, _ := provider.UserInfo(context.Background(), token)
		_ = user
	}

# 注意事项

  - 本包面向微信开放平台网站应用登录，默认 scope 为 snsapi_login。
  - state 参数需要由业务侧生成并校验。
*/
package wechat
