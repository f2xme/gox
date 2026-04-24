package gormbase

import (
	"database/sql"
	"fmt"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/database"
	"gorm.io/gorm"
)

// New 使用给定的 dialector、gorm.Config 与连接池选项创建新的 GORM 数据库
func New(dialector gorm.Dialector, dbType DBType, gormCfg *gorm.Config, opts *Options) (database.DB, error) {
	pool := opts
	if pool == nil {
		o := defaultOptions()
		pool = &o
	}
	if err := pool.Validate(); err != nil {
		return nil, err
	}

	cfg := gormCfg
	if cfg == nil {
		cfg = &gorm.Config{}
	}
	db, err := gorm.Open(dialector, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", dbType, err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	applyConnectionPoolConfig(sqlDB, pool)

	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 注册回调忽略 ErrRecordNotFound 错误
	if pool.IgnoreRecordNotFound {
		if err = db.Callback().Query().After("gorm:after_query").Register("ignore_record_not_found", func(db *gorm.DB) {
			if db.Error == gorm.ErrRecordNotFound {
				db.Error = nil
			}
		}); err != nil {
			return nil, fmt.Errorf("db callback register error: %w", err)
		}
	}

	return &GormDB{db: db, sqlDB: sqlDB}, nil
}

// NewWithConfig 使用 config.Config 中的连接池配置创建 GORM 数据库
// 配置键：
//   - db.maxOpenConns (int): 最大打开连接数
//   - db.maxIdleConns (int): 最大空闲连接数
//   - db.connMaxLifetime (duration): 连接最大生命周期
//   - db.connMaxIdleTime (duration): 连接最大空闲时间
//
// 注意：此函数需要单独传递 dialector，因为它无法从配置文件中配置
func NewWithConfig(dialector gorm.Dialector, dbType DBType, gormCfg *gorm.Config, cfg config.Config) (database.DB, error) {
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

	return New(dialector, dbType, gormCfg, &opts)
}

func applyConnectionPoolConfig(sqlDB *sql.DB, opts *Options) {
	sqlDB.SetMaxOpenConns(opts.MaxOpenConns)
	sqlDB.SetMaxIdleConns(opts.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(opts.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(opts.ConnMaxIdleTime)
}
