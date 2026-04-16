package sqlitedb

import (
	"github.com/f2xme/gox/config"
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

// NewWithConfig 使用 config.Config 中的配置创建 SQLite 数据库连接
// 配置键详见 gormbase.NewWithConfig 文档
// 注意：SQLite 默认使用单连接配置以避免并发问题
func NewWithConfig(file string, cfg config.Config) (database.DB, error) {
	return gormbase.NewWithConfig(sqlite.Open(file), "sqlite "+file, cfg)
}

// MustNewWithConfig 是 NewWithConfig 的 Must 版本，失败时终止程序
func MustNewWithConfig(file string, cfg config.Config) database.DB {
	return gormbase.MustNewWithConfig(sqlite.Open(file), "sqlite "+file, cfg)
}
