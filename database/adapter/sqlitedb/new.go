package sqlitedb

import (
	"fmt"
	"log"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/database"
	"github.com/f2xme/gox/database/adapter/internal/gormbase"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// New 创建由 SQLite 支持的 database.DB
// file 示例: "test.db" 或 ":memory:" (内存数据库)
// 若使用 WithSQLiteConfig 且配置中已含 DSN，则 file 可为空字符串
func New(opts ...Option) (database.DB, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	var dialector gorm.Dialector
	if o.sqliteConfig != nil {
		dialector = sqlite.Open(o.sqliteConfig.DSN)
	} else {
		if o.file == "" {
			return nil, fmt.Errorf("sqlitedb: file path is required")
		}
		dialector = sqlite.Open(o.file)
	}

	base, err := gormbase.New(dialector, gormbase.DBTypeSqlite, o.gormConfig, &o.Options)
	if err != nil {
		return nil, err
	}
	return wrapSQLite(base)
}

// MustNew 创建由 SQLite 支持的 database.DB，失败时终止程序
func MustNew(opts ...Option) database.DB {
	db, err := New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// NewWithConfig 使用 config.Config 中的配置创建 SQLite 数据库连接
// 配置键（prefix 默认为 "db"）：
//   - {prefix}.file (string): SQLite 数据库文件路径
//   - {prefix}.maxOpenConns (int): 最大打开连接数（SQLite 建议为 1）
//   - {prefix}.maxIdleConns (int): 最大空闲连接数（SQLite 建议为 1）
//   - {prefix}.connMaxLifetime (duration): 连接最大生命周期
//   - {prefix}.connMaxIdleTime (duration): 连接最大空闲时间
//   - {prefix}.ignoreRecordNotFound (bool): 是否忽略 ErrRecordNotFound 错误
//
// 示例配置（YAML）：
//
//	db:
//	  file: "app.db"
//	  maxOpenConns: 1
//	  maxIdleConns: 1
//	  connMaxLifetime: 1h
//	  connMaxIdleTime: 10m
//	  ignoreRecordNotFound: true
//
// 多数据库实例示例：
//
//	db_main:
//	  file: "main.db"
//	db_cache:
//	  file: "cache.db"
//
//	main, _ := sqlitedb.NewWithConfig(cfg, "db_main")
//	cache, _ := sqlitedb.NewWithConfig(cfg, "db_cache")
func NewWithConfig(cfg config.Config, prefixes ...string) (database.DB, error) {
	prefix := "db"
	if len(prefixes) > 0 && prefixes[0] != "" {
		prefix = prefixes[0]
	}

	opts := make([]Option, 0, 6)
	key := func(suffix string) string { return prefix + "." + suffix }

	if file := cfg.GetString(key("file")); file != "" {
		opts = append(opts, WithFile(file))
	}

	if maxOpen := cfg.GetInt(key("maxOpenConns")); maxOpen > 0 {
		opts = append(opts, WithMaxOpenConns(maxOpen))
	}

	if maxIdle := cfg.GetInt(key("maxIdleConns")); maxIdle > 0 {
		opts = append(opts, WithMaxIdleConns(maxIdle))
	}

	if lifetime := cfg.GetDuration(key("connMaxLifetime")); lifetime > 0 {
		opts = append(opts, WithConnMaxLifetime(lifetime))
	}

	if idleTime := cfg.GetDuration(key("connMaxIdleTime")); idleTime > 0 {
		opts = append(opts, WithConnMaxIdleTime(idleTime))
	}

	if cfg.GetBool(key("ignoreRecordNotFound")) {
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

func wrapSQLite(base database.DB) (database.DB, error) {
	g, ok := base.(*gormbase.GormDB)
	if !ok {
		return nil, fmt.Errorf("sqlitedb: unexpected database type %T", base)
	}
	log.Println("sqlitedb: database connected successfully")
	return &SQLiteDB{GormDB: g}, nil
}
