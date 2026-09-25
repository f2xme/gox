/*
Package facebook 提供 Facebook 网站应用授权码登录适配器。

# 功能特性

  - 生成 Facebook Login 授权地址
  - 默认申请 public_profile、email
  - 使用授权码换取短期用户访问令牌
  - 把短期用户访问令牌换成约 60 天的长期令牌
  - 获取用户编号、姓名和头像
  - 邮箱保留在原始响应中

# 快速开始

基本使用：

	package main

	import (
		"context"

		"github.com/f2xme/gox/oauth2"
		"github.com/f2xme/gox/oauth2/adapter/facebook"
	)

	func main() {
		provider := facebook.New(
			facebook.WithClientID("app_id"),
			facebook.WithClientSecret("app_secret"),
			facebook.WithRedirectURL("https://example.com/auth/facebook/callback"),
		)

		_ = provider.AuthCodeURL("state")

		token, _ := provider.Exchange(context.Background(), "code")
		longLived, _ := provider.RefreshToken(context.Background(), token.AccessToken)
		user, _ := provider.UserInfo(context.Background(), longLived)
		_ = user
	}

# 注意事项

  - 本包面向 Facebook Login 的网站应用手动授权码流程，默认使用 Graph API v26.0。
  - 回调地址必须与应用后台登记的重定向 URI 完全一致。
  - 换票请求包含应用密钥，只能在服务端发起。
  - Facebook 不签发 refresh_token。RefreshToken 接收的是仍有效的短期 access_token。
  - 已过期的用户令牌不能续期，需要重新授权。
  - ID 和 OpenID 都是应用范围内的用户编号，UnionID 为空。
  - 用户可以拒绝 email 权限，此时 Raw 中没有邮箱。
  - state 参数需要由业务侧生成并校验。
*/
package facebook
