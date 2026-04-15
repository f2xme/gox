// Package pager 提供 API 和数据库的分页工具。
//
// pager 包支持三种分页策略，适用于不同的场景和需求。
// 所有类型都是线程安全的，可以轻松集成到 HTTP 处理器中。
//
// # 功能特性
//
//   - 偏移分页（Offset-based）：使用 limit 和 offset 参数，适合简单列表查询
//   - 页码分页（Page-based）：使用 page 和 size 参数，适合传统分页 UI
//   - 游标分页（Cursor-based）：使用不透明游标，适合大数据量和实时数据流
//   - HTTP 请求解析：从查询参数自动解析分页参数
//   - 游标编解码：提供 Base64 URL 安全编码的游标工具
//   - 线程安全：所有类型不可变，支持并发使用
//
// # 快速开始
//
// 偏移分页示例：
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/f2xme/gox/pager"
//	)
//
//	func main() {
//		// 创建偏移分页参数
//		page := pager.NewOffset(10, 20) // limit=10, offset=20
//
//		// 模拟查询数据
//		items := []string{"item1", "item2", "item3"}
//		total := int64(100)
//
//		// 创建分页结果
//		result := pager.NewOffsetResult(page, items, total)
//		fmt.Printf("总数: %d, 是否有下一页: %v\n", result.Total, result.HasNext())
//	}
//
// 页码分页示例：
//
//	// 创建页码分页参数
//	page := pager.NewPage(2, 10) // page=2, size=10
//
//	// 转换为 offset 用于数据库查询
//	offset := page.ToOffset()
//	// SELECT * FROM users LIMIT offset.Limit OFFSET offset.Offset
//
//	// 创建分页结果
//	items := []string{"item1", "item2"}
//	result := pager.NewPageResult(page, items, 100)
//	fmt.Printf("当前页: %d/%d\n", result.Page, result.TotalPages)
//
// 游标分页示例：
//
//	// 创建游标分页参数
//	page := pager.NewCursor("", 10) // 空游标表示第一页
//
//	// 查询数据后获取下一页游标
//	items := []string{"item1", "item2"}
//	nextCursor := pager.EncodeCursor("user_id:123")
//
//	// 创建分页结果
//	result := pager.NewCursorResult(page, items, nextCursor)
//	fmt.Printf("是否有下一页: %v\n", result.HasNext())
//
// # 从 HTTP 请求解析分页参数
//
// 偏移分页：
//
//	import "net/http"
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//		// 从查询参数解析：?limit=20&offset=40
//		page := pager.NewOffsetFromRequest(r)
//		// 使用 page.Limit 和 page.Offset 查询数据
//	}
//
// 页码分页：
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//		// 从查询参数解析：?page=2&size=20
//		page := pager.NewPageFromRequest(r)
//		// 使用 page.ToOffset() 转换为数据库查询参数
//	}
//
// 游标分页：
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//		// 从查询参数解析：?cursor=abc123&limit=20
//		page := pager.NewCursorFromRequest(r)
//		// 解码游标并查询数据
//		cursor, _ := pager.DecodeCursor(page.Cursor)
//	}
//
// # 分页策略选择指南
//
// 偏移分页（Offset-based）：
//   - 适用场景：简单列表查询、数据量不大（< 10 万条）
//   - 优点：简单直观，易于实现
//   - 缺点：offset 越大性能越差，数据变化可能导致重复或遗漏
//
// 页码分页（Page-based）：
//   - 适用场景：传统分页 UI、需要跳转到指定页、搜索结果
//   - 优点：用户友好，符合使用习惯
//   - 缺点：本质是 offset 分页，有相同的性能问题
//
// 游标分页（Cursor-based）：
//   - 适用场景：大数据量（> 10 万条）、实时数据流、移动应用无限滚动
//   - 优点：性能稳定，不受数据位置影响，避免数据重复/遗漏
//   - 缺点：无法跳转到指定页，游标需要编解码
//
// # 性能优化建议
//
// 1. 大数据量使用游标分页：
//
//	// 推荐：游标分页（性能稳定）
//	SELECT * FROM users WHERE id > ? ORDER BY id LIMIT 10
//
//	// 不推荐：偏移分页（offset 越大越慢）
//	SELECT * FROM users ORDER BY id LIMIT 10 OFFSET 100000
//
// 2. 确保排序字段有索引：
//
//	CREATE INDEX idx_users_created_at ON users(created_at);
//
// 3. 游标分页使用主键或唯一索引：
//
//	SELECT * FROM users WHERE id > ? ORDER BY id LIMIT 10
//
// 4. 设置合理的默认值和限制：
//
//	// 默认每页 10 条，最大 100 条
//	const (
//		DefaultLimit = 10
//		MaxLimit     = 100
//	)
//
// # 注意事项
//
//   - 所有类型都是不可变的，可以在多个 goroutine 中并发使用
//   - 偏移分页和页码分页在数据频繁变化时可能出现重复或遗漏
//   - 游标分页需要确保排序字段的唯一性和稳定性
//   - 游标应该使用 Base64 编码避免 URL 特殊字符问题
//   - 建议在 API 文档中明确说明分页参数的默认值和限制
package pager
