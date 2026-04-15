package gormbase

import (
	"errors"
	"time"

	"gorm.io/gorm/logger"
)

// Options 定义 GORM 数据库配置选项
type Options struct {
	// MaxOpenConns 最大打开连接数，默认 100
	MaxOpenConns int
	// MaxIdleConns 最大空闲连接数，默认 10
	MaxIdleConns int
	// ConnMaxLifetime 连接最大生命周期，默认 1 小时
	ConnMaxLifetime time.Duration
	// ConnMaxIdleTime 连接最大空闲时间，默认 10 分钟
	ConnMaxIdleTime time.Duration
	// Logger GORM 日志记录器
	Logger logger.Interface
	// DryRun 启用 DryRun 模式（不执行实际 SQL）
	DryRun bool
	// SingularTable 使用单数表名
	SingularTable bool
	// TablePrefix 表名前缀
	TablePrefix string
}

// Validate 验证配置选项的有效性
func (o *Options) Validate() error {
	if o.MaxOpenConns <= 0 {
		return errors.New("gormbase: MaxOpenConns must be positive")
	}
	if o.MaxIdleConns < 0 {
		return errors.New("gormbase: MaxIdleConns cannot be negative")
	}
	if o.MaxIdleConns > o.MaxOpenConns {
		return errors.New("gormbase: MaxIdleConns cannot exceed MaxOpenConns")
	}
	if o.ConnMaxLifetime < 0 {
		return errors.New("gormbase: ConnMaxLifetime cannot be negative")
	}
	if o.ConnMaxIdleTime < 0 {
		return errors.New("gormbase: ConnMaxIdleTime cannot be negative")
	}
	return nil
}

func defaultOptions() Options {
	return Options{
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Options {
	opts := defaultOptions()
	return &opts
}

// DefaultSQLiteConfig 返回 SQLite 的默认配置
// SQLite 使用单连接以避免并发问题
func DefaultSQLiteConfig() *Options {
	return &Options{
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// Option 定义配置选项函数
type Option func(*Options)

// WithMaxOpenConns 设置最大打开连接数
func WithMaxOpenConns(n int) Option { return func(o *Options) { o.MaxOpenConns = n } }

// WithMaxIdleConns 设置最大空闲连接数
func WithMaxIdleConns(n int) Option { return func(o *Options) { o.MaxIdleConns = n } }

// WithConnMaxLifetime 设置连接最大生命周期
func WithConnMaxLifetime(d time.Duration) Option {
	return func(o *Options) { o.ConnMaxLifetime = d }
}

// WithConnMaxIdleTime 设置连接最大空闲时间
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(o *Options) { o.ConnMaxIdleTime = d }
}

// WithLogger 设置 GORM 日志记录器
func WithLogger(l logger.Interface) Option { return func(o *Options) { o.Logger = l } }

// WithDryRun 启用 DryRun 模式
func WithDryRun() Option { return func(o *Options) { o.DryRun = true } }

// WithSingularTable 启用单数表名
func WithSingularTable() Option { return func(o *Options) { o.SingularTable = true } }

// WithTablePrefix 设置表名前缀
func WithTablePrefix(prefix string) Option { return func(o *Options) { o.TablePrefix = prefix } }
