// Package httpadapter 提供基于 HTTP 远端配置文件的配置适配器实现。
//
// 该包通过 HTTP GET 拉取 JSON 或 YAML 配置内容，并封装为 gox/config
// 接口的实现。它适合轻量级在线配置场景，例如内部配置服务、对象存储、
// 静态文件服务或网关后面的配置文件。
//
// # 功能特性
//
//   - 支持 JSON 和 YAML 配置内容
//   - 支持默认值
//   - 支持自定义请求头
//   - 支持请求超时
//   - 默认限制远端响应体为 1 MiB
//   - 支持通过轮询实现热更新
//   - 类型安全的配置读取
//
// # 快速开始
//
//	cfg, err := httpadapter.New(
//		"http://config.internal/app.yaml",
//		httpadapter.WithFormat(httpadapter.YAML),
//		httpadapter.WithDefaults(map[string]any{
//			"server.port": 8080,
//		}),
//	)
//	if err != nil {
//		// 处理配置加载失败
//	}
//
//	port := cfg.GetInt("server.port")
//	_ = port
//
// # 变更监听
//
//	cfg, _ := httpadapter.New(
//		"http://config.internal/app.yaml",
//		httpadapter.WithWatch(10*time.Second),
//	)
//	cfg.(config.Watcher).Watch(func() {
//		// 远端配置内容发生变化
//	})
//
// # 注意事项
//
//   - Watch 通过轮询 URL 实现，不依赖服务端推送
//   - 远端配置只在 HTTP 2xx 响应时生效
//   - 配置更新失败时会保留上一份可用配置
//   - 该适配器不提供配置发布、权限、灰度或审计能力
package httpadapter
