package pgsqldb

import (
	"fmt"
	"log"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/database"
	"github.com/f2xme/gox/database/adapter/internal/gormbase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// New 创建由 PostgreSQL 支持的 database.DB
// dsn 示例: "host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable TimeZone=Asia/Shanghai"
// 若使用 WithPostgresConfig 且配置中已含 DSN，则 dsn 可为空字符串
func New(opts ...Option) (database.DB, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	var dialector gorm.Dialector
	if o.postgresConfig != nil {
		dialector = postgres.New(*o.postgresConfig)
	} else {
		if o.dsn == "" {
			return nil, fmt.Errorf("pgsqldb: dsn is required")
		}
		dialector = postgres.Open(o.dsn)
	}

	base, err := gormbase.New(dialector, gormbase.DBTypePostgres, o.gormConfig, &o.Options)
	if err != nil {
		return nil, err
	}
	log.Println("pgsqldb: database connected successfully")
	return base, nil
}

// MustNew 创建由 PostgreSQL 支持的 database.DB，失败时终止程序
func MustNew(opts ...Option) database.DB {
	db, err := New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// NewWithConfig 使用 config.Config 中的配置创建 PostgreSQL 数据库连接
// 配置键（prefix 默认为 "db"）：
//   - {prefix}.dsn (string): PostgreSQL DSN 连接字符串
//   - {prefix}.maxOpenConns (int): 最大打开连接数
//   - {prefix}.maxIdleConns (int): 最大空闲连接数
//   - {prefix}.connMaxLifetime (duration): 连接最大生命周期
//   - {prefix}.connMaxIdleTime (duration): 连接最大空闲时间
//   - {prefix}.ignoreRecordNotFound (bool): 是否忽略 ErrRecordNotFound 错误
//
// 示例配置（YAML）：
//
//	db:
//	  dsn: "host=localhost user=postgres password=pass dbname=mydb port=5432 sslmode=disable"
//	  maxOpenConns: 100
//	  maxIdleConns: 10
//	  connMaxLifetime: 1h
//	  connMaxIdleTime: 10m
//	  ignoreRecordNotFound: true
//
// 多数据库实例示例：
//
//	db_primary:
//	  dsn: "host=primary user=postgres password=pass dbname=db port=5432"
//	db_replica:
//	  dsn: "host=replica user=postgres password=pass dbname=db port=5432"
//
//	primary, _ := pgsqldb.NewWithConfig(cfg, "db_primary")
//	replica, _ := pgsqldb.NewWithConfig(cfg, "db_replica")
func NewWithConfig(cfg config.Config, prefixes ...string) (database.DB, error) {
	prefix := "db"
	if len(prefixes) > 0 && prefixes[0] != "" {
		prefix = prefixes[0]
	}

	opts := []Option{}

	if dsn := cfg.GetString(prefix + ".dsn"); dsn != "" {
		opts = append(opts, WithDSN(dsn))
	}

	if maxOpen := cfg.GetInt(prefix + ".maxOpenConns"); maxOpen > 0 {
		opts = append(opts, WithMaxOpenConns(maxOpen))
	}

	if maxIdle := cfg.GetInt(prefix + ".maxIdleConns"); maxIdle > 0 {
		opts = append(opts, WithMaxIdleConns(maxIdle))
	}

	if lifetime := cfg.GetDuration(prefix + ".connMaxLifetime"); lifetime > 0 {
		opts = append(opts, WithConnMaxLifetime(lifetime))
	}

	if idleTime := cfg.GetDuration(prefix + ".connMaxIdleTime"); idleTime > 0 {
		opts = append(opts, WithConnMaxIdleTime(idleTime))
	}

	if cfg.GetBool(prefix + ".ignoreRecordNotFound") {
		opts = append(opts, WithIgnoreRecordNotFound(true))
	}

	return New(opts...)
}

// MustNewWithConfig 是 NewWithConfig 的 Must 版本，失败时终止程序
func MustNewWithConfig(cfg config.Config, prefixes ...string) database.DB {
	db, err := NewWithConfig(cfg, prefixes...)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
