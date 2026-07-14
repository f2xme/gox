// Package baidu 提供基于百度 IP 查询接口的地区查询适配器。
//
// baidu 包实现 geo.Locator 接口，调用百度公开 IP 查询接口解析归属地。
// 该接口响应通常为 GBK 编码，适配器会自动转成 UTF-8 再解析。
//
// # 功能特性
//
//   - 实现 geo.Locator：可替换其他 IP 地区数据源
//   - GBK 解码：自动处理百度返回的中文编码
//   - 结构化结果：尽量从 location 文本拆分省、市、运营商
//   - Options 配置：支持超时、自定义客户端与端点
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"context"
//		"log"
//
//		"github.com/f2xme/gox/geo/adapter/baidu"
//	)
//
//	func main() {
//		locator, err := baidu.New()
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		loc, err := locator.Lookup(context.Background(), "114.247.50.2")
//		if err != nil {
//			log.Fatal(err)
//		}
//		log.Println(loc.Province, loc.City, loc.ISP)
//	}
//
// # 注意事项
//
//   - 该接口为非官方稳定契约，字段与可用性可能变化
//   - 生产环境建议优先使用正式商业 API 或离线库
//   - 请控制请求频率，避免被限流
//   - endpoint 必须来自可信配置；自定义 endpoint 时注意 SSRF 风险
//   - 国内结果填充 Country=中国 / CountryCode=CN（需命中省/自治区/直辖市等明确结构）
//   - 海外文本将首段作为 Country；单独「xx市」不会默认判为中国
//   - 内网/局域网等占位文案返回 NotFound
//   - 海外 ISP/省拆分不可靠，完整原文见 Extra["location"]
//   - Timeout<=0 回落默认 5 秒，与其他 HTTP 类适配器一致
package baidu
