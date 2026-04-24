package sqlitedb

import (
	"time"

	"github.com/f2xme/gox/database/adapter/internal/gormbase"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Options struct {
	gormbase.Options
	gormConfig   *gorm.Config
	sqliteConfig *sqlite.Config
	file         string
}

// defaultOptions 返回 SQLite 的默认配置（单连接以降低并发写入风险）
func defaultOptions() Options {
	return Options{
		Options: gormbase.Options{
			MaxOpenConns:    1,
			MaxIdleConns:    1,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		gormConfig:   nil,
		sqliteConfig: nil,
		file:         "",
	}
}

// Option 定义 SQLite 数据库配置选项
type Option func(*Options)

// WithGormConfig 设置底层 gorm.Config
func WithGormConfig(c *gorm.Config) Option {
	return func(o *Options) {
		o.gormConfig = c
	}
}

// WithSQLiteConfig 设置 SQLite 数据库配置
func WithSQLiteConfig(c *sqlite.Config) Option {
	return func(o *Options) {
		o.sqliteConfig = c
	}
}

// WithFile 设置 SQLite 数据库文件路径
func WithFile(file string) Option {
	return func(o *Options) {
		o.file = file
	}
}

// WithMaxOpenConns 设置最大打开连接数（SQLite 建议为 1）
func WithMaxOpenConns(n int) Option {
	return func(o *Options) {
		o.MaxOpenConns = n
	}
}

// WithMaxIdleConns 设置最大空闲连接数（SQLite 建议为 1）
func WithMaxIdleConns(n int) Option {
	return func(o *Options) {
		o.MaxIdleConns = n
	}
}

// WithConnMaxLifetime 设置连接最大生命周期
func WithConnMaxLifetime(d time.Duration) Option {
	return func(o *Options) {
		o.ConnMaxLifetime = d
	}
}

// WithConnMaxIdleTime 设置连接最大空闲时间
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(o *Options) {
		o.ConnMaxIdleTime = d
	}
}

// WithOptions 设置多个选项
func WithOptions(opts ...Option) Option {
	return func(o *Options) {
		for _, opt := range opts {
			opt(o)
		}
	}
}
