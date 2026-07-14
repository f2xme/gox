// Package memory 提供基于内存的 IP 地区查询适配器。
//
// memory 包实现 geo.Locator 接口，不会访问外部数据源或网络。
// 通过预置 IP 到 Location 的映射，适合单元测试和本地开发。
//
// # 功能特性
//
//   - 实现 geo.Locator：可替换真实数据源适配器
//   - 预置映射：支持构造时批量注册 IP 地区信息
//   - 动态维护：运行时支持 Set、Delete、Reset
//   - 错误注入：支持配置 Lookup 固定错误，用于测试失败分支
//   - 并发安全：可在并发测试中安全调用
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"context"
//
//		"github.com/f2xme/gox/geo"
//		"github.com/f2xme/gox/geo/adapter/memory"
//	)
//
//	func main() {
//		locator, err := memory.New(
//			memory.WithLocation("1.2.3.4", &geo.Location{
//				Country:  "中国",
//				Province: "广东省",
//				City:     "深圳市",
//			}),
//		)
//		if err != nil {
//			return
//		}
//
//		loc, err := locator.Lookup(context.Background(), "1.2.3.4")
//		_ = loc
//		_ = err
//	}
//
// # 注意事项
//
//   - 仅适用于测试、本地开发和单进程场景
//   - 进程退出后映射数据会丢失
//   - 未注册的 IP 查询会返回 NotFound 错误
package memory
