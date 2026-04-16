package database_test

import (
	"context"
	"fmt"
	"log"

	"github.com/f2xme/gox/database"
	"github.com/f2xme/gox/database/adapter/sqlitedb"
)

type User struct {
	ID   uint   `gorm:"primaryKey"`
	Name string
	Age  int
}

// Example_basicCRUD 演示基础 CRUD 操作
func Example_basicCRUD() {
	db, err := sqlitedb.New(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 自动迁移
	db.AutoMigrate(&User{})

	ctx := context.Background()

	// 创建记录
	user := &User{Name: "Alice", Age: 25}
	if err := db.Create(ctx, user); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created user ID:", user.ID)

	// 查询记录
	var result User
	if err := db.First(ctx, &result, "name = ?", "Alice"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Found user:", result.Name, ", age:", result.Age)

	// 更新记录
	if err := db.Update(ctx, &result, "age", 26); err != nil {
		log.Fatal(err)
	}

	// 删除记录
	if err := db.Delete(ctx, &result); err != nil {
		log.Fatal(err)
	}

	// Output:
	// Created user ID: 1
	// Found user: Alice , age: 25
}

// Example_chainQuery 演示链式查询
func Example_chainQuery() {
	db, err := sqlitedb.New(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.AutoMigrate(&User{})
	ctx := context.Background()

	// 批量创建
	users := []User{
		{Name: "Alice", Age: 25},
		{Name: "Bob", Age: 30},
		{Name: "Charlie", Age: 20},
	}
	for _, u := range users {
		db.Create(ctx, &u)
	}

	// 链式查询
	var results []User
	db.Where("age > ?", 20).
		Order("age DESC").
		Limit(2).
		Find(ctx, &results)

	for _, u := range results {
		fmt.Println(u.Name, ":", u.Age)
	}

	// 统计
	var count int64
	db.Model(&User{}).Where("age > ?", 20).Count(ctx, &count)
	fmt.Println("Count:", count)

	// Output:
	// Bob : 30
	// Alice : 25
	// Count: 2
}

// Example_transaction 演示事务处理
func Example_transaction() {
	db, err := sqlitedb.New(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.AutoMigrate(&User{})
	ctx := context.Background()

	// 自动事务
	err = db.Transaction(ctx, func(tx database.DB) error {
		user1 := &User{Name: "Alice", Age: 25}
		if err := tx.Create(ctx, user1); err != nil {
			return err
		}

		user2 := &User{Name: "Bob", Age: 30}
		if err := tx.Create(ctx, user2); err != nil {
			return err
		}

		return nil // 自动提交
	})

	if err != nil {
		log.Fatal(err)
	}

	// 验证
	var count int64
	db.Model(&User{}).Count(ctx, &count)
	fmt.Println("Total users:", count)

	// Output:
	// Total users: 2
}

// Example_manualTransaction 演示手动事务控制
func Example_manualTransaction() {
	db, err := sqlitedb.New(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.AutoMigrate(&User{})
	ctx := context.Background()

	// 手动事务
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}

	user := &User{Name: "Alice", Age: 25}
	if err := tx.Create(ctx, user); err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction committed")

	// Output:
	// Transaction committed
}

// Example_rawSQL 演示原生 SQL
func Example_rawSQL() {
	db, err := sqlitedb.New(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.AutoMigrate(&User{})
	ctx := context.Background()

	// 插入数据
	db.Create(ctx, &User{Name: "Alice", Age: 25})
	db.Create(ctx, &User{Name: "Bob", Age: 30})

	// 原生查询
	var results []User
	db.Raw("SELECT * FROM users WHERE age > ?", 20).Scan(ctx, &results)

	for _, u := range results {
		fmt.Println(u.Name, ":", u.Age)
	}

	// 原生执行
	db.Exec(ctx, "UPDATE users SET age = age + 1 WHERE name = ?", "Alice")

	// Output:
	// Alice : 25
	// Bob : 30
}
