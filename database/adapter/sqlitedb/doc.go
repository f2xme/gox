// Package sqlitedb 提供 SQLite 数据库适配器。
//
// # 功能特性
//
//   - 基于 GORM 的 SQLite 支持
//   - 单连接模式（避免并发写入问题）
//   - 事务支持
//   - 自动表结构迁移
//   - 适合开发和测试环境
//
// # 快速开始
//
// 基本使用：
//
//	import "github.com/f2xme/gox/database/adapter/sqlitedb"
//
//	db, err := sqlitedb.New("test.db")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
// 配置选项：
//
//	// 使用 Option 函数
//	db, err := sqlitedb.New("test.db",
//		gormbase.WithLogger(customLogger),
//		gormbase.WithTablePrefix("app_"),
//		gormbase.WithSingularTable(),
//	)
//
//	// 或直接使用 Options 结构体
//	db, err := sqlitedb.New("test.db",
//		func(o *gormbase.Options) {
//			o.Logger = customLogger
//			o.TablePrefix = "app_"
//			o.SingularTable = true
//		},
//	)
//
// SQLite 默认使用单连接模式 (MaxOpenConns=1, MaxIdleConns=1)
// 以避免并发写入问题。谨慎修改连接池设置。
package sqlitedb
