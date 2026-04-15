// Package mysqldb 提供 MySQL 数据库适配器。
//
// # 功能特性
//
//   - 基于 GORM 的 MySQL 支持
//   - 连接池管理
//   - 事务支持
//   - 自动表结构迁移
//   - 灵活的配置选项
//
// # 快速开始
//
// 基本使用：
//
//	import "github.com/f2xme/gox/database/adapter/mysqldb"
//
//	dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
//	db, err := mysqldb.New(dsn)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
// 配置选项：
//
//	// 使用 Option 函数
//	db, err := mysqldb.New(dsn,
//		gormbase.WithMaxOpenConns(200),
//		gormbase.WithMaxIdleConns(20),
//		gormbase.WithConnMaxLifetime(2*time.Hour),
//		gormbase.WithTablePrefix("app_"),
//	)
//
//	// 或直接使用 Options 结构体
//	db, err := mysqldb.New(dsn,
//		func(o *gormbase.Options) {
//			o.MaxOpenConns = 200
//			o.MaxIdleConns = 20
//			o.ConnMaxLifetime = 2 * time.Hour
//			o.TablePrefix = "app_"
//		},
//	)
package mysqldb
