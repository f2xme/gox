/*
Package x 提供 X 网站应用 OAuth 2.0 授权码登录适配器。

# 功能特性

  - 生成带 S256 PKCE 的授权地址
  - 默认申请 tweet.read、users.read、offline.access
  - 使用授权码和 code_verifier 换取 access_token
  - 使用 refresh_token 续期，并保留轮换后的新刷新令牌
  - 获取当前用户的编号、显示名和头像

# 快速开始

基本使用：

	package main

	import (
		"context"

		"github.com/f2xme/gox/oauth2/adapter/x"
	)

	func main() {
		provider := x.New(
			x.WithClientID("client_id"),
			x.WithClientSecret("client_secret"),
			x.WithRedirectURL("https://example.com/auth/x/callback"),
		)

		authURL, verifier, _ := provider.Begin("state")
		_ = authURL

		token, _ := provider.ExchangeCode(context.Background(), "code", verifier)
		user, _ := provider.UserInfo(context.Background(), token)
		_ = user
	}

# 注意事项

  - 本包面向 X 的 OAuth 2.0 授权码加 PKCE，不包含 OAuth 1.0a。
  - 授权码大约 30 秒内必须换成令牌。
  - verifier 要和 state 一起存在服务端，换票时传给 ExchangeCode，或放进 WithCodeVerifier。
  - 机密客户端会用 Client Secret 做 HTTP Basic 认证。公开客户端把 Client Secret 留空。
  - offline.access 才会签发 refresh_token。X 会轮换刷新令牌，必须保存响应中的新值。
  - 回调地址必须与开发者后台登记的地址完全一致。
  - ID 和 OpenID 都是 X 用户编号。用户名保留在 Raw 中。
  - state 参数需要由业务侧生成并校验。
*/
package x
