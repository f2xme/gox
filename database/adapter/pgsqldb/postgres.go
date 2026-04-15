package pgsqldb

import (
	"github.com/f2xme/gox/database"
	"github.com/f2xme/gox/database/adapter/gormbase"
	"gorm.io/driver/postgres"
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
