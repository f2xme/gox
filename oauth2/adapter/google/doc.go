/*
Package google 提供 Google 网站应用 OpenID Connect 登录适配器。

# 功能特性

  - 生成 Google 授权码登录地址
  - 默认申请 openid、email、profile，并设置 access_type=offline
  - 使用授权码换取 access_token
  - 使用 refresh_token 续期 access_token
  - 通过 OpenID Connect userinfo 获取用户基础信息
  - 将 sub 映射为用户唯一标识，头像和姓名映射到通用字段

# 快速开始

基本使用：

	package main

	import (
		"context"

		"github.com/f2xme/gox/oauth2"
		"github.com/f2xme/gox/oauth2/adapter/google"
	)

	func main() {
		provider := google.New(
			google.WithClientID("client_id"),
			google.WithClientSecret("client_secret"),
			google.WithRedirectURL("https://example.com/auth/google/callback"),
		)

		_ = provider.AuthCodeURL("state", oauth2.WithScopes("openid", "email", "profile"))

		token, _ := provider.Exchange(context.Background(), "code")
		user, _ := provider.UserInfo(context.Background(), token)
		_ = user
	}

# 注意事项

  - 本包面向 Google Cloud 中的网站应用客户端，使用客户端密钥完成授权码流程。
  - 回调地址必须与控制台登记的重定向 URI 完全一致。
  - 默认 access_type=offline。Google 通常只在用户首次同意时返回 refresh_token；
    需要每次授权都拿到刷新令牌时，追加 oauth2.WithAuthParam("prompt", "consent")。
  - 刷新响应没有新的 refresh_token 时，沿用本次提交的刷新令牌。
  - Google 没有 openid 与 unionid 之分。ID 和 OpenID 都使用 sub，UnionID 为空。
  - 邮箱、email_verified、locale，以及令牌响应中的 id_token 保留在 Raw 中。
  - state 参数需要由业务侧生成并校验。
  - 客户端密钥和用户令牌只应保存在服务端。
*/
package google
