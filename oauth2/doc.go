/*
Package oauth2 提供标准 OAuth2 授权码流程和第三方登录 Provider 抽象。

oauth2 核心层提供 Endpoint、AuthCodeClient、Token、授权地址生成、授权码换 token、
刷新 token 和统一 HTTP 错误处理。登录层通过 Provider 接口扩展 UserInfo 能力，
具体平台实现放在 adapter 子包中，业务代码可以依赖 Provider 接口，在微信、QQ、
支付宝、抖音等登录服务之间切换，也可以直接使用标准 AuthCodeClient 对接普通 OAuth2 服务。

# 功能特性

  - 标准 OAuth2 授权码客户端
  - 统一的第三方登录 Provider 接口
  - 支持授权地址生成、授权码换 token、刷新 token、获取用户信息
  - Token 和 User 保留平台原始响应，方便排查字段差异
  - 统一处理 form 请求、非 2xx 响应和 ProviderError
  - provider 响应体限制为 1 MiB
  - 适配器独立模块，避免主包引入不必要依赖
  - 面向网站应用 OAuth2 授权码登录流程

# 快速开始

标准 OAuth2 授权码流程：

	package main

	import (
		"context"
		"log"

		"github.com/f2xme/gox/oauth2"
	)

	func main() {
		client := oauth2.NewClient(
			oauth2.WithClientID("client_id"),
			oauth2.WithClientSecret("client_secret"),
			oauth2.WithRedirectURL("https://example.com/callback"),
			oauth2.WithEndpoint(oauth2.Endpoint{
				AuthURL:  "https://provider.example/oauth/authorize",
				TokenURL: "https://provider.example/oauth/token",
			}),
		)

		loginURL := client.AuthCodeURL("csrf-state", oauth2.WithScopes("profile", "email"))
		_ = loginURL

		token, err := client.Exchange(context.Background(), "code-from-callback")
		if err != nil {
			log.Fatal(err)
		}
		_ = token
	}

第三方登录 Provider：

	package main

	import (
		"context"
		"log"

		"github.com/f2xme/gox/oauth2"
		"github.com/f2xme/gox/oauth2/adapter/wechat"
	)

	func main() {
		provider := wechat.New(
			wechat.WithClientID("appid"),
			wechat.WithClientSecret("secret"),
			wechat.WithRedirectURL("https://example.com/callback"),
		)

		loginURL := provider.AuthCodeURL("csrf-state", oauth2.WithScopes("snsapi_login"))
		_ = loginURL

		token, err := provider.Exchange(context.Background(), "code-from-callback")
		if err != nil {
			log.Fatal(err)
		}

		user, err := provider.UserInfo(context.Background(), token)
		if err != nil {
			log.Fatal(err)
		}
		_ = user
	}

# 适配器

首批适配器：

  - wechat：微信开放平台网站应用登录
  - qq：QQ 互联网站应用登录
  - alipay：支付宝开放平台网页授权登录
  - douyin：抖音开放平台网站应用登录

# 注意事项

  - oauth2 首版重点覆盖 Web/服务端 OAuth2 授权码登录，不包含移动端 SDK、小程序登录或 JS 嵌入扫码。
  - state 参数应由业务侧生成、存储并在回调时校验，用于防止 CSRF。
  - client secret 和用户 token 应保存在服务端，不应暴露给浏览器或移动端。
*/
package oauth2
