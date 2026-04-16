package zap

import (
	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/logx"
	gozap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New 创建一个基于 zap 的 logx.Logger
//
// 示例：
//
//	logger := zap.New(
//		zap.WithInfoLevel(),
//		zap.WithFile("/var/log/app.log"),
//	)
func New(opts ...Option) logx.Logger {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(&cfg)
	}
	result := buildCore(&cfg)
	logger := gozap.New(result.core, gozap.AddCallerSkip(cfg.CallerSkip))
	return &zapLogger{
		logger:          logger,
		bufferedWriters: result.bufferedWriters,
	}
}

// NewLoggers 创建一个支持多个输出配置的 logx.Logger
//
// 可以同时输出到控制台和文件，或者多个不同配置的文件。
//
// 示例：
//
//	logger := zap.NewLoggers(
//		[]zap.Option{zap.WithInfoLevel()},  // 控制台
//		[]zap.Option{zap.WithFile("/var/log/app.log"), zap.WithDisableConsole()},  // 文件
//	)
func NewLoggers(optSets ...[]Option) logx.Logger {
	var allCores []zapcore.Core
	var allBuffered []*zapcore.BufferedWriteSyncer

	for _, opts := range optSets {
		cfg := defaultOptions()
		for _, opt := range opts {
			opt(&cfg)
		}
		result := buildCore(&cfg)
		allCores = append(allCores, result.core)
		allBuffered = append(allBuffered, result.bufferedWriters...)
	}

	logger := gozap.New(zapcore.NewTee(allCores...))
	return &zapLogger{
		logger:          logger,
		bufferedWriters: allBuffered,
	}
}

// NewWithConfig 从配置文件创建 logx.Logger
//
// 支持的配置项：
//   - log.level (string): 日志级别 - "debug", "info", "warn", "error"（默认 "info"）
//   - log.file (string): 日志文件路径（可选）
//   - log.disableConsole (bool): 禁用控制台输出（默认 false）
//   - log.timeLayout (string): 时间格式（可选）
//   - log.callerSkip (int): 调用栈跳过层数（默认 0）
//   - log.asyncBuffer (bool): 启用异步缓冲（默认 false）
//   - log.bufferSize (int): 缓冲区大小（默认 256KB）
//   - log.flushInterval (duration): 刷新间隔（默认 5s）
//   - log.file.maxSize (int): 单个文件最大大小 MB（默认 10）
//   - log.file.maxBackups (int): 保留的旧文件数量（默认 10）
//   - log.file.maxAge (int): 保留旧文件的天数（默认 10）
//   - log.file.compress (bool): 压缩轮转文件（默认 false）
//   - log.file.localTime (bool): 使用本地时间轮转（默认 false）
func NewWithConfig(cfg config.Config) logx.Logger {
	opts := []Option{}

	// Set log level
	level := logx.ParseLevel(cfg.GetString("log.level"))
	opts = append(opts, WithLevel(level))

	// File logging with rotation
	if filename := cfg.GetString("log.file"); filename != "" {
		if cfg.GetInt("log.file.maxSize") > 0 || cfg.GetInt("log.file.maxBackups") > 0 {
			fileOpt := &FileOption{
				Filename:   filename,
				MaxSize:    cfg.GetInt("log.file.maxSize"),
				MaxBackups: cfg.GetInt("log.file.maxBackups"),
				MaxAge:     cfg.GetInt("log.file.maxAge"),
				Compress:   cfg.GetBool("log.file.compress"),
				LocalTime:  cfg.GetBool("log.file.localTime"),
			}
			opts = append(opts, WithFileRotation(fileOpt))
		} else {
			opts = append(opts, WithFile(filename))
		}
	}

	// Console output
	if cfg.GetBool("log.disableConsole") {
		opts = append(opts, WithDisableConsole())
	}

	// Time layout
	if layout := cfg.GetString("log.timeLayout"); layout != "" {
		opts = append(opts, WithTimeLayout(layout))
	}

	// Caller skip
	if skip := cfg.GetInt("log.callerSkip"); skip > 0 {
		opts = append(opts, WithCallerSkip(skip))
	}

	// Async buffer
	if cfg.GetBool("log.asyncBuffer") {
		opts = append(opts, WithAsyncBuffer())
		if size := cfg.GetInt("log.bufferSize"); size > 0 {
			opts = append(opts, WithBufferSize(size))
		}
		if interval := cfg.GetDuration("log.flushInterval"); interval > 0 {
			opts = append(opts, WithFlushInterval(interval))
		}
	}

	return New(opts...)
}
