// Package testkit 提供面向 httpx.Engine 的 HTTP 集成测试工具。
//
// testkit 使用 net/http/httptest 启动真实 HTTP 服务，测试会经过路由、
// 中间件、请求绑定、错误处理和响应写出等完整链路。它适合黑盒测试
// httpx 应用，也可通过 RunEngineSuite 复用同一组测试用例验证多个
// httpx adapter 的行为一致性。
//
// 基本用法：
//
//	client := testkit.New(t, engine)
//	defer client.Close()
//
//	client.GET("/users/123").
//		ExpectStatus(200).
//		ExpectJSONValue("success", true)
//
//	client.POSTJSON("/users", map[string]any{"name": "Alice"}).
//		ExpectStatus(201)
//
// 对于没有专用 helper 的方法，可使用 Do：
//
//	client.Do(http.MethodTrace, "/debug", nil).
//		ExpectStatus(200)
package testkit
