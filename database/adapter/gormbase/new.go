package gormbase

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/database"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// New 使用给定的 dialector 和选项创建新的 GORM 数据库
func New(dialector gorm.Dialector, dbType string, opts *Options) (database.DB, error) {
	gormCfg := &gorm.Config{DryRun: opts.DryRun}
	if opts.SingularTable || opts.TablePrefix != "" {
		gormCfg.NamingStrategy = schema.NamingStrategy{
			SingularTable: opts.SingularTable,
			TablePrefix:   opts.TablePrefix,
		}
	}
	if opts.Logger != nil {
		gormCfg.Logger = opts.Logger
	}
	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", dbType, err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	applyConnectionPoolConfig(sqlDB, opts)
	return &GormDB{db: db, sqlDB: sqlDB}, nil
}

// MustNew 创建新的 GORM 数据库，失败时终止程序
func MustNew(dialector gorm.Dialector, dbType string, opts *Options) database.DB {
	db, err := New(dialector, dbType, opts)
	if err != nil {
		log.Fatalf("failed to create database: %v", err)
	}
	return db
}

// NewWithConfig 使用 config.Config 中的配置创建 GORM 数据库
// 配置键：
//   - db.maxOpenConns (int): 最大打开连接数
//   - db.maxIdleConns (int): 最大空闲连接数
//   - db.connMaxLifetime (duration): 连接最大生命周期
//   - db.connMaxIdleTime (duration): 连接最大空闲时间
//   - db.dryRun (bool): 启用 DryRun 模式
//   - db.singularTable (bool): 使用单数表名
//   - db.tablePrefix (string): 表名前缀
//
// 注意：此函数需要单独传递 dialector，因为它无法从配置文件中配置
func NewWithConfig(dialector gorm.Dialector, dbType string, cfg config.Config) (database.DB, error) {
	opts := defaultOptions()

	if maxOpen := cfg.GetInt("db.maxOpenConns"); maxOpen > 0 {
		opts.MaxOpenConns = maxOpen
	}

	if maxIdle := cfg.GetInt("db.maxIdleConns"); maxIdle > 0 {
		opts.MaxIdleConns = maxIdle
	}

	if lifetime := cfg.GetDuration("db.connMaxLifetime"); lifetime > 0 {
		opts.ConnMaxLifetime = lifetime
	}

	if idleTime := cfg.GetDuration("db.connMaxIdleTime"); idleTime > 0 {
		opts.ConnMaxIdleTime = idleTime
	}

	if cfg.GetBool("db.dryRun") {
		opts.DryRun = true
	}

	if cfg.GetBool("db.singularTable") {
		opts.SingularTable = true
	}

	if prefix := cfg.GetString("db.tablePrefix"); prefix != "" {
		opts.TablePrefix = prefix
	}

	return New(dialector, dbType, &opts)
}

// MustNewWithConfig 使用配置创建 GORM 数据库，失败时终止程序
func MustNewWithConfig(dialector gorm.Dialector, dbType string, cfg config.Config) database.DB {
	db, err := NewWithConfig(dialector, dbType, cfg)
	if err != nil {
		log.Fatalf("gormbase: failed to create database from config: %v", err)
	}
	return db
}

func applyConnectionPoolConfig(sqlDB *sql.DB, opts *Options) {
	sqlDB.SetMaxOpenConns(opts.MaxOpenConns)
	sqlDB.SetMaxIdleConns(opts.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(opts.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(opts.ConnMaxIdleTime)
}
