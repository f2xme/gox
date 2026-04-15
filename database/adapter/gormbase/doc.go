// Package gormbase 提供通用的 GORM 数据库实现。
//
// # 功能特性
//
//   - 统一的 GORM 配置管理
//   - 连接池配置（最大连接数、空闲连接数、连接生命周期）
//   - 日志记录器配置
//   - 表名策略（单数表名、表名前缀）
//   - DryRun 模式支持
//   - 为 MySQL、PostgreSQL 和 SQLite 提供共享基础
//
// # 快速开始
//
// 本包通常通过特定的数据库适配器使用：
//
//	import (
//		"github.com/f2xme/gox/database/adapter/mysqldb"
//		"github.com/f2xme/gox/database/adapter/gormbase"
//	)
//
//	db, err := mysqldb.New(dsn,
//		gormbase.WithMaxOpenConns(200),
//		gormbase.WithMaxIdleConns(20),
//	)
//
// 配置选项：
//
//	// 使用 Option 函数
//	db, err := mysqldb.New(dsn,
//		gormbase.WithMaxOpenConns(200),
//		gormbase.WithMaxIdleConns(20),
//		gormbase.WithConnMaxLifetime(time.Hour),
//		gormbase.WithTablePrefix("app_"),
//	)
//
//	// 或直接使用 Options 结构体
//	db, err := mysqldb.New(dsn,
//		func(o *gormbase.Options) {
//			o.MaxOpenConns = 200
//			o.MaxIdleConns = 20
//			o.ConnMaxLifetime = time.Hour
//			o.TablePrefix = "app_"
//		},
//	)
//
// 默认配置：
//
//	DefaultConfig()       // MaxOpenConns: 100, MaxIdleConns: 10
//	DefaultSQLiteConfig() // MaxOpenConns: 1, MaxIdleConns: 1 (单连接)
package gormbase
