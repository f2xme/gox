package pager_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/f2xme/gox/pager"
)

func ExampleNewOffset() {
	page := pager.NewOffset(10, 20)
	fmt.Printf("Limit: %d, Offset: %d\n", page.Limit, page.Offset)
	// Output: Limit: 10, Offset: 20
}

func ExampleOffsetPage_Next() {
	page := pager.NewOffset(10, 20)
	next := page.Next()
	fmt.Printf("Next offset: %d\n", next.Offset)
	// Output: Next offset: 30
}

func ExampleNewOffsetResult() {
	page := pager.NewOffset(10, 0)
	items := []string{"item1", "item2", "item3"}
	result := pager.NewOffsetResult(page, items, 100)

	fmt.Printf("Items: %d, Total: %d, HasNext: %v\n",
		len(result.Items), result.Total, result.HasNext())
	// Output: Items: 3, Total: 100, HasNext: true
}

func ExampleNewPage() {
	page := pager.NewPage(2, 20)
	fmt.Printf("Page: %d, Size: %d\n", page.Page, page.Size)
	// Output: Page: 2, Size: 20
}

func ExamplePageNumber_ToOffset() {
	page := pager.NewPage(3, 10)
	offset := page.ToOffset()
	fmt.Printf("Offset: %d, Limit: %d\n", offset.Offset, offset.Limit)
	// Output: Offset: 20, Limit: 10
}

func ExampleNewPageResult() {
	page := pager.NewPage(1, 10)
	items := []string{"item1", "item2"}
	result := pager.NewPageResult(page, items, 25)

	fmt.Printf("Page: %d/%d, HasNext: %v\n",
		result.Page, result.TotalPages, result.HasNext())
	// Output: Page: 1/3, HasNext: true
}

func ExampleNewCursor() {
	page := pager.NewCursor("abc123", 20)
	fmt.Printf("Cursor: %s, Limit: %d\n", page.Cursor, page.Limit)
	// Output: Cursor: abc123, Limit: 20
}

func ExampleEncodeCursor() {
	cursor := "user:123:2024-01-01"
	encoded := pager.EncodeCursor(cursor)
	fmt.Printf("Encoded length: %d\n", len(encoded))
	// Output: Encoded length: 28
}

func ExampleDecodeCursor() {
	encoded := pager.EncodeCursor("user:123")
	decoded, _ := pager.DecodeCursor(encoded)
	fmt.Println(decoded)
	// Output: user:123
}

func ExampleNewCursorResult() {
	page := pager.NewCursor("", 10)
	items := []string{"item1", "item2"}
	result := pager.NewCursorResult(page, items, "next_cursor_value")

	fmt.Printf("Items: %d, HasNext: %v\n", len(result.Items), result.HasNext())
	// Output: Items: 2, HasNext: true
}

func ExampleNewOffsetFromRequest() {
	req := httptest.NewRequest("GET", "http://example.com/api?limit=20&offset=40", nil)
	page := pager.NewOffsetFromRequest(req)
	fmt.Printf("Limit: %d, Offset: %d\n", page.Limit, page.Offset)
	// Output: Limit: 20, Offset: 40
}

func ExampleNewPageFromRequest() {
	req := httptest.NewRequest("GET", "http://example.com/api?page=3&size=25", nil)
	page := pager.NewPageFromRequest(req)
	fmt.Printf("Page: %d, Size: %d\n", page.Page, page.Size)
	// Output: Page: 3, Size: 25
}

func ExampleNewCursorFromRequest() {
	req := httptest.NewRequest("GET", "http://example.com/api?cursor=abc123&limit=15", nil)
	page := pager.NewCursorFromRequest(req)
	fmt.Printf("Cursor: %s, Limit: %d\n", page.Cursor, page.Limit)
	// Output: Cursor: abc123, Limit: 15
}
