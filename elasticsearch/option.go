package elasticsearch

// RequestOption 定义搜索请求构建选项。
type RequestOption func(*Builder)

// NewRequest 使用选项创建搜索请求。
//
// 推荐直接使用 NewBuilder 进行链式构建。
func NewRequest(index string, opts ...RequestOption) Request {
	b := NewBuilder(index)
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// WithQuery 设置 multi_match 搜索关键词。
func WithQuery(query string) RequestOption {
	return func(b *Builder) {
		if query != "" {
			b.boolQuery.Must = append(b.boolQuery.Must, Clause{
				MultiMatch: &MultiMatchQuery{Query: query},
			})
		}
	}
}

// WithQueryType 设置最近一次 multi_match 的查询类型。
func WithQueryType(queryType string) RequestOption {
	return func(b *Builder) {
		if len(b.boolQuery.Must) == 0 {
			return
		}
		last := &b.boolQuery.Must[len(b.boolQuery.Must)-1]
		if last.MultiMatch != nil {
			last.MultiMatch.Type = queryType
		}
	}
}

// WithFields 设置最近一次 multi_match 的查询字段。
func WithFields(fields ...string) RequestOption {
	return func(b *Builder) {
		if len(b.boolQuery.Must) == 0 {
			return
		}
		last := &b.boolQuery.Must[len(b.boolQuery.Must)-1]
		if last.MultiMatch != nil {
			last.MultiMatch.Fields = fields
			if last.MultiMatch.Type == "" {
				last.MultiMatch.Type = "phrase"
			}
		}
	}
}

// WithPager 设置页码分页。
func WithPager(page, size int64) RequestOption {
	return func(b *Builder) {
		b.Pager(page, size)
	}
}

// WithFrom 设置 from 偏移。
func WithFrom(from int64) RequestOption {
	return func(b *Builder) {
		b.From(from)
	}
}

// WithSize 设置返回数量。
func WithSize(size int64) RequestOption {
	return func(b *Builder) {
		b.Size(size)
	}
}

// WithFilter 添加 filter 子句。
func WithFilter(filters ...Clause) RequestOption {
	return func(b *Builder) {
		for _, f := range filters {
			b.Filter(f)
		}
	}
}

// WithMust 添加 must 子句。
func WithMust(conditions ...Clause) RequestOption {
	return func(b *Builder) {
		for _, c := range conditions {
			b.Must(c)
		}
	}
}

// WithMustNot 添加 must_not 子句。
func WithMustNot(conditions ...Clause) RequestOption {
	return func(b *Builder) {
		for _, c := range conditions {
			b.MustNot(c)
		}
	}
}

// WithSort 添加排序条件。
func WithSort(sorts ...any) RequestOption {
	return func(b *Builder) {
		b.Sort(sorts...)
	}
}
