// Package session 提供将 gox/session 接入 auth 中间件的适配器。
//
// # 功能特性
//
//   - 将 session.Manager 适配为 auth.Validator
//   - 从 Cookie 中提取 session ID
//   - 支持 session 验证器的 UID 键和滑动过期配置
//   - 保持 auth 核心包不依赖 session 包
//
// # 快速开始
//
// 基本用法：
//
//	app.Use(auth.New(
//		auth.WithValidator(authsession.NewValidator(manager)),
//		auth.WithTokenExtractor(authsession.NewExtractor(authsession.DefaultCookieName)),
//	))
package session
