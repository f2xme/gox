// Package collection 提供泛型集合工具函数，类似 lodash 的功能。
//
// 基于 Go 泛型实现，提供函数式编程风格的切片操作和集合数据结构。
//
// # 功能特性
//
//   - 切片操作：Filter、Map、Reduce、Contains、Unique、Chunk
//   - Set 集合：基于 map 实现的泛型集合类型
//   - 集合运算：Union（并集）、Intersection（交集）、Difference（差集）
//   - 类型安全：完全基于 Go 泛型，编译时类型检查
//
// # 快速开始
//
// 切片操作示例：
//
//	package main
//
//	import (
//		"fmt"
//
//		"github.com/f2xme/gox/collection"
//	)
//
//	func main() {
//		// Filter: 过滤偶数
//		nums := []int{1, 2, 3, 4, 5}
//		evens := collection.Filter(nums, func(n int) bool {
//			return n%2 == 0
//		})
//		fmt.Println(evens) // [2 4]
//
//		// Map: 转换元素
//		doubled := collection.Map(nums, func(n int) int {
//			return n * 2
//		})
//		fmt.Println(doubled) // [2 4 6 8 10]
//
//		// Reduce: 求和
//		sum := collection.Reduce(nums, 0, func(acc, n int) int {
//			return acc + n
//		})
//		fmt.Println(sum) // 15
//
//		// Unique: 去重
//		items := []int{1, 2, 2, 3, 3, 3}
//		unique := collection.Unique(items)
//		fmt.Println(unique) // [1 2 3]
//	}
//
// Set 集合示例：
//
//	package main
//
//	import (
//		"fmt"
//
//		"github.com/f2xme/gox/collection"
//	)
//
//	func main() {
//		// 创建集合
//		s1 := collection.NewSet(1, 2, 3)
//		s2 := collection.NewSet(2, 3, 4)
//
//		// 添加和删除元素
//		s1.Add(4)
//		s1.Remove(1)
//
//		// 检查元素
//		if s1.Contains(2) {
//			fmt.Println("包含 2")
//		}
//
//		// 并集
//		union := s1.Union(s2)
//		fmt.Println(union.ToSlice()) // [2 3 4]
//
//		// 交集
//		intersection := s1.Intersection(s2)
//		fmt.Println(intersection.ToSlice()) // [2 3 4]
//
//		// 差集
//		diff := s1.Difference(s2)
//		fmt.Println(diff.ToSlice()) // []
//	}
//
// # 注意事项
//
//   - Filter、Map、Reduce 等函数不会修改原切片，而是返回新切片
//   - Set 是基于 map 实现的，元素类型必须是 comparable
//   - Set.ToSlice() 返回的切片元素顺序是不确定的（map 的迭代顺序）
//   - Chunk 函数在 size <= 0 时返回 nil
package collection
