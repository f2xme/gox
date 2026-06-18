package elasticsearch

const (
	// OrderAsc 表示升序排序。
	OrderAsc = "asc"
	// OrderDesc 表示降序排序。
	OrderDesc = "desc"
)

// NewSort 创建排序条件。
func NewSort(field, order string) map[string]any {
	return map[string]any{field: map[string]string{"order": order}}
}

// NewSortAsc 创建升序排序条件。
func NewSortAsc(field string) map[string]any {
	return NewSort(field, OrderAsc)
}

// NewSortDesc 创建降序排序条件。
func NewSortDesc(field string) map[string]any {
	return NewSort(field, OrderDesc)
}

// NewScoreSort 创建按相关性分数降序排序条件。
func NewScoreSort() map[string]any {
	return NewSortDesc("_score")
}

// NewMap 创建单键 map。
func NewMap(key string, value any) map[string]any {
	return map[string]any{key: value}
}
