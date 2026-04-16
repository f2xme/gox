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

异步缓冲模式：

	logger := zap.New(
		zap.WithInfoLevel(),
		zap.WithFile("/var/log/app.log"),
		zap.WithAsyncBuffer(),
	)
	// 程序退出前刷新缓冲区
	defer logx.Flush()

从配置文件创建：

	import "github.com/f2xme/gox/config/viper"

	cfg := viper.New("config.yaml")
	logger := zap.NewWithConfig(cfg)

# 注意事项

  - 使用异步缓冲时，务必在程序退出前调用 logx.Flush()
  - 日志轮转默认值：MaxSize 10MB、MaxBackups 10 个、MaxAge 10 天
  - 生产环境建议使用 Info 或 Warn 级别
  - 高并发场景建议启用异步缓冲
*/
package zap
