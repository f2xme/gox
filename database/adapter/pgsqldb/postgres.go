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
// 配置键：
//   - db.maxOpenConns (int): 最大打开连接数
//   - db.maxIdleConns (int): 最大空闲连接数
//   - db.connMaxLifetime (duration): 连接最大生命周期
//   - db.connMaxIdleTime (duration): 连接最大空闲时间
//   - db.dryRun (bool): 启用 DryRun 模式
//   - db.singularTable (bool): 使用单数表名
//   - db.tablePrefix (string): 表名前缀
func NewWithConfig(dsn string, cfg config.Config) (database.DB, error) {
	return gormbase.NewWithConfig(postgres.Open(dsn), "postgres "+dsn, cfg)
}

// MustNewWithConfig 使用配置创建 PostgreSQL 数据库连接，失败时终止程序
func MustNewWithConfig(dsn string, cfg config.Config) database.DB {
	return gormbase.MustNewWithConfig(postgres.Open(dsn), "postgres "+dsn, cfg)
}
