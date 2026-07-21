/*
Package traceid 提供 HTTP Trace ID 中间件。

中间件优先复用请求头中的 Trace ID；请求头缺失时生成新的 128 位随机 ID，
并将 ID 写入响应头和 httpx.Context。

基本用法：

	engine.Use(traceid.New())

	engine.GET("/hello", func(ctx httpx.Context) error {
		id := traceid.Get(ctx)
		return ctx.String(http.StatusOK, id)
	})

默认请求头为 X-Trace-ID，上下文键为 trace_id。
*/
package traceid
