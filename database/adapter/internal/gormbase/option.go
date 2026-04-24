package gormbase

import (
	"errors"
	"time"
)

type DBType string

const (
	DBTypeMySQL DBType = "mysql"
	DBTypePostgres DBType = "postgres"
	DBTypeSqlite DBType = "sqlite"
)

// Options 定义数据库连接池配置选项
type Options struct {
	// MaxOpenConns 最大打开连接数，默认 100
	MaxOpenConns int
	// MaxIdleConns 最大空闲连接数，默认 10
	MaxIdleConns int
	// ConnMaxLifetime 连接最大生命周期，默认 1 小时
	ConnMaxLifetime time.Duration
	// ConnMaxIdleTime 连接最大空闲时间，默认 10 分钟
	ConnMaxIdleTime time.Duration
	// IgnoreRecordNotFound 是否忽略 ErrRecordNotFound 错误，默认 false
	IgnoreRecordNotFound bool
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

// DefaultOptions 返回默认连接池配置
func DefaultOptions() *Options {
	opts := defaultOptions()
	return &opts
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

// WithIgnoreRecordNotFound 设置是否忽略 ErrRecordNotFound 错误
func WithIgnoreRecordNotFound(ignore bool) Option {
	return func(o *Options) { o.IgnoreRecordNotFound = ignore }
}
