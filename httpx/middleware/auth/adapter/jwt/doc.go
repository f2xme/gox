// Package jwt 提供将 gox/jwt 接入 auth 中间件的适配器。
//
// # 功能特性
//
//   - 将 gox/jwt.JWT 适配为 auth.Validator
//   - 支持 HS256、HS384、HS512 快捷构造函数
//   - 保持 jwt 核心包不依赖 HTTP 中间件
//
// # 快速开始
//
// 基本用法：
//
//	app.Use(auth.New(
//		auth.WithValidator(jwtadapter.NewHS256Validator([]byte("secret"))),
//	))
package jwt
