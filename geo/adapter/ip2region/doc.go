// Package ip2region 提供基于 ip2region xdb 的离线 IP 地区查询适配器。
//
// ip2region 包实现 geo.Locator 接口，使用 lionsoul2014/ip2region 的
// 并发安全查询服务，支持 IPv4 / IPv6（按配置的 xdb 文件决定）。
//
// # 功能特性
//
//   - 离线查询：无需访问外部网络
//   - 双栈支持：可同时配置 v4 / v6 xdb
//   - 缓存策略：支持 NoCache、VIndexCache、BufferCache
//   - 统一结果：将 region 字符串解析为 geo.Location
//   - 资源释放：提供 Close 关闭底层查询服务
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
//		"github.com/f2xme/gox/geo/adapter/ip2region"
//	)
//
//	func main() {
//		locator, err := ip2region.New(
//			ip2region.WithV4DBPath("/data/ip2region_v4.xdb"),
//			ip2region.WithCachePolicy(ip2region.CachePolicyVIndex),
//		)
//		if err != nil {
//			log.Fatal(err)
//		}
//		defer locator.Close()
//
//		loc, err := locator.Lookup(context.Background(), "1.2.3.4")
//		if err != nil {
//			log.Fatal(err)
//		}
//		log.Println(loc.Country, loc.Province, loc.City, loc.ISP)
//	}
//
// # 数据文件
//
// xdb 数据文件需从 ip2region 官方仓库获取，例如：
//
//	https://github.com/lionsoul2014/ip2region
//
// # 注意事项
//
//   - 至少配置 V4DBPath 或 V6DBPath 之一
//   - BufferCache 会把整个 xdb 载入内存，查询最快但占用更大
//   - 进程退出前应调用 Close 释放资源
//   - 查询结果含「内网IP」等占位文案时返回 NotFound，与 amap 等适配器策略一致
//   - 底层 Search 不支持 context 中途取消，仅入口检查 ctx.Err()
//   - 集成测试需设置 GEO_IP2REGION_V4_DB 环境变量指向真实 xdb
package ip2region
