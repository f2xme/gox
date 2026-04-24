package mysqldb_test

import (
	"fmt"
	"time"

	"github.com/f2xme/gox/database/adapter/mysqldb"
)

// 演示基本用法
func ExampleNew() {
	dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True"
	db, err := mysqldb.New(
		mysqldb.WithDSN(dsn),
		mysqldb.WithMaxOpenConns(100),
		mysqldb.WithMaxIdleConns(10),
	)
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer db.Close()

	fmt.Println("数据库连接成功")
}

// 演示使用函数选项
func ExampleNew_withFunctionalOptions() {
	dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True"
	db, err := mysqldb.New(
		func(o *mysqldb.Options) {
			o.MaxOpenConns = 200
			o.MaxIdleConns = 20
			o.ConnMaxLifetime = 2 * time.Hour
			o.ConnMaxIdleTime = 10 * time.Minute
		},
		mysqldb.WithDSN(dsn),
	)
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer db.Close()

	fmt.Println("数据库连接成功")
}

// 演示从配置文件加载（伪代码）
func ExampleNewWithConfig() {
	// 假设配置文件内容：
	// db:
	//   dsn: "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True"
	//   maxOpenConns: 100
	//   maxIdleConns: 10
	//   connMaxLifetime: 1h
	//   connMaxIdleTime: 10m

	// cfg := config.Load("config.yaml")
	// db, err := mysqldb.NewWithConfig(cfg)
	// if err != nil {
	//     log.Fatal(err)
	// }
	// defer db.Close()

	fmt.Println("从配置文件加载数据库连接")
}

// 演示多数据库实例
func ExampleNewWithConfig_multipleInstances() {
	// 假设配置文件内容：
	// db_primary:
	//   dsn: "user:pass@tcp(primary:3306)/db"
	//   maxOpenConns: 100
	// db_replica:
	//   dsn: "user:pass@tcp(replica:3306)/db"
	//   maxOpenConns: 50

	// cfg := config.Load("config.yaml")
	// primary, _ := mysqldb.NewWithConfig(cfg, "db_primary")
	// replica, _ := mysqldb.NewWithConfig(cfg, "db_replica")
	// defer primary.Close()
	// defer replica.Close()

	fmt.Println("多数据库实例配置")
}
