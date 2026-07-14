// Package http 提供基于 HTTP API 的 IP 地区查询适配器。
//
// http 包实现 geo.Locator 接口，通过调用可配置的 HTTP 端点查询 IP 地区。
// 默认按常见 JSON 字段名解析响应，也支持自定义解析函数。
//
// # 功能特性
//
//   - 实现 geo.Locator：可替换离线库或其他在线服务
//   - 可配置端点：支持 URL 模板或前缀拼接 IP
//   - 默认 JSON 解析：识别 country、region/province、city、isp 等常见字段
//   - 自定义解析：通过 WithResponseParser 适配任意响应格式
//   - 超时与客户端：支持自定义 http.Client 与请求超时
//
// # 快速开始
//
// 基本使用（以 ip-api.com 风格端点为例）：
//
//	package main
//
//	import (
//		"context"
//
//		geohttp "github.com/f2xme/gox/geo/adapter/http"
//	)
//
//	func main() {
//		locator, err := geohttp.New(
//			geohttp.WithEndpoint("http://ip-api.com/json/"),
//		)
//		if err != nil {
//			return
//		}
//
//		loc, err := locator.Lookup(context.Background(), "8.8.8.8")
//		_ = loc
//		_ = err
//	}
//
// # 响应格式
//
// 默认解析器会尝试读取以下 JSON 字段（不区分大小写）：
//
//   - country / country_name
//   - countryCode / country_code
//   - regionName / region / province（映射到 Province）
//   - city
//   - district / county（映射到 District）
//   - isp / org / as
//   - lat / latitude
//   - lon / lng / longitude
//
// # 注意事项
//
//   - 请遵守目标 API 的使用条款与速率限制
//   - 生产环境建议配置合理超时与可重试策略
//   - 端点必须使用 HTTP 或 HTTPS，且必须来自可信配置（勿拼接用户输入，防 SSRF）
//   - 与标准库 net/http 同用时建议使用 import 别名，例如 geohttp
//   - URL 模板中仅替换第一个 "%s" 占位符；endpoint 中已有的 %XX 编码不会被破坏
package http
