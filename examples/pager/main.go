package main

import (
	"fmt"

	"github.com/f2xme/gox/pager"
)

type User struct {
	ID   int
	Name string
}

func main() {
	fmt.Println("=== pager 包使用示例 ===")

	// 模拟用户数据
	users := []User{
		{ID: 1, Name: "张三"},
		{ID: 2, Name: "李四"},
		{ID: 3, Name: "王五"},
		{ID: 4, Name: "赵六"},
		{ID: 5, Name: "钱七"},
		{ID: 6, Name: "孙八"},
		{ID: 7, Name: "周九"},
		{ID: 8, Name: "吴十"},
		{ID: 9, Name: "郑十一"},
		{ID: 10, Name: "王十二"},
		{ID: 11, Name: "李十三"},
		{ID: 12, Name: "赵十四"},
	}

	// 示例 1: Offset 分页
	fmt.Println("\n示例 1: Offset 分页")
	offsetPage := pager.NewOffset(5, 0)
	fmt.Printf("当前页参数: limit=%d, offset=%d\n", offsetPage.Limit, offsetPage.Offset)

	// 模拟查询结果（前 5 条）
	offsetItems := users[offsetPage.Offset : offsetPage.Offset+offsetPage.Limit]
	offsetResult := pager.NewOffsetResult(offsetPage, offsetItems, int64(len(users)))

	fmt.Printf("总数: %d\n", offsetResult.Total)
	fmt.Printf("本页数据: %v\n", offsetResult.Items)
	fmt.Printf("是否有下一页: %v\n", offsetResult.HasNext())
	fmt.Printf("是否有上一页: %v\n", offsetResult.HasPrev())

	nextOffset := offsetPage.Next()
	fmt.Printf("下一页参数: limit=%d, offset=%d\n", nextOffset.Limit, nextOffset.Offset)

	// 示例 2: Page 分页
	fmt.Println("\n示例 2: Page 分页")
	pageNumber := pager.NewPage(2, 5)
	fmt.Printf("当前页参数: page=%d, size=%d\n", pageNumber.Page, pageNumber.Size)

	pageOffset := pageNumber.ToOffset()
	fmt.Printf("转换为 offset: limit=%d, offset=%d\n", pageOffset.Limit, pageOffset.Offset)

	// 模拟查询结果（第 2 页，第 6-10 条）
	pageItems := users[pageOffset.Offset : pageOffset.Offset+pageOffset.Limit]
	pageResult := pager.NewPageResult(pageNumber, pageItems, int64(len(users)))

	fmt.Printf("总数: %d\n", pageResult.Total)
	fmt.Printf("总页数: %d\n", pageResult.TotalPages)
	fmt.Printf("当前页: %d\n", pageResult.Page)
	fmt.Printf("本页数据: %v\n", pageResult.Items)
	fmt.Printf("是否有下一页: %v\n", pageResult.HasNext())
	fmt.Printf("是否有上一页: %v\n", pageResult.HasPrev())

	nextPage := pageNumber.Next()
	prevPage := pageNumber.Prev()
	fmt.Printf("下一页: page=%d\n", nextPage.Page)
	fmt.Printf("上一页: page=%d\n", prevPage.Page)

	// 示例 3: Cursor 分页
	fmt.Println("\n示例 3: Cursor 分页")
	cursorValue := "user_id:5"
	encodedCursor := pager.EncodeCursor(cursorValue)
	fmt.Printf("原始游标: %s\n", cursorValue)
	fmt.Printf("编码后游标: %s\n", encodedCursor)

	decodedCursor, err := pager.DecodeCursor(encodedCursor)
	if err != nil {
		fmt.Printf("解码失败: %v\n", err)
	} else {
		fmt.Printf("解码后游标: %s\n", decodedCursor)
	}

	cursorPage := pager.NewCursor(encodedCursor, 3)
	fmt.Printf("Cursor 分页参数: cursor=%s, limit=%d\n", cursorPage.Cursor, cursorPage.Limit)

	// 模拟 cursor 查询结果
	cursorItems := users[5:8] // 从 ID 5 之后的 3 条
	nextCursorValue := pager.EncodeCursor("user_id:8")
	cursorResult := pager.NewCursorResult(cursorPage, cursorItems, nextCursorValue)

	fmt.Printf("本页数据: %v\n", cursorResult.Items)
	fmt.Printf("下一页游标: %s\n", cursorResult.NextCursor)
	fmt.Printf("是否有下一页: %v\n", cursorResult.HasNext())

	nextCursorPage := cursorResult.Next()
	fmt.Printf("下一页参数: cursor=%s, limit=%d\n", nextCursorPage.Cursor, nextCursorPage.Limit)

	// 示例 4: 默认值处理
	fmt.Println("\n示例 4: 默认值处理")
	invalidOffset := pager.NewOffset(-1, -10)
	fmt.Printf("无效 offset 参数修正后: limit=%d, offset=%d\n", invalidOffset.Limit, invalidOffset.Offset)

	invalidPage := pager.NewPage(0, 0)
	fmt.Printf("无效 page 参数修正后: page=%d, size=%d\n", invalidPage.Page, invalidPage.Size)

	invalidCursor := pager.NewCursor("", 0)
	fmt.Printf("无效 cursor 参数修正后: cursor=%q, limit=%d\n", invalidCursor.Cursor, invalidCursor.Limit)

	// 示例 5: 计算总页数
	fmt.Println("\n示例 5: 计算总页数")
	totalPages1 := pager.CalculateTotalPages(100, 10)
	totalPages2 := pager.CalculateTotalPages(103, 10)
	fmt.Printf("100 条数据，每页 10 条，总页数: %d\n", totalPages1)
	fmt.Printf("103 条数据，每页 10 条，总页数: %d\n", totalPages2)

	fmt.Println("\n=== 示例结束 ===")
}
