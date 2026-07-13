/*
Package qq 提供 QQ 互联网站应用登录适配器。

# 功能特性

  - 生成 QQ 登录授权地址
  - 使用授权码换取 access_token
  - 自动调用 openid 接口补齐用户标识
  - 使用 refresh_token 续期 access_token
  - 获取 QQ 用户基础信息

# 快速开始

基本使用：

	package main

	import (
		"context"

		"github.com/f2xme/gox/oauth2"
		"github.com/f2xme/gox/oauth2/adapter/qq"
	)

	func main() {
		provider := qq.New(
			qq.WithClientID("appid"),
			qq.WithClientSecret("appkey"),
			qq.WithRedirectURL("https://example.com/auth/qq/callback"),
		)

		_ = provider.AuthCodeURL("state", oauth2.WithScopes("get_user_info"))

		token, _ := provider.Exchange(context.Background(), "code")
		user, _ := provider.UserInfo(context.Background(), token)
		_ = user
	}

# 注意事项

  - QQ 需要先通过 token 调用 openid 接口，再用 openid 获取用户信息。
  - state 参数需要由业务侧生成并校验。
*/
package qq
