// Package pgsqldb 提供 PostgreSQL 数据库适配器。
//
// # 功能特性
//
//   - 基于 GORM 的 PostgreSQL 支持
//   - 连接池管理
//   - 事务支持
//   - 自动表结构迁移
//   - 灵活的配置选项
//
// # 快速开始
//
// 基本使用：
//
//	import "github.com/f2xme/gox/database/adapter/pgsqldb"
//
//	dsn := "host=localhost user=postgres password=pass dbname=mydb port=5432 sslmode=disable"
//	db, err := pgsqldb.New(pgsqldb.WithDSN(dsn))
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
// 配置选项：
//
//	// 使用 Option 函数
//	db, err := pgsqldb.New(
//		pgsqldb.WithDSN(dsn),
//		pgsqldb.WithMaxOpenConns(200),
//		pgsqldb.WithMaxIdleConns(20),
//		pgsqldb.WithConnMaxLifetime(2*time.Hour),
//		pgsqldb.WithConnMaxIdleTime(10*time.Minute),
//	)
//
//	// 或函数选项直接改 pgsqldb.Options（内嵌连接池字段）
//	db, err := pgsqldb.New(
//		func(o *pgsqldb.Options) {
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
//	//   dsn: "host=localhost user=postgres password=pass dbname=mydb port=5432 sslmode=disable"
//	//   maxOpenConns: 100
//	//   maxIdleConns: 10
//	//   connMaxLifetime: 1h
//	//   connMaxIdleTime: 10m
//
//	cfg := config.Load("config.yaml")
//	db, err := pgsqldb.NewWithConfig(cfg)
//
//	// 多数据库实例
//	primary, _ := pgsqldb.NewWithConfig(cfg, "db_primary")
//	replica, _ := pgsqldb.NewWithConfig(cfg, "db_replica")
package pgsqldb
