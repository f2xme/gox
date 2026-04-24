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
//	db, err := mysqldb.New(
//		mysqldb.WithDSN(dsn),
//		mysqldb.WithMaxOpenConns(200),
//		mysqldb.WithMaxIdleConns(20),
//		mysqldb.WithConnMaxLifetime(2*time.Hour),
//		mysqldb.WithConnMaxIdleTime(10*time.Minute),
//	)
//
//	// 或函数选项直接改 mysqldb.Options（内嵌连接池字段）
//	db, err := mysqldb.New(
//		func(o *mysqldb.Options) {
//			o.dsn = dsn
//			o.MaxOpenConns = 200
//			o.MaxIdleConns = 20
//			o.ConnMaxLifetime = 2 * time.Hour
//			o.ConnMaxIdleTime = 10 * time.Minute
//		},
//	)
//
// 从配置文件加载：
//
//	// config.yaml:
//	// db:
//	//   dsn: "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True"
//	//   maxOpenConns: 100
//	//   maxIdleConns: 10
//	//   connMaxLifetime: 1h
//	//   connMaxIdleTime: 10m
//
//	cfg := config.Load("config.yaml")
//	db, err := mysqldb.NewWithConfig(cfg)
//
//	// 多数据库实例
//	primary, _ := mysqldb.NewWithConfig(cfg, "db_primary")
//	replica, _ := mysqldb.NewWithConfig(cfg, "db_replica")
package mysqldb
