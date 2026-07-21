// Package amap 提供基于高德地图 IP 定位 API 的地区查询适配器。
//
// amap 包实现 geo.Locator 接口，调用高德 Web 服务 IP 定位接口
// （https://restapi.amap.com/v3/ip）查询 IPv4 对应的省市信息。
//
// # 功能特性
//
//   - 实现 geo.Locator：可替换其他 IP 地区数据源
//   - Options 配置：支持 API Key、数字签名私钥、超时、自定义 HTTP 客户端与端点
//   - 结构化结果：解析 province、city、adcode 到 geo.Location
//   - 统一错误：将高德 status/info 映射为 geo 错误码
//   - 来源定位：LookupCurrent 可省略 IP，由高德定位当前请求来源
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
//		"github.com/f2xme/gox/geo/adapter/amap"
//	)
//
//	func main() {
//		locator, err := amap.New(amap.WithKey("your-amap-key"))
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		loc, err := locator.Lookup(context.Background(), "114.247.50.2")
//		if err != nil {
//			log.Fatal(err)
//		}
//		log.Println(loc.Province, loc.City)
//	}
//
// # 注意事项
//
//   - 必须配置有效的高德 Web 服务 Key
//   - 数字签名认证用户需通过 WithPrivateKey 配置私钥，sig 会按请求动态生成
//   - 高德 IP 定位主要面向国内 IPv4，海外或内网 IP 可能无结果（返回 NotFound）
//   - Key 无效、配额耗尽等上游故障映射为 Unavailable
//   - endpoint 必须来自可信配置，勿与用户输入拼接（防 SSRF / Key 泄露）
//   - 请遵守高德开放平台配额与使用条款
//   - 不要把 Key 提交到公开仓库
package amap
