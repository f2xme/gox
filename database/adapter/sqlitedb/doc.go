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
//	db, err := sqlitedb.New(sqlitedb.WithFile("test.db"))
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
// 配置选项：
//
//	// 使用 Option 函数
//	db, err := sqlitedb.New(
//		sqlitedb.WithFile("test.db"),
//		sqlitedb.WithMaxOpenConns(1),
//		sqlitedb.WithConnMaxIdleTime(5*time.Minute),
//	)
//
//	// 或函数选项直接改 sqlitedb.Options（内嵌连接池字段）
//	db, err := sqlitedb.New(
//		func(o *sqlitedb.Options) {
//			o.file = "test.db"
//			o.MaxOpenConns = 1
//			o.MaxIdleConns = 1
//		},
//	)
//
// 从配置文件加载：
//
//	// config.yaml:
//	// db:
//	//   file: "app.db"
//	//   maxOpenConns: 1
//	//   maxIdleConns: 1
//	//   connMaxLifetime: 1h
//	//   connMaxIdleTime: 10m
//
//	cfg := config.Load("config.yaml")
//	db, err := sqlitedb.NewWithConfig(cfg)
//
//	// 多数据库实例
//	main, _ := sqlitedb.NewWithConfig(cfg, "db_main")
//	cache, _ := sqlitedb.NewWithConfig(cfg, "db_cache")
//
// SQLite 默认使用单连接模式 (MaxOpenConns=1, MaxIdleConns=1)
// 以避免并发写入问题。谨慎修改连接池设置。
package sqlitedb
