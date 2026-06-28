// Package nullx 提供 database/sql nullable 类型的便捷转换函数。
//
// nullx 主要用于在普通 Go 值、指针值、sql.Null* 类型和数据库参数之间转换。值转换函数
// 采用零值转 NULL 的约定：空字符串、非正整数会转换为无效值；指针转换函数使用 nil 表示
// NULL，非 nil 指针会保留原始值。
//
// # 功能特性
//
//   - 字符串转换：支持 string、*string 和 sql.NullString 互转
//   - 整数转换：支持 int64、*int64 和 sql.NullInt64 互转
//   - 布尔转换：支持 bool 转 sql.NullBool，false 也是有效值
//   - 时间转换：支持 time.Time、*time.Time、sql.NullTime 和时间字符串互转
//   - 扫描辅助：支持将 NULL 字符串和整数扫描为零值
//   - 零值约定：值转换时将空字符串和非正整数视为 NULL；布尔值不使用零值转 NULL
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"fmt"
//
//		"github.com/f2xme/gox/nullx"
//	)
//
//	func main() {
//		name := nullx.String("alice")
//		empty := nullx.String("")
//
//		fmt.Println(name.Valid)  // true
//		fmt.Println(empty.Valid) // false
//	}
//
// # 导入别名
//
// 其他包需要为 nullx 设置导入别名时，默认使用 nx：
//
//	import nx "github.com/f2xme/gox/nullx"
//
//	name := nx.String("alice")
//
// # 指针转换
//
// 指针转换使用 nil 表示 NULL，非 nil 指针会保留原始值：
//
//	var value = ""
//	nullValue := nullx.StringFromPtr(&value)
//	fmt.Println(nullValue.Valid) // true
package nullx
