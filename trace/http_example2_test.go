package trace_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/f2xme/gox/trace"
)

func Example_httpClientWithTracing_complete() {
	// 完整的 HTTP 客户端追踪示例
	ctx := context.Background()
	ctx = trace.WithTraceID(ctx, "client-trace-123")
	ctx = trace.WithSpanID(ctx, "client-span-456")

	// 创建请求
	req, err := http.NewRequest(http.MethodGet, "http://api.example.com/users", nil)
	if err != nil {
		fmt.Println("error creating request:", err)
		return
	}

	// 手动注入追踪信息到请求头
	trace.InjectToHeaders(ctx, req.Header)

	// 打印注入的头信息
	fmt.Printf("X-Trace-ID: %s\n", req.Header.Get(trace.HeaderTraceID))
	fmt.Printf("X-Span-ID: %s\n", req.Header.Get(trace.HeaderSpanID))

	// Output:
	// X-Trace-ID: client-trace-123
	// X-Span-ID: client-span-456
}
