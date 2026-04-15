/*
Package database 提供统一的数据库连接管理抽象层。

# 功能特性

  - 统一的数据库接口抽象
  - 支持多种数据库（MySQL、PostgreSQL、SQLite）
  - 基于 GORM 的 ORM 支持
  - 事务管理
  - 连接池配置
  - 自动表结构迁移

# 快速开始

基本使用：

	package main

	import (
		"context"
		"log"

		"github.com/f2xme/gox/database"
		"github.com/f2xme/gox/database/adapter/mysqldb"
	)

	func main() {
		// 创建 MySQL 数据库连接
		dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
		db, err := mysqldb.New(dsn)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// 自动迁移表结构
		db.AutoMigrate(&User{})

		// 执行事务
		ctx := context.Background()
		err = db.Transaction(ctx, func(tx database.DB) error {
			// 在事务中执行操作
			return nil
		})
	}

# 核心接口

## DB - 数据库连接接口

所有数据库实现都必须实现此接口：

	type DB interface {
		Engine() any
		Transaction(ctx context.Context, fn func(tx DB) error) error
		AutoMigrate(models ...any) error
		Close() error
	}

使用示例：

	// 获取底层引擎（如 *gorm.DB）
	gormDB := db.Engine().(*gorm.DB)

	// 执行事务
	err := db.Transaction(ctx, func(tx database.DB) error {
		// 在事务中执行操作
		return nil
	})

	// 自动迁移表结构
	db.AutoMigrate(&User{}, &Order{})

	// 关闭连接
	defer db.Close()

# 事务处理

使用 Transaction 方法执行事务操作：

	err := db.Transaction(ctx, func(tx database.DB) error {
		// 操作 1
		if err := createUser(tx); err != nil {
			return err // 自动回滚
		}

		// 操作 2
		if err := createOrder(tx); err != nil {
			return err // 自动回滚
		}

		return nil // 自动提交
	})

# 最佳实践

## 1. 使用 context 控制超时

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.Transaction(ctx, func(tx database.DB) error {
		// 数据库操作
		return nil
	})

## 2. 优雅关闭连接

	db := adapter.New(...)
	defer db.Close()

## 3. 使用事务保证数据一致性

	// 推荐：使用事务
	db.Transaction(ctx, func(tx database.DB) error {
		createUser(tx)
		createProfile(tx)
		return nil
	})

	// 不推荐：分开执行
	createUser(db)
	createProfile(db) // 如果失败，用户已创建

# 线程安全

所有数据库实现都应该是线程安全的，可以在多个 goroutine 中并发使用。
*/
package database
