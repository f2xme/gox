package mysqldb

import (
	"time"

	"github.com/f2xme/gox/database/adapter/internal/gormbase"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Options struct {
	gormbase.Options
	gormConfig  *gorm.Config
	mysqlConfig *mysql.Config
	dsn         string
}

func defaultOptions() Options {
	return Options{
		Options:     *gormbase.DefaultOptions(),
		gormConfig:  nil,
		mysqlConfig: nil,
		dsn:         "",
	}
}

// Option 定义 MySQL 数据库配置选项
type Option func(*Options)

// WithGormConfig 设置底层 gorm.Config
func WithGormConfig(c *gorm.Config) Option {
	return func(o *Options) {
		o.gormConfig = c
	}
}

// WithMySQLConfig 设置 MySQL 数据库配置
func WithMySQLConfig(c *mysql.Config) Option {
	return func(o *Options) {
		o.mysqlConfig = c
	}
}

// WithDSN 设置 MySQL 数据库 DSN
func WithDSN(dsn string) Option {
	return func(o *Options) {
		o.dsn = dsn
	}
}

// WithMaxOpenConns 设置最大打开连接数
func WithMaxOpenConns(n int) Option {
	return func(o *Options) {
		o.MaxOpenConns = n
	}
}

// WithMaxIdleConns 设置最大空闲连接数
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
