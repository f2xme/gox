package zap

import (
	"os"
	"sync"
	"time"

	goxconfig "github.com/f2xme/gox/config"
	"github.com/f2xme/gox/logx"
	gozap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger          *gozap.Logger
	bufferedWriters []*zapcore.BufferedWriteSyncer
	mu              sync.Mutex
}

var (
	_ logx.Logger  = (*zapLogger)(nil)
	_ logx.Flusher = (*zapLogger)(nil)
	_ logx.Syncer  = (*zapLogger)(nil)
	_ logx.Stopper = (*zapLogger)(nil)
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

func (l *zapLogger) Info(msg string, metas ...logx.Meta) {
	l.logger.Info(msg, toFields(metas)...)
}

func (l *zapLogger) Warn(msg string, metas ...logx.Meta) {
	l.logger.Warn(msg, toFields(metas)...)
}

func (l *zapLogger) Error(err error, metas ...logx.Meta) {
	l.logger.Error(err.Error(), toFields(metas)...)
}

func (l *zapLogger) Fatal(err error, metas ...logx.Meta) {
	l.logger.Fatal(err.Error(), toFields(metas)...)
}

func (l *zapLogger) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	var lastErr error
	for _, bw := range l.bufferedWriters {
		if err := bw.Sync(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

func (l *zapLogger) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	var lastErr error
	for _, bw := range l.bufferedWriters {
		if err := bw.Stop(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func toFields(metas []logx.Meta) []gozap.Field {
	if len(metas) == 0 {
		return nil
	}
	fields := make([]gozap.Field, 0, len(metas))
	for _, m := range metas {
		fields = append(fields, gozap.Any(m.Key(), m.Value()))
	}
	return fields
}

type coreResult struct {
	core            zapcore.Core
	bufferedWriters []*zapcore.BufferedWriteSyncer
}

func buildCore(cfg *Options) *coreResult {
	result := &coreResult{}

	encoderCfg := gozap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(cfg.TimeLayout))
	}
	jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)

	levelEnabler := gozap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= cfg.Level
	})

	var cores []zapcore.Core

	if !cfg.DisableConsole {
		cores = append(cores, zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), levelEnabler))
	}

	if cfg.File != nil {
		var ws zapcore.WriteSyncer
		if cfg.Buffer != nil {
			bws := &zapcore.BufferedWriteSyncer{
				WS:            zapcore.AddSync(cfg.File),
				Size:          cfg.Buffer.Size,
				FlushInterval: cfg.Buffer.Interval,
			}
			result.bufferedWriters = append(result.bufferedWriters, bws)
			ws = bws
		} else {
			ws = zapcore.AddSync(cfg.File)
		}
		cores = append(cores, zapcore.NewCore(jsonEncoder, ws, levelEnabler))
	}

	if len(cores) == 0 {
		cores = append(cores, zapcore.NewNopCore())
	}

	result.core = zapcore.NewTee(cores...)
	return result
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
func NewWithConfig(cfg goxconfig.Config) logx.Logger {
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
