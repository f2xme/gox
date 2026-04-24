package pager

// 页码分页默认值
const (
	DefaultPage = 1
	DefaultSize = 10
)

// PageNumber 表示基于页码的分页参数
type PageNumber struct {
	Page int `json:"pn" form:"pn"`
	Size int `json:"ps" form:"ps"`
}

// NewPage 创建给定页码和大小的新页码分页
// 如果 page <= 0，则使用 DefaultPage (1)
// 如果 size <= 0，则使用 DefaultSize
func NewPage(page, size int) PageNumber {
	if page <= 0 {
		page = DefaultPage
	}
	if size <= 0 {
		size = DefaultSize
	}
	return PageNumber{
		Page: page,
		Size: size,
	}
}

// Next 返回下一页
func (p PageNumber) Next() PageNumber {
	return PageNumber{
		Page: p.Page + 1,
		Size: p.Size,
	}
}

// Prev 返回上一页
func (p PageNumber) Prev() PageNumber {
	page := max(1, p.Page-1)
	return PageNumber{
		Page: page,
		Size: p.Size,
	}
}

// ToOffset 将页码分页转换为偏移量分页
func (p PageNumber) ToOffset() OffsetPage {
	offset := (p.Page - 1) * p.Size
	return OffsetPage{
		Limit:  p.Size,
		Offset: offset,
	}
}

// PageResult 表示基于页码的分页结果
type PageResult[T any] struct {
	Items      []T
	Total      int64
	Page       int
	Size       int
	TotalPages int
}

// NewPageResult 创建新的页码分页结果
func NewPageResult[T any](page PageNumber, items []T, total int64) PageResult[T] {
	return PageResult[T]{
		Items:      items,
		Total:      total,
		Page:       page.Page,
		Size:       page.Size,
		TotalPages: CalculateTotalPages(total, page.Size),
	}
}

func (r PageResult[T]) HasNext() bool {
	return r.Page < r.TotalPages
}

func (r PageResult[T]) HasPrev() bool {
	return r.Page > 1
}

// GetOffset 返回 GORM 使用的 offset 值
func (p PageNumber) GetOffset() int {
	return (p.Page - 1) * p.Size
}

// GetLimit 返回 GORM 使用的 limit 值
func (p PageNumber) GetLimit() int {
	return p.Size
}

// CalculateTotalPages 根据总项数和页面大小计算总页数
func CalculateTotalPages(total int64, size int) int {
	if total == 0 || size <= 0 {
		return 0
	}
	t := int(total)
	pages := t / size
	if t%size > 0 {
		pages++
	}
	return pages
}
