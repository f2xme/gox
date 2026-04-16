/*
Package database 提供统一的数据库连接管理抽象层。

# 功能特性

  - 统一的数据库接口抽象
  - 支持多种数据库（MySQL、PostgreSQL、SQLite）
  - 基于 GORM 的 ORM 支持
  - 完整的 CRUD 操作
  - 链式查询构建
  - 事务管理
  - 连接池配置
  - 自动表结构迁移

# 快速开始

基本使用：

	package main

	import (
		"context"
		"log"

		"github.com/f2xme/gox/database/adapter/mysqldb"
	)

	type User struct {
		ID   uint   `gorm:"primaryKey"`
		Name string
		Age  int
	}

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

		ctx := context.Background()

		// 创建记录
		user := &User{Name: "Alice", Age: 25}
		if err := db.Create(ctx, user); err != nil {
			log.Fatal(err)
		}

		// 查询记录
		var result User
		if err := db.First(ctx, &result, "name = ?", "Alice"); err != nil {
			log.Fatal(err)
		}

		// 更新记录
		if err := db.Update(ctx, &result, "age", 26); err != nil {
			log.Fatal(err)
		}

		// 删除记录
		if err := db.Delete(ctx, &result); err != nil {
			log.Fatal(err)
		}
	}

# 核心接口

## DB - 数据库连接接口

所有数据库实现都必须实现此接口，提供完整的 CRUD 和查询功能。

## 基础 CRUD 操作

	// 创建记录
	db.Create(ctx, &user)

	// 查询第一条记录
	db.First(ctx, &user, "id = ?", 1)

	// 查询所有记录
	var users []User
	db.Find(ctx, &users)

	// 保存（插入或更新）
	db.Save(ctx, &user)

	// 更新单个字段
	db.Update(ctx, &user, "name", "Bob")

	// 更新多个字段
	db.Updates(ctx, &user, map[string]any{"name": "Bob", "age": 30})

	// 删除记录
	db.Delete(ctx, &user)

## 链式查询

	// 条件查询
	var users []User
	db.Where("age > ?", 18).
		Order("created_at DESC").
		Limit(10).
		Offset(0).
		Find(ctx, &users)

	// 统计记录数
	var count int64
	db.Model(&User{}).Where("age > ?", 18).Count(ctx, &count)

	// 选择字段
	db.Select("name, age").Find(ctx, &users)

	// 预加载关联
	db.Preload("Orders").Find(ctx, &users)

## 原生 SQL

	// 执行原生 SQL
	db.Exec(ctx, "UPDATE users SET age = age + 1 WHERE id = ?", 1)

	// 原生查询
	var result []map[string]any
	db.Raw("SELECT * FROM users WHERE age > ?", 18).Scan(ctx, &result)

## 事务处理

### 自动事务（推荐）

	err := db.Transaction(ctx, func(tx database.DB) error {
		// 操作 1
		if err := tx.Create(ctx, &user); err != nil {
			return err // 自动回滚
		}

		// 操作 2
		if err := tx.Create(ctx, &order); err != nil {
			return err // 自动回滚
		}

		return nil // 自动提交
	})

### 手动事务

	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if err := tx.Create(ctx, &user); err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

## 底层访问

如果需要使用 GORM 的高级功能，可以通过 Unwrap() 获取底层 *gorm.DB：

	gormDB := db.Unwrap().(*gorm.DB)
	// 使用 GORM 的高级功能

注意：不推荐频繁使用 Unwrap()，应优先使用接口提供的方法。

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
		tx.Create(ctx, &user)
		tx.Create(ctx, &profile)
		return nil
	})

	// 不推荐：分开执行
	db.Create(ctx, &user)
	db.Create(ctx, &profile) // 如果失败，用户已创建

## 4. 链式查询提高可读性

	// 推荐
	db.Where("status = ?", "active").
		Order("created_at DESC").
		Limit(10).
		Find(ctx, &users)

	// 不推荐
	db.Find(ctx, &users, "status = ? ORDER BY created_at DESC LIMIT 10", "active")

# 线程安全

所有数据库实现都是线程安全的，可以在多个 goroutine 中并发使用。
*/
package database
