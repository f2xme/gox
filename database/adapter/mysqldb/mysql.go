package mysqldb

import (
	"github.com/f2xme/gox/database"
	"github.com/f2xme/gox/database/adapter/gormbase"
	"gorm.io/driver/mysql"
)

// New 创建由 MySQL 支持的 database.DB
// dsn 示例: "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
func New(dsn string, opts ...gormbase.Option) (database.DB, error) {
	cfg := gormbase.DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return gormbase.New(mysql.Open(dsn), "mysql "+dsn, cfg)
}

// MustNew 创建由 MySQL 支持的 database.DB，失败时终止程序
func MustNew(dsn string, opts ...gormbase.Option) database.DB {
	cfg := gormbase.DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return gormbase.MustNew(mysql.Open(dsn), "mysql "+dsn, cfg)
}
