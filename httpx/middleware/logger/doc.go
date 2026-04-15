/*
Package logger 提供 HTTP 请求日志中间件。

# 概述

logger 中间件记录每个 HTTP 请求的基本信息，包括方法、路径、客户端 IP、响应时间和错误。
通过 Logger 接口，你可以集成任何日志库（如 slog、zap、logrus）。

# 快速开始

基本用法：

	import (
		"log/slog"
		"github.com/f2xme/gox/httpx/middleware/logger"
	)

	// 使用标准库 slog
	type slogAdapter struct {
		logger *slog.Logger
	}

	func (s *slogAdapter) Info(msg string, keysAndValues ...any) {
		s.logger.Info(msg, keysAndValues...)
	}

	func main() {
		l := logger.New(
			logger.WithLogger(&slogAdapter{logger: slog.Default()}),
		)

		// 在 HTTP 框架中使用
		// engine.Use(l)
	}

# 配置选项

## WithLogger - 设置日志实现

指定日志记录器，必须实现 Logger 接口：

	logger.New(
		logger.WithLogger(myLogger),
	)

如果未设置 logger，中间件将不执行任何操作（no-op）。

## WithSkipPath - 跳过特定路径

排除不需要记录的路径（如健康检查）：

	logger.New(
		logger.WithLogger(myLogger),
		logger.WithSkipPath("/health", "/metrics", "/ping"),
	)

## WithSkipMethod - 跳过特定 HTTP 方法

排除不需要记录的 HTTP 方法：

	logger.New(
		logger.WithLogger(myLogger),
		logger.WithSkipMethod("OPTIONS", "HEAD"),
	)

# Logger 接口

实现此接口以集成你的日志库：

	type Logger interface {
		Info(msg string, keysAndValues ...any)
	}

keysAndValues 参数遵循结构化日志格式（key1, value1, key2, value2, ...）。

# 记录的字段

每个请求记录以下字段：

  - method: HTTP 方法（GET、POST 等）
  - path: 请求路径
  - ip: 客户端 IP 地址
  - duration: 请求处理时间
  - error: 错误信息（如果有）

# 集成示例

## 使用 slog

	import "log/slog"

	type slogAdapter struct {
		logger *slog.Logger
	}

	func (s *slogAdapter) Info(msg string, keysAndValues ...any) {
		s.logger.Info(msg, keysAndValues...)
	}

	l := logger.New(
		logger.WithLogger(&slogAdapter{logger: slog.Default()}),
	)

## 使用 zap

	import "go.uber.org/zap"

	type zapAdapter struct {
		logger *zap.SugaredLogger
	}

	func (z *zapAdapter) Info(msg string, keysAndValues ...any) {
		z.logger.Infow(msg, keysAndValues...)
	}

	zapLogger, _ := zap.NewProduction()
	l := logger.New(
		logger.WithLogger(&zapAdapter{logger: zapLogger.Sugar()}),
	)

# 最佳实践

## 1. 跳过健康检查端点

避免日志噪音：

	logger.New(
		logger.WithLogger(myLogger),
		logger.WithSkipPath("/health", "/readiness", "/liveness"),
	)

## 2. 跳过 OPTIONS 请求

CORS 预检请求通常不需要记录：

	logger.New(
		logger.WithLogger(myLogger),
		logger.WithSkipMethod("OPTIONS"),
	)

## 3. 使用结构化日志

推荐使用支持结构化日志的库（slog、zap），便于日志分析和查询。

## 4. 中间件顺序

将 logger 中间件放在错误恢复中间件之后，确保能记录所有请求（包括 panic 恢复后的请求）：

	engine.Use(recovery.New())  // 先恢复 panic
	engine.Use(logger.New(...)) // 再记录日志

# 性能考虑

  - 中间件开销极小（仅时间戳和字段收集）
  - 跳过路径/方法的检查使用 map，时间复杂度 O(1)
  - 如果未设置 logger，中间件直接透传，零开销
*/
package logger
