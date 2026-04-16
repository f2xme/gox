package pgsqldb

import (
	"gorm.io/driver/postgres"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/database"
	"github.com/f2xme/gox/database/adapter/gormbase"
)

// New 创建由 PostgreSQL 支持的 database.DB
// dsn 示例: "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
func New(dsn string, opts ...gormbase.Option) (database.DB, error) {
	cfg := gormbase.DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return gormbase.New(postgres.Open(dsn), "postgres "+dsn, cfg)
}

// MustNew 创建由 PostgreSQL 支持的 database.DB，失败时终止程序
func MustNew(dsn string, opts ...gormbase.Option) database.DB {
	cfg := gormbase.DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return gormbase.MustNew(postgres.Open(dsn), "postgres "+dsn, cfg)
}

// NewWithConfig 使用 config.Config 中的配置创建 PostgreSQL 数据库连接
// 配置键详见 gormbase.NewWithConfig 文档
func NewWithConfig(dsn string, cfg config.Config) (database.DB, error) {
	return gormbase.NewWithConfig(postgres.Open(dsn), "postgres "+dsn, cfg)
}

// MustNewWithConfig 是 NewWithConfig 的 Must 版本，失败时终止程序
func MustNewWithConfig(dsn string, cfg config.Config) database.DB {
	return gormbase.MustNewWithConfig(postgres.Open(dsn), "postgres "+dsn, cfg)
}
