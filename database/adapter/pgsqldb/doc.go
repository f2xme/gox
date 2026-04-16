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
//	dsn := "host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable"
//	db, err := pgsqldb.New(dsn)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
// 配置选项：
//
//	// 使用 Option 函数
//	db, err := pgsqldb.New(dsn,
//		gormbase.WithMaxOpenConns(200),
//		gormbase.WithMaxIdleConns(20),
//		gormbase.WithConnMaxLifetime(2*time.Hour),
//		gormbase.WithTablePrefix("app_"),
//	)
//
//	// 或直接使用 Options 结构体
//	db, err := pgsqldb.New(dsn,
//		func(o *gormbase.Options) {
//			o.MaxOpenConns = 200
//			o.MaxIdleConns = 20
//			o.ConnMaxLifetime = 2 * time.Hour
//			o.TablePrefix = "app_"
//		},
//	)
//
//	// 使用 config.Config 接口从配置文件读取
//	import "github.com/f2xme/gox/config/adapter/viper"
//	cfg, err := viper.New("config.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//	db, err := pgsqldb.NewWithConfig(dsn, cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
package pgsqldb
