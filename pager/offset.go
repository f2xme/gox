package pager

// 分页默认值
const (
	DefaultLimit  = 10
	DefaultOffset = 0
)

// OffsetPage 表示基于偏移量的分页参数
type OffsetPage struct {
	Limit  int
	Offset int
}

// NewOffset 创建给定 limit 和 offset 的新偏移量分页
// 如果 limit <= 0，则使用 DefaultLimit
// 如果 offset < 0，则使用 0
func NewOffset(limit, offset int) OffsetPage {
	if limit <= 0 {
		limit = DefaultLimit
	}
	if offset < 0 {
		offset = 0
	}
	return OffsetPage{
		Limit:  limit,
		Offset: offset,
	}
}

// Next 返回下一页
func (p OffsetPage) Next() OffsetPage {
	return OffsetPage{
		Limit:  p.Limit,
		Offset: p.Offset + p.Limit,
	}
}

// Prev 返回上一页
func (p OffsetPage) Prev() OffsetPage {
	offset := max(0, p.Offset-p.Limit)
	return OffsetPage{
		Limit:  p.Limit,
		Offset: offset,
	}
}

// OffsetResult 表示基于偏移量的分页结果
type OffsetResult[T any] struct {
	Items  []T
	Total  int64
	Limit  int
	Offset int
}

// NewOffsetResult 创建新的偏移量分页结果
func NewOffsetResult[T any](page OffsetPage, items []T, total int64) OffsetResult[T] {
	return OffsetResult[T]{
		Items:  items,
		Total:  total,
		Limit:  page.Limit,
		Offset: page.Offset,
	}
}

func (r OffsetResult[T]) HasNext() bool {
	return int64(r.Offset)+int64(r.Limit) < r.Total
}

func (r OffsetResult[T]) HasPrev() bool {
	return r.Offset > 0
}
