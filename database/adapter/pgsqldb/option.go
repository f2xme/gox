package pgsqldb

import (
	"time"

	"github.com/f2xme/gox/database/adapter/internal/gormbase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Options struct {
	gormbase.Options
	gormConfig     *gorm.Config
	postgresConfig *postgres.Config
	dsn            string
}

func defaultOptions() Options {
	return Options{
		Options:        *gormbase.DefaultOptions(),
		gormConfig:     nil,
		postgresConfig: nil,
		dsn:            "",
	}
}

// Option 定义 PostgreSQL 数据库配置选项
type Option func(*Options)

// WithGormConfig 设置底层 gorm.Config
func WithGormConfig(c *gorm.Config) Option {
	return func(o *Options) {
		o.gormConfig = c
	}
}

// WithPostgresConfig 设置 PostgreSQL 数据库配置
func WithPostgresConfig(c *postgres.Config) Option {
	return func(o *Options) {
		o.postgresConfig = c
	}
}

// WithDSN 设置 PostgreSQL 数据库 DSN
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

// WithIgnoreRecordNotFound 设置是否忽略 ErrRecordNotFound 错误
func WithIgnoreRecordNotFound(ignore bool) Option {
	return func(o *Options) {
		o.IgnoreRecordNotFound = ignore
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
