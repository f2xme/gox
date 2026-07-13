/*
Package alipay 提供支付宝开放平台网页授权登录适配器。

# 功能特性

  - 生成支付宝网页授权登录地址
  - 使用授权码换取 access_token
  - 使用 refresh_token 续期 access_token
  - 使用 RSA/RSA2 对支付宝网关请求签名
  - 使用支付宝公钥验证网关响应签名
  - 获取支付宝用户基础信息
  - 映射 user_id、open_id、昵称、头像、地区等通用字段

# 快速开始

基本使用：

	package main

	import (
		"context"

		"github.com/f2xme/gox/oauth2"
		"github.com/f2xme/gox/oauth2/adapter/alipay"
	)

	func main() {
		provider := alipay.New(
			alipay.WithClientID("app_id"),
			alipay.WithPrivateKey("-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"),
			alipay.WithAlipayPublicKey("-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----"),
			alipay.WithRedirectURL("https://example.com/auth/alipay/callback"),
		)

		_ = provider.AuthCodeURL("state", oauth2.WithScopes("auth_user"))

		token, _ := provider.Exchange(context.Background(), "auth_code")
		user, _ := provider.UserInfo(context.Background(), token)
		_ = user
	}

# 注意事项

  - 本包面向支付宝开放平台网页授权登录流程，默认 scope 为 auth_user。
  - 应用私钥只应保存在服务端，不应暴露给浏览器或移动端。
  - 必须配置支付宝公钥，以验证网关 token 和用户信息响应。
  - state 参数需要由业务侧生成并校验。
*/
package alipay
