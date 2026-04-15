package zap

import (
	"io"
	"time"

	"github.com/f2xme/gox/logx"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultBufferSize    = 256 * 1024
	defaultFlushInterval = 5 * time.Second
	defaultTimeLayout    = "2006-01-02 15:04:05"
)

// BufferConfig 定义异步缓冲配置
type BufferConfig struct {
	Size     int
	Interval time.Duration
}

// Options 定义 zap 适配器的配置选项
type Options struct {
	Level          zapcore.Level
	File           io.Writer
	TimeLayout     string
	DisableConsole bool
	CallerSkip     int
	Buffer         *BufferConfig // nil 表示禁用缓冲
}

// Option 定义配置选项函数
type Option func(*Options)

func defaultOptions() Options {
	return Options{
		Level:      zapcore.InfoLevel,
		TimeLayout: defaultTimeLayout,
	}
}

// WithLevel 设置日志级别
//
// 示例：
//
//	New(WithLevel(logx.DebugLevel))
func WithLevel(level logx.Level) Option {
	return func(c *Options) {
		switch level {
		case logx.DebugLevel:
			c.Level = zapcore.DebugLevel
		case logx.InfoLevel:
			c.Level = zapcore.InfoLevel
		case logx.WarnLevel:
			c.Level = zapcore.WarnLevel
		case logx.ErrorLevel:
			c.Level = zapcore.ErrorLevel
		default:
			c.Level = zapcore.InfoLevel
		}
	}
}

// WithDebugLevel 设置日志级别为 Debug
func WithDebugLevel() Option {
	return WithLevel(logx.DebugLevel)
}

// WithInfoLevel 设置日志级别为 Info
func WithInfoLevel() Option {
	return WithLevel(logx.InfoLevel)
}

// WithWarnLevel 设置日志级别为 Warn
func WithWarnLevel() Option {
	return WithLevel(logx.WarnLevel)
}

// WithErrorLevel 设置日志级别为 Error
func WithErrorLevel() Option {
	return WithLevel(logx.ErrorLevel)
}

// WithFile 设置日志文件路径
//
// 示例：
//
//	New(WithFile("/var/log/app.log"))
func WithFile(filename string) Option {
	return func(c *Options) {
		c.File = &lumberjack.Logger{Filename: filename}
	}
}

// WithFileRotation 设置日志文件轮转配置
//
// 示例：
//
//	New(WithFileRotation(&FileOption{
//		Filename:   "/var/log/app.log",
//		MaxSize:    100,  // 100MB
//		MaxBackups: 10,
//		MaxAge:     30,   // 30 days
//		Compress:   true,
//	}))
func WithFileRotation(opt *FileOption) Option {
	return func(c *Options) {
		c.File = &lumberjack.Logger{
			Filename:   opt.Filename,
			MaxSize:    defaultInt(opt.MaxSize, 10),
			MaxBackups: defaultInt(opt.MaxBackups, 10),
			MaxAge:     defaultInt(opt.MaxAge, 10),
			LocalTime:  opt.LocalTime,
			Compress:   opt.Compress,
		}
	}
}

func defaultInt(val, def int) int {
	if val == 0 {
		return def
	}
	return val
}

// WithDisableConsole 禁用控制台输出
func WithDisableConsole() Option {
	return func(c *Options) { c.DisableConsole = true }
}

// WithTimeLayout 设置时间格式
//
// 示例：
//
//	New(WithTimeLayout(time.RFC3339))
func WithTimeLayout(layout string) Option {
	return func(c *Options) { c.TimeLayout = layout }
}

// WithCallerSkip 设置调用栈跳过层数
func WithCallerSkip(skip int) Option {
	return func(c *Options) { c.CallerSkip = skip }
}

// WithAsyncBuffer 启用异步缓冲
//
// 启用后日志会先写入缓冲区，定期刷新到磁盘，提高性能。
// 使用默认配置：256KB 缓冲区，5 秒刷新间隔。
func WithAsyncBuffer() Option {
	return func(c *Options) {
		c.Buffer = &BufferConfig{
			Size:     defaultBufferSize,
			Interval: defaultFlushInterval,
		}
	}
}

// WithBufferConfig 设置自定义缓冲配置
//
// 示例：
//
//	New(WithBufferConfig(&BufferConfig{
//		Size:     512 * 1024,  // 512KB
//		Interval: 10 * time.Second,
//	}))
func WithBufferConfig(cfg *BufferConfig) Option {
	return func(c *Options) { c.Buffer = cfg }
}

// WithBufferSize 设置缓冲区大小（字节）
//
// 默认 256KB。如果未启用缓冲，会自动启用并使用默认刷新间隔。
func WithBufferSize(size int) Option {
	return func(c *Options) {
		if c.Buffer == nil {
			c.Buffer = &BufferConfig{
				Size:     size,
				Interval: defaultFlushInterval,
			}
		} else {
			c.Buffer.Size = size
		}
	}
}

// WithFlushInterval 设置缓冲区刷新间隔
//
// 默认 5 秒。如果未启用缓冲，会自动启用并使用默认缓冲区大小。
func WithFlushInterval(d time.Duration) Option {
	return func(c *Options) {
		if c.Buffer == nil {
			c.Buffer = &BufferConfig{
				Size:     defaultBufferSize,
				Interval: d,
			}
		} else {
			c.Buffer.Interval = d
		}
	}
}
