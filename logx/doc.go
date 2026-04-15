/*
Package logx 提供统一的日志记录抽象层。

logx 包定义了日志记录的标准接口，支持多种日志库（zap、logrus、zerolog 等）。
通过这些接口，你可以轻松地在不同的日志实现之间切换，而无需修改业务代码。

# 功能特性

  - 统一的日志接口：定义标准的 Logger 接口，支持多种日志库实现
  - 结构化日志：通过 Meta 接口支持键值对形式的结构化字段
  - Context 集成：自动从 context 提取字段，支持请求级别的日志追踪
  - 全局日志：提供包级别的日志函数，简化常见场景的使用
  - 多种适配器：内置 zap 等主流日志库的适配器
  - 线程安全：所有实现都是并发安全的

# 快速开始

基本使用：

	package main

	import (
		"github.com/f2xme/gox/logx"
		"github.com/f2xme/gox/logx/adapter/zap"
	)

	func main() {
		// 创建 logger
		logger := zap.New()
		logx.Init(logger)

		// 记录日志
		logx.Info("server started", logx.NewKV("port", 8080))
		logx.Warn("cache miss", logx.NewKV("key", "user:123"))

		// 错误日志
		if err := doSomething(); err != nil {
			logx.Error(err, logx.NewKV("operation", "db_query"))
		}
	}

使用 Context 记录日志：

	func HandleRequest(ctx context.Context) {
		// 使用 context 记录日志，自动添加调用位置
		logx.InfoCtx(ctx, "handling request",
			logx.NewKV("user_id", "123"),
			logx.NewKV("request_id", "abc"),
		)

		// 或者创建 ContextLogger
		logger := logx.Ctx(ctx).With(logx.NewKV("user_id", "123"))
		logger.Info("processing")
	}

# 核心接口

Logger 接口定义了日志记录的标准方法：

	type Logger interface {
		Info(msg string, metas ...Meta)
		Warn(msg string, metas ...Meta)
		Error(err error, metas ...Meta)
		Fatal(err error, metas ...Meta)
	}

Meta 接口用于添加结构化字段：

	type Meta interface {
		Key() string
		Value() any
	}

# 可用适配器

Zap 适配器：

	import "github.com/f2xme/gox/logx/adapter/zap"

	logger := zap.New(
		zap.WithInfoLevel(),
		zap.WithFile("/var/log/app.log"),
	)

# 最佳实践

使用结构化日志：

	// 推荐：结构化日志
	logger.Info("user login",
		logx.NewKV("user_id", userID),
		logx.NewKV("ip", ip),
	)

	// 不推荐：字符串拼接
	logger.Info(fmt.Sprintf("user %s login from %s", userID, ip))

使用 Context 传递日志字段：

	// 设置 context 提取器
	logx.SetContextExtractor(func(ctx context.Context) []logx.Meta {
		if reqID := ctx.Value("request_id"); reqID != nil {
			return []logx.Meta{logx.NewKV("request_id", reqID)}
		}
		return nil
	})

	// 使用 context 记录日志
	logx.InfoCtx(ctx, "processing request")

优雅关闭：

	logger := zap.New()
	defer func() {
		if err := logx.Flush(); err != nil {
			// 处理错误
		}
	}()

# 日志级别

	DebugLevel  // 调试信息
	InfoLevel   // 一般信息
	WarnLevel   // 警告信息
	ErrorLevel  // 错误信息

# 线程安全

所有日志实现都应该是线程安全的，可以在多个 goroutine 中并发使用。
*/
package logx
