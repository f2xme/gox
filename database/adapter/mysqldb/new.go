package mysqldb

import (
	"fmt"
	"log"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/database"
	"github.com/f2xme/gox/database/adapter/internal/gormbase"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// New 创建由 MySQL 支持的 database.DB
// dsn 示例: "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
// 若使用 WithMySQLConfig 且配置中已含 DSN，则 dsn 可为空字符串
func New(opts ...Option) (database.DB, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	var dialector gorm.Dialector
	if o.mysqlConfig != nil {
		dialector = mysql.New(*o.mysqlConfig)
	} else {
		if o.dsn == "" {
			return nil, fmt.Errorf("mysqldb: dsn is required")
		}
		dialector = mysql.Open(o.dsn)
	}

	base, err := gormbase.New(dialector, gormbase.DBTypeMySQL, o.gormConfig, &o.Options)
	if err != nil {
		return nil, err
	}
	return wrapMySQL(base)
}

// MustNew 创建由 MySQL 支持的 database.DB，失败时终止程序
func MustNew(opts ...Option) database.DB {
	db, err := New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// NewWithConfig 使用 config.Config 中的配置创建 MySQL 数据库连接
// 配置键（prefix 默认为 "db"）：
//   - {prefix}.dsn (string): MySQL DSN 连接字符串
//   - {prefix}.maxOpenConns (int): 最大打开连接数
//   - {prefix}.maxIdleConns (int): 最大空闲连接数
//   - {prefix}.connMaxLifetime (duration): 连接最大生命周期
//   - {prefix}.connMaxIdleTime (duration): 连接最大空闲时间
//   - {prefix}.ignoreRecordNotFound (bool): 是否忽略 ErrRecordNotFound 错误
//
// 示例配置（YAML）：
//
//	db:
//	  dsn: "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True"
//	  maxOpenConns: 100
//	  maxIdleConns: 10
//	  connMaxLifetime: 1h
//	  connMaxIdleTime: 10m
//	  ignoreRecordNotFound: true
//
// 多数据库实例示例：
//
//	db_primary:
//	  dsn: "user:pass@tcp(primary:3306)/db"
//	db_replica:
//	  dsn: "user:pass@tcp(replica:3306)/db"
//
//	primary, _ := mysqldb.NewWithConfig(cfg, "db_primary")
//	replica, _ := mysqldb.NewWithConfig(cfg, "db_replica")
func NewWithConfig(cfg config.Config, prefixes ...string) (database.DB, error) {
	prefix := "db"
	if len(prefixes) > 0 && prefixes[0] != "" {
		prefix = prefixes[0]
	}

	opts := make([]Option, 0, 6)
	key := func(suffix string) string { return prefix + "." + suffix }

	if dsn := cfg.GetString(key("dsn")); dsn != "" {
		opts = append(opts, WithDSN(dsn))
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

func wrapMySQL(base database.DB) (database.DB, error) {
	g, ok := base.(*gormbase.GormDB)
	if !ok {
		return nil, fmt.Errorf("mysqldb: unexpected database type %T", base)
	}
	log.Println("mysqldb: database connected successfully")
	return &MySQLDB{GormDB: g}, nil
}
