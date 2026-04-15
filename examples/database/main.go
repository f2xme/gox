package main

import (
	"context"
	"fmt"
	"os"

	"github.com/f2xme/gox/database"
	"github.com/f2xme/gox/database/adapter/sqlitedb"
)

// User 用户模型
type User struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"size:100;not null"`
	Email string `gorm:"size:100;uniqueIndex"`
	Age   int
}

func main() {
	fmt.Println("=== Database 使用示例 ===")

	// 1. 创建 SQLite 数据库连接（使用临时文件）
	dbFile := "/tmp/gox-example.db"
	defer os.Remove(dbFile)

	db, err := sqlitedb.New(dbFile)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("1. 数据库连接成功")

	// 2. 自动迁移（创建表）
	fmt.Println("\n2. 创建数据表:")
	if err := db.AutoMigrate(&User{}); err != nil {
		fmt.Printf("迁移失败: %v\n", err)
		return
	}
	fmt.Println("用户表创建成功")

	// 3. 插入数据（Create）
	fmt.Println("\n3. 插入数据:")
	ctx := context.Background()

	err = db.Transaction(ctx, func(tx database.DB) error {
		engine := tx.Engine()
		gormDB := engine.(interface {
			Create(value interface{}) interface{ Error() error }
		})

		users := []User{
			{Name: "张三", Email: "zhangsan@example.com", Age: 25},
			{Name: "李四", Email: "lisi@example.com", Age: 30},
			{Name: "王五", Email: "wangwu@example.com", Age: 28},
		}

		for _, user := range users {
			if result := gormDB.Create(&user); result.Error() != nil {
				return result.Error()
			}
			fmt.Printf("  插入用户: %s (ID: %d)\n", user.Name, user.ID)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("插入数据失败: %v\n", err)
		return
	}

	// 4. 查询数据（Read）
	fmt.Println("\n4. 查询数据:")
	engine := db.Engine()
	gormDB := engine.(interface {
		Find(dest interface{}, conds ...interface{}) interface{ Error() error }
		First(dest interface{}, conds ...interface{}) interface{ Error() error }
		Where(query interface{}, args ...interface{}) interface {
			Find(dest interface{}, conds ...interface{}) interface{ Error() error }
		}
	})

	// 查询所有用户
	var allUsers []User
	if result := gormDB.Find(&allUsers); result.Error() != nil {
		fmt.Printf("查询失败: %v\n", result.Error())
		return
	}
	fmt.Printf("查询到 %d 个用户:\n", len(allUsers))
	for _, u := range allUsers {
		fmt.Printf("  ID: %d, 姓名: %s, 邮箱: %s, 年龄: %d\n", u.ID, u.Name, u.Email, u.Age)
	}

	// 5. 条件查询
	fmt.Println("\n5. 条件查询:")
	var user User
	if result := gormDB.Where("name = ?", "张三").Find(&user); result.Error() != nil {
		fmt.Printf("查询失败: %v\n", result.Error())
		return
	}
	fmt.Printf("查询到用户: %s, 年龄: %d\n", user.Name, user.Age)

	// 6. 更新数据（Update）
	fmt.Println("\n6. 更新数据:")
	err = db.Transaction(ctx, func(tx database.DB) error {
		engine := tx.Engine()
		gormDB := engine.(interface {
			Model(value interface{}) interface {
				Where(query interface{}, args ...interface{}) interface {
					Update(column string, value interface{}) interface{ Error() error }
				}
			}
		})

		result := gormDB.Model(&User{}).Where("name = ?", "张三").Update("age", 26)
		if result.Error() != nil {
			return result.Error()
		}
		fmt.Println("  更新张三的年龄为 26")
		return nil
	})

	if err != nil {
		fmt.Printf("更新失败: %v\n", err)
		return
	}

	// 验证更新
	gormDB.Where("name = ?", "张三").Find(&user)
	fmt.Printf("  验证: 张三的年龄现在是 %d\n", user.Age)

	// 7. 删除数据（Delete）
	fmt.Println("\n7. 删除数据:")
	err = db.Transaction(ctx, func(tx database.DB) error {
		engine := tx.Engine()
		gormDB := engine.(interface {
			Where(query interface{}, args ...interface{}) interface {
				Delete(value interface{}, conds ...interface{}) interface{ Error() error }
			}
		})

		result := gormDB.Where("name = ?", "王五").Delete(&User{})
		if result.Error() != nil {
			return result.Error()
		}
		fmt.Println("  删除用户: 王五")
		return nil
	})

	if err != nil {
		fmt.Printf("删除失败: %v\n", err)
		return
	}

	// 8. 事务示例
	fmt.Println("\n8. 事务示例:")
	err = db.Transaction(ctx, func(tx database.DB) error {
		engine := tx.Engine()
		gormDB := engine.(interface {
			Create(value interface{}) interface{ Error() error }
		})

		// 在事务中插入新用户
		newUser := User{Name: "赵六", Email: "zhaoliu@example.com", Age: 35}
		if result := gormDB.Create(&newUser); result.Error() != nil {
			return result.Error()
		}
		fmt.Printf("  事务中插入用户: %s\n", newUser.Name)

		// 事务会自动提交（如果没有返回错误）
		return nil
	})

	if err != nil {
		fmt.Printf("事务失败: %v\n", err)
		return
	}
	fmt.Println("  事务提交成功")

	// 9. 最终统计
	fmt.Println("\n9. 最终统计:")
	var finalUsers []User
	gormDB.Find(&finalUsers)
	fmt.Printf("数据库中共有 %d 个用户\n", len(finalUsers))

	fmt.Println("\n数据库示例完成")
}
