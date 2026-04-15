package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/f2xme/gox/metrics"
	"github.com/f2xme/gox/metrics/adapter/prometheus"
)

func must[T any](v T, err error) T {
	if err != nil {
		log.Fatalf("错误: %v", err)
	}
	return v
}

func main() {
	fmt.Println("=== metrics 包使用示例 ===")
	fmt.Println()

	adapter := prometheus.New(
		prometheus.WithNamespace("example"),
		prometheus.WithSubsystem("demo"),
	)

	ctx := context.Background()

	// 示例 1: Counter
	fmt.Println("示例 1: Counter")
	counter := must(adapter.Counter(ctx, "requests_total", metrics.Labels{
		"method": "GET",
		"path":   "/api/users",
	}))
	counter.Inc(ctx)
	counter.Add(ctx, 5)

	// 示例 2: Gauge
	fmt.Println("\n示例 2: Gauge")
	gauge := must(adapter.Gauge(ctx, "active_connections", nil))
	gauge.Set(ctx, 42)
	gauge.Inc(ctx)
	gauge.Dec(ctx)

	// 示例 3: Histogram
	fmt.Println("\n示例 3: Histogram")
	histogram := must(adapter.Histogram(ctx, "request_duration_seconds", metrics.Labels{
		"endpoint": "/users",
	}))
	start := time.Now()
	time.Sleep(100 * time.Millisecond)
	histogram.Observe(ctx, time.Since(start).Seconds())

	// 示例 4: 并发使用
	fmt.Println("\n示例 4: 并发场景")
	for range 10 {
		go counter.Inc(ctx)
	}
	time.Sleep(100 * time.Millisecond)

	// 启动 HTTP 服务器
	fmt.Println("\n启动 HTTP 服务器...")
	fmt.Println("访问 http://localhost:8080/metrics 查看指标")
	fmt.Println("按 Ctrl+C 退出")
	fmt.Println()

	server := &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}
	http.Handle("/metrics", adapter.Handler())

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP 服务器错误: %v", err)
	}
}
