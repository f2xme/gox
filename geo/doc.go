// Package geo 提供统一的 IP 地区查询抽象。
//
// geo 包只定义 IP 地理位置查询的通用接口、结果结构和错误模型，
// 不绑定具体数据源或服务商。具体实现放在 adapter 子包中。
//
// # 功能特性
//
//   - 统一接口：通过 Locator 屏蔽离线库与在线 API 差异
//   - 结构化结果：Location 描述国家、省、市、运营商等信息
//   - 上下文控制：Lookup 接收 context.Context，支持取消和超时
//   - 统一错误：提供 InvalidIP、NotFound 等错误码与判断函数
//   - Adapter 扩展：新增实现时只需实现 Locator 接口
//
// # 快速开始
//
// 使用内存适配器（适合测试）：
//
//	package main
//
//	import (
//		"context"
//		"fmt"
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
//				ISP:      "电信",
//			}),
//		)
//		if err != nil {
//			return
//		}
//
//		loc, err := locator.Lookup(context.Background(), "1.2.3.4")
//		if err != nil {
//			return
//		}
//		fmt.Println(loc.Country, loc.Province, loc.City)
//	}
//
// 使用 ip2region 离线库：
//
//	import "github.com/f2xme/gox/geo/adapter/ip2region"
//
//	locator, err := ip2region.New(
//		ip2region.WithV4DBPath("/path/to/ip2region_v4.xdb"),
//	)
//
// 使用 HTTP 在线 API：
//
//	import geohttp "github.com/f2xme/gox/geo/adapter/http"
//
//	locator, err := geohttp.New(
//		geohttp.WithEndpoint("http://ip-api.com/json/"),
//	)
//
// 使用高德 IP 定位：
//
//	import "github.com/f2xme/gox/geo/adapter/amap"
//
//	locator, err := amap.New(amap.WithKey("your-amap-key"))
//
// 使用百度 IP 查询：
//
//	import "github.com/f2xme/gox/geo/adapter/baidu"
//
//	locator, err := baidu.New()
//
// # 核心接口
//
// Locator 定义 IP 地区查询能力：
//
//	type Locator interface {
//		Lookup(ctx context.Context, ip string) (*Location, error)
//	}
//
// # 错误处理
//
// Adapter 会尽量把失败转换为 *geo.Error：
//
//	loc, err := locator.Lookup(ctx, "bad-ip")
//	if geo.IsInvalidIP(err) {
//		// IP 格式无效
//	} else if geo.IsNotFound(err) {
//		// 未查到地区信息
//	}
//
// # Adapter 约定
//
// 新增 adapter 时建议放在 geo/adapter/<provider> 下，使用 Options 模式创建实例，
// 实现 Locator 接口。若实现持有可关闭资源（如文件句柄、连接池），可额外提供 Close 方法。
//
// # 注意事项
//
//   - Locator 实现应当是并发安全的
//   - 未提供的 Location 字段保持零值，业务侧勿假设所有字段都有数据
//   - 私有 IP / 内网占位结果：amap、ip2region、baidu 等返回 NotFound；memory 仅查映射表
//   - 依赖本地数据文件或外部服务的 adapter 应通过集成测试环境变量显式启用
//   - HTTP 类 adapter 的 endpoint 必须来自可信配置，勿拼接用户输入
package geo
