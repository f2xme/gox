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

// GetOffset 返回 SQL 使用的 offset 值
func (p PageNumber) GetOffset() int {
	page := p.normalize()
	return (page.Page - 1) * page.Size
}

// GetLimit 返回 SQL 使用的 limit 值
func (p PageNumber) GetLimit() int {
	return p.normalize().Size
}

func (p PageNumber) normalize() PageNumber {
	return NewPage(p.Page, p.Size)
}
