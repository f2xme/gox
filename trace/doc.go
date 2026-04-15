// Package trace 提供轻量级链路追踪功能。
//
// 用于在应用各层之间传递追踪信息，记录操作耗时和属性。
// 提供简单的 API 用于在 context 中存储和提取追踪信息。
//
// # 功能特性
//
//   - 在 context 中传递追踪信息（TraceID、SpanID、DeviceID、RequestID）
//   - 提供 HTTP 中间件自动注入和提取追踪信息
//   - 支持 Span 记录操作耗时和属性
//   - 与 gin.Context 等框架兼容
//   - 轻量级，零外部依赖
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"context"
//		"net/http"
//
//		"github.com/f2xme/gox/trace"
//	)
//
//	func main() {
//		// 1. 使用 HTTP 中间件自动注入追踪信息
//		mux := http.NewServeMux()
//		mux.HandleFunc("/api/user", handleUser)
//
//		handler := trace.HTTPMiddleware(mux)
//		http.ListenAndServe(":8080", handler)
//	}
//
//	func handleUser(w http.ResponseWriter, r *http.Request) {
//		ctx := r.Context()
//
//		// 2. 从 context 提取追踪信息
//		info := trace.FromContext(ctx)
//		println("TraceID:", info.TraceID)
//
//		// 3. 使用 Span 记录操作耗时
//		getUserFromDB(ctx, 123)
//	}
//
//	func getUserFromDB(ctx context.Context, userID int64) (err error) {
//		defer trace.DAO(ctx, "GetUser", "user_id", userID)(&err)
//		// 查询数据库...
//		return nil
//	}
//
// # 追踪信息字段
//
//   - TraceID: 链路追踪 ID（全局唯一，贯穿整个请求链路）
//   - SpanID: 跨度 ID（当前操作的唯一标识）
//   - DeviceID: 设备 ID（客户端设备标识）
//   - RequestID: 请求 ID（单次请求的唯一标识）
//
// # 使用 Span 记录操作
//
// Span 用于记录各层操作的耗时和属性：
//
//	// Service 层
//	func (s *Service) GetUser(ctx context.Context, id int64) (user *User, err error) {
//		defer trace.Service(ctx, "GetUser", "id", id)(&err)
//		// 业务逻辑...
//		return s.dao.GetUser(ctx, id)
//	}
//
//	// DAO 层
//	func (d *Dao) GetUser(ctx context.Context, id int64) (user *User, err error) {
//		defer trace.DAO(ctx, "GetUser", "id", id)(&err)
//		// 数据库查询...
//		return user, nil
//	}
//
//	// Cache 层
//	func (c *Cache) Get(ctx context.Context, key string) (val string, err error) {
//		defer trace.Cache(ctx, "Get", "key", key)(&err)
//		// 缓存操作...
//		return val, nil
//	}
//
// # 设置 Span 回调
//
// 通过回调函数记录 Span 结果到日志或监控系统：
//
//	trace.SetCallback(func(r *trace.SpanResult) {
//		log.Printf("[%s] %s duration=%dms success=%v",
//			r.Kind(), r.Name(), r.DurationMs(), r.Success())
//	})
//
// # 跨服务传递追踪信息
//
// 使用 InjectToHeaders 将追踪信息注入到 HTTP 请求头：
//
//	func callDownstream(ctx context.Context) error {
//		req, _ := http.NewRequest("GET", "http://api.example.com/user", nil)
//		trace.InjectToHeaders(ctx, req.Header)
//		resp, err := http.DefaultClient.Do(req)
//		return err
//	}
//
// 或使用 HTTPClient 自动注入：
//
//	func callDownstream(ctx context.Context) error {
//		client := trace.HTTPClient(ctx)
//		resp, err := client.Get("http://api.example.com/user")
//		return err
//	}
//
// # 与日志库集成
//
// 在日志中包含追踪信息：
//
//	import "github.com/f2xme/gox/logx"
//
//	func handleRequest(ctx context.Context) {
//		info := trace.FromContext(ctx)
//		logx.Info("processing request",
//			logx.String("trace_id", info.TraceID),
//			logx.String("span_id", info.SpanID),
//		)
//	}
//
// # 与 OpenTelemetry 的区别
//
// trace 包是轻量级的追踪信息传递工具，适用于：
//   - 简单的链路追踪需求
//   - 不想引入重量级依赖
//   - 只需要传递追踪 ID 和记录基本耗时
//
// 如果需要完整的分布式追踪功能（采样、导出、可视化），建议使用 OpenTelemetry。
//
// # 性能说明
//
//   - context.WithValue 的性能开销很小
//   - 追踪信息只是简单的字符串传递
//   - 不涉及网络调用或持久化
//   - 适合高频调用场景
//
// # 线程安全
//
// 所有函数都是线程安全的，可以在多个 goroutine 中并发使用。
// context.Context 本身是不可变的，可以安全地在 goroutine 之间传递。
package trace
