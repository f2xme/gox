/*
Package zap 提供基于 uber-go/zap 的 logx.Logger 实现。

zap 是一个高性能的结构化日志库，本适配器将其封装为 logx.Logger 接口。

# 功能特性

  - 高性能：基于 uber-go/zap，提供极高的日志性能
  - 结构化日志：原生支持 JSON 格式输出
  - 日志轮转：集成 lumberjack，支持按大小、时间轮转
  - 异步缓冲：可选的异步缓冲模式，进一步提升性能
  - 多输出：支持同时输出到控制台和文件
  - 灵活配置：通过 Options 模式灵活配置

# 快速开始

基本使用：

	package main

	import (
		"github.com/f2xme/gox/logx"
		"github.com/f2xme/gox/logx/adapter/zap"
	)

	func main() {
		// 创建 logger
		logger := zap.New(
			zap.WithInfoLevel(),
		)
		logx.Init(logger)

		// 记录日志
		logx.Info("server started", logx.NewKV("port", 8080))
	}

输出到文件：

	logger := zap.New(
		zap.WithInfoLevel(),
		zap.WithFile("/var/log/app.log"),
	)

日志轮转：

	logger := zap.New(
		zap.WithInfoLevel(),
		zap.WithFileRotation(&zap.FileOption{
			Filename:   "/var/log/app.log",
			MaxSize:    100,  // 100MB
			MaxBackups: 10,   // 保留 10 个旧文件
			MaxAge:     30,   // 保留 30 天
			Compress:   true, // 压缩旧文件
		}),
	)

# 高级用法

异步缓冲模式：

	logger := zap.New(
		zap.WithInfoLevel(),
		zap.WithFile("/var/log/app.log"),
		zap.WithAsyncBuffer(),
		zap.WithBufferSize(512 * 1024),  // 512KB 缓冲区
		zap.WithFlushInterval(5 * time.Second),
	)

	// 程序退出前刷新缓冲区
	defer logx.Flush()

多输出配置：

	logger := zap.NewLoggers(
		// 控制台输出（Info 级别）
		[]zap.Option{
			zap.WithInfoLevel(),
		},
		// 文件输出（Error 级别）
		[]zap.Option{
			zap.WithErrorLevel(),
			zap.WithFile("/var/log/error.log"),
			zap.WithDisableConsole(),
		},
	)

从配置文件创建：

	import "github.com/f2xme/gox/config/viper"

	cfg := viper.New("config.yaml")
	logger := zap.NewWithConfig(cfg)

配置文件示例（YAML）：

	log:
	  level: info
	  file: /var/log/app.log
	  disableConsole: false
	  timeLayout: "2006-01-02 15:04:05"
	  asyncBuffer: true
	  bufferSize: 262144  # 256KB
	  flushInterval: 5s
	  file:
	    maxSize: 100      # MB
	    maxBackups: 10
	    maxAge: 30        # days
	    compress: true

# 配置选项

日志级别：

  - WithDebugLevel() - 调试级别
  - WithInfoLevel() - 信息级别（默认）
  - WithWarnLevel() - 警告级别
  - WithErrorLevel() - 错误级别

输出配置：

  - WithFile(filename) - 输出到文件
  - WithFileRotation(opt) - 输出到文件并启用轮转
  - WithDisableConsole() - 禁用控制台输出

格式配置：

  - WithTimeLayout(layout) - 自定义时间格式

性能优化：

  - WithAsyncBuffer() - 启用异步缓冲
  - WithBufferSize(size) - 设置缓冲区大小
  - WithFlushInterval(d) - 设置刷新间隔

# 注意事项

使用异步缓冲时，务必在程序退出前调用 Flush：

	logger := zap.New(zap.WithAsyncBuffer())
	defer logx.Flush()

日志轮转的默认值：

  - MaxSize: 10MB
  - MaxBackups: 10 个文件
  - MaxAge: 10 天

# 性能建议

  - 生产环境建议使用 Info 或 Warn 级别
  - 高并发场景建议启用异步缓冲
  - 避免在循环中创建 Meta 对象
  - 使用结构化日志而非字符串拼接
*/
package zap
