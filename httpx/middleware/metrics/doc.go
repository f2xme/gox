/*
Package metrics 提供 HTTP 指标收集中间件。

# 概述

metrics 中间件自动收集 HTTP 请求的关键指标，包括请求计数、响应时间、错误率和响应大小。
支持自定义收集器、路径归一化、业务指标扩展等功能。

# 快速开始

基本用法：

	import (
		"github.com/f2xme/gox/httpx/middleware/metrics"
		"github.com/f2xme/gox/httpx/adapter/gin"
	)

	func main() {
		app := gin.New()

		// 使用默认内存收集器
		app.Use(metrics.New())

		// 或使用自定义收集器
		app.Use(metrics.New(
			metrics.WithCollector(myCollector),
		))

		app.Run(":8080")
	}

# 配置选项

## WithCollector - 设置收集器

指定自定义的指标收集器：

	app.Use(metrics.New(
		metrics.WithCollector(prometheusCollector),
	))

默认使用内存收集器（NewMemoryCollector）。

## WithSkipPaths - 跳过特定路径

排除健康检查、指标端点等不需要监控的路径：

	app.Use(metrics.New(
		metrics.WithSkipPaths("/health", "/metrics", "/ping"),
	))

## WithPathNormalizer - 路径归一化

将动态路径参数归一化，避免指标爆炸：

	app.Use(metrics.New(
		metrics.WithPathNormalizer(func(path string) string {
			// /users/123 -> /users/:id
			// /posts/456/comments/789 -> /posts/:id/comments/:id
			return normalizePath(path)
		}),
	))

## WithDetailedMetrics - 启用详细指标

收集响应大小等额外指标：

	app.Use(metrics.New(
		metrics.WithDetailedMetrics(true),
	))

注意：详细指标会增加性能开销。

## WithCustomLabels - 自定义标签

从请求上下文提取自定义标签：

	app.Use(metrics.New(
		metrics.WithCustomLabels(func(ctx any) map[string]string {
			// 从上下文提取租户 ID、用户 ID 等
			return map[string]string{
				"tenant": getTenantID(ctx),
				"region": getRegion(ctx),
			}
		}),
	))

## WithBusinessMetrics - 业务指标

记录自定义业务指标：

	app.Use(metrics.New(
		metrics.WithBusinessMetrics(func(ctx any, collector Collector) {
			// 记录订单金额、商品数量等业务指标
			if order := getOrder(ctx); order != nil {
				collector.RecordCustom("order_amount", order.Amount)
			}
		}),
	))

# 收集的指标

默认收集以下指标：

  - 请求计数：按方法和路径分组
  - 请求时长：按方法和路径分组
  - 错误计数：按方法和路径分组
  - 响应大小：启用详细指标时收集

# 收集器接口

实现 Collector 接口以支持不同的指标后端：

	type Collector interface {
		RecordRequest(method, path string)
		RecordDuration(method, path string, duration time.Duration)
		RecordError(method, path string)
		RecordResponseSize(method, path string, size int64)
	}

内置收集器：

  - MemoryCollector：内存收集器，适用于开发和测试
  - 可集成 Prometheus、StatsD 等第三方收集器

# 使用示例

## 完整配置示例

	app.Use(metrics.New(
		metrics.WithCollector(prometheusCollector),
		metrics.WithSkipPaths("/health", "/metrics"),
		metrics.WithPathNormalizer(normalizePath),
		metrics.WithDetailedMetrics(true),
		metrics.WithCustomLabels(func(ctx any) map[string]string {
			return map[string]string{
				"service": "api",
				"env":     os.Getenv("ENV"),
			}
		}),
	))

## 与 Prometheus 集成

	import (
		"github.com/prometheus/client_golang/prometheus"
		"github.com/prometheus/client_golang/prometheus/promhttp"
	)

	// 创建 Prometheus 收集器
	collector := NewPrometheusCollector()

	// 注册中间件
	app.Use(metrics.New(
		metrics.WithCollector(collector),
	))

	// 暴露指标端点
	app.GET("/metrics", gin.WrapH(promhttp.Handler()))

# 最佳实践

## 1. 路径归一化

避免动态路径参数导致指标维度爆炸：

	// 推荐：归一化路径
	metrics.WithPathNormalizer(func(path string) string {
		return regexp.MustCompile(`/\d+`).ReplaceAllString(path, "/:id")
	})

	// 不推荐：不归一化，导致 /users/1, /users/2, ... 产生大量指标

## 2. 跳过无关路径

排除健康检查等高频低价值端点：

	metrics.WithSkipPaths("/health", "/ping", "/ready")

## 3. 合理使用详细指标

仅在需要时启用详细指标，避免性能开销：

	// 生产环境：关闭详细指标
	metrics.WithDetailedMetrics(false)

	// 调试环境：启用详细指标
	metrics.WithDetailedMetrics(true)

## 4. 自定义标签控制维度

避免高基数标签（如用户 ID、请求 ID）：

	// 推荐：低基数标签
	metrics.WithCustomLabels(func(ctx any) map[string]string {
		return map[string]string{
			"tenant": getTenantID(ctx),  // 有限数量
			"region": getRegion(ctx),    // 有限数量
		}
	})

	// 不推荐：高基数标签
	metrics.WithCustomLabels(func(ctx any) map[string]string {
		return map[string]string{
			"user_id": getUserID(ctx),  // 无限增长
		}
	})

# 性能考虑

  - 默认配置性能开销极小（微秒级）
  - 详细指标会增加少量开销
  - 路径归一化函数应尽量高效
  - 自定义标签提取应避免复杂计算

# 线程安全

所有收集器实现都应是线程安全的，支持并发调用。
*/
package metrics
