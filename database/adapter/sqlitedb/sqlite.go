package sqlitedb

import (
	"github.com/f2xme/gox/database"
	"github.com/f2xme/gox/database/adapter/gormbase"
	"gorm.io/driver/sqlite"
)

// New 创建由 SQLite 支持的 database.DB
func New(file string, opts ...gormbase.Option) (database.DB, error) {
	cfg := gormbase.DefaultSQLiteConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return gormbase.New(sqlite.Open(file), "sqlite "+file, cfg)
}

// MustNew 创建由 SQLite 支持的 database.DB，失败时终止程序
func MustNew(file string, opts ...gormbase.Option) database.DB {
	cfg := gormbase.DefaultSQLiteConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return gormbase.MustNew(sqlite.Open(file), "sqlite "+file, cfg)
}
