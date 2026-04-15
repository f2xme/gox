package pager

import (
	"encoding/base64"
	"fmt"
)

// CursorPage 表示基于游标的分页参数
type CursorPage struct {
	Cursor string
	Limit  int
}

// NewCursor 创建给定游标和 limit 的新游标分页
// 如果 limit <= 0，则使用 DefaultLimit
func NewCursor(cursor string, limit int) CursorPage {
	if limit <= 0 {
		limit = DefaultLimit
	}
	return CursorPage{
		Cursor: cursor,
		Limit:  limit,
	}
}

// CursorResult 表示基于游标的分页结果
type CursorResult[T any] struct {
	Items      []T
	Cursor     string
	NextCursor string
	Limit      int
}

// NewCursorResult 创建新的游标分页结果
func NewCursorResult[T any](page CursorPage, items []T, nextCursor string) CursorResult[T] {
	return CursorResult[T]{
		Items:      items,
		Cursor:     page.Cursor,
		NextCursor: nextCursor,
		Limit:      page.Limit,
	}
}

func (r CursorResult[T]) HasNext() bool {
	return r.NextCursor != ""
}

// Next 使用下一个游标返回下一页
func (r CursorResult[T]) Next() CursorPage {
	return CursorPage{
		Cursor: r.NextCursor,
		Limit:  r.Limit,
	}
}

// EncodeCursor 将游标值编码为 base64 URL 安全编码
// 如果输入为空则返回空字符串
func EncodeCursor(cursor string) string {
	if cursor == "" {
		return ""
	}
	return base64.URLEncoding.EncodeToString([]byte(cursor))
}

// DecodeCursor 解码 base64 URL 安全编码的游标
// 如果输入为空则返回空字符串
func DecodeCursor(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}

	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("invalid cursor: %w", err)
	}

	return string(decoded), nil
}
