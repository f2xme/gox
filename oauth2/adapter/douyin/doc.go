/*
Package douyin 提供抖音开放平台网站应用登录适配器。

# 功能特性

  - 生成抖音登录授权地址
  - 使用授权码换取 access_token
  - 使用 refresh_token 续期 access_token
  - 获取抖音用户基础信息
  - 映射 open_id、union_id、昵称、头像、地区等通用字段

# 快速开始

基本使用：

	package main

	import (
		"context"

		"github.com/f2xme/gox/oauth2"
		"github.com/f2xme/gox/oauth2/adapter/douyin"
	)

	func main() {
		provider := douyin.New(
			douyin.WithClientID("client_key"),
			douyin.WithClientSecret("client_secret"),
			douyin.WithRedirectURL("https://example.com/auth/douyin/callback"),
		)

		_ = provider.AuthCodeURL("state", oauth2.WithScopes("user_info"))

		token, _ := provider.Exchange(context.Background(), "code")
		user, _ := provider.UserInfo(context.Background(), token)
		_ = user
	}

# 注意事项

  - 本包面向抖音开放平台网站应用授权码登录流程。
  - state 参数需要由业务侧生成并校验。
*/
package douyin
