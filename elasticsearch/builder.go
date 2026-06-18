package elasticsearch

import (
	"bytes"
	"encoding/json"
	"io"
)

// Builder 提供 Elasticsearch 搜索请求构建器。
type Builder struct {
	index     string
	boolQuery *BoolQuery
	sort      []any
	from      int64
	size      int64
	minScore  float64
}

// NewBuilder 创建查询构建器。
func NewBuilder(index string) *Builder {
	return &Builder{
		index: index,
		boolQuery: &BoolQuery{
			Must:   make([]Clause, 0),
			Filter: make([]Clause, 0),
			Should: make([]Clause, 0),
		},
	}
}

// Term 添加 term 精确匹配过滤条件。
func (b *Builder) Term(field string, value any) *Builder {
	if field != "" && !isEmptyValue(value) {
		b.boolQuery.Filter = append(b.boolQuery.Filter, NewTermClause(field, value))
	}
	return b
}

// Terms 添加 terms 多值过滤条件。
func (b *Builder) Terms(field string, value any) *Builder {
	if field != "" && !isEmptyValue(value) {
		b.boolQuery.Filter = append(b.boolQuery.Filter, NewTermsClause(field, value))
	}
	return b
}

// TermFilters 批量添加 term 过滤条件。
func (b *Builder) TermFilters(filters map[string]any) *Builder {
	for field, value := range filters {
		b.Term(field, value)
	}
	return b
}

// Match 添加 match 全文搜索条件。
func (b *Builder) Match(field, query string, boost ...float64) *Builder {
	if field != "" && query != "" {
		b.boolQuery.Must = append(b.boolQuery.Must, NewMatchClause(field, query, boost...))
	}
	return b
}

// MultiMatch 添加 multi_match 多字段搜索条件。
func (b *Builder) MultiMatch(query string, fields []string, matchType ...string) *Builder {
	if query != "" && len(fields) > 0 {
		b.boolQuery.Must = append(b.boolQuery.Must, NewMultiMatchClause(query, fields, matchType...))
	}
	return b
}

// Range 添加 range 范围过滤条件。
func (b *Builder) Range(field string, conditions map[string]any) *Builder {
	if field != "" && len(conditions) > 0 {
		b.boolQuery.Filter = append(b.boolQuery.Filter, NewRangeClause(field, conditions))
	}
	return b
}

// RangeGte 添加大于等于过滤条件。
func (b *Builder) RangeGte(field string, value any) *Builder {
	return b.Range(field, map[string]any{"gte": value})
}

// RangeLte 添加小于等于过滤条件。
func (b *Builder) RangeLte(field string, value any) *Builder {
	return b.Range(field, map[string]any{"lte": value})
}

// RangeGt 添加大于过滤条件。
func (b *Builder) RangeGt(field string, value any) *Builder {
	return b.Range(field, map[string]any{"gt": value})
}

// RangeLt 添加小于过滤条件。
func (b *Builder) RangeLt(field string, value any) *Builder {
	return b.Range(field, map[string]any{"lt": value})
}

// RangeBetween 添加范围区间过滤条件。
func (b *Builder) RangeBetween(field string, gte, lte any) *Builder {
	return b.Range(field, map[string]any{"gte": gte, "lte": lte})
}

// Exists 添加 exists 字段存在过滤条件。
func (b *Builder) Exists(field string) *Builder {
	if field != "" {
		b.boolQuery.Filter = append(b.boolQuery.Filter, NewExistsClause(field))
	}
	return b
}

// Wildcard 添加 wildcard 通配符过滤条件。
func (b *Builder) Wildcard(field, value string) *Builder {
	if field != "" && value != "" {
		b.boolQuery.Filter = append(b.boolQuery.Filter, NewWildcardClause(field, value))
	}
	return b
}

// Prefix 添加 prefix 前缀过滤条件。
func (b *Builder) Prefix(field, value string) *Builder {
	if field != "" && value != "" {
		b.boolQuery.Filter = append(b.boolQuery.Filter, NewPrefixClause(field, value))
	}
	return b
}

// Must 添加 must 子句。
func (b *Builder) Must(clause Clause) *Builder {
	b.boolQuery.Must = append(b.boolQuery.Must, clause)
	return b
}

// MustNot 添加 must_not 子句。
func (b *Builder) MustNot(clause Clause) *Builder {
	b.boolQuery.MustNot = append(b.boolQuery.MustNot, clause)
	return b
}

// Should 添加 should 子句。
func (b *Builder) Should(clause Clause) *Builder {
	b.boolQuery.Should = append(b.boolQuery.Should, clause)
	return b
}

// Filter 添加 filter 子句。
func (b *Builder) Filter(clause Clause) *Builder {
	b.boolQuery.Filter = append(b.boolQuery.Filter, clause)
	return b
}

// MinimumShouldMatch 设置 should 子句至少匹配数量。
func (b *Builder) MinimumShouldMatch(count int) *Builder {
	b.boolQuery.MinimumShouldMatch = count
	return b
}

// DisMax 添加 dis_max 查询。
func (b *Builder) DisMax(tieBreaker float64, queries ...Clause) *Builder {
	if len(queries) > 0 {
		b.boolQuery.Must = append(b.boolQuery.Must, NewDisMaxClause(tieBreaker, queries...))
	}
	return b
}

// FunctionScore 添加 function_score 查询。
func (b *Builder) FunctionScore(query *BoolQuery, script *Script, boostMode string) *Builder {
	if script != nil {
		b.boolQuery.Must = append(b.boolQuery.Must, NewFunctionScoreClause(query, script, boostMode))
	}
	return b
}

// Sort 添加排序条件。
func (b *Builder) Sort(sorts ...any) *Builder {
	b.sort = append(b.sort, sorts...)
	return b
}

// SortDesc 添加降序排序。
func (b *Builder) SortDesc(field string) *Builder {
	return b.Sort(NewSortDesc(field))
}

// SortAsc 添加升序排序。
func (b *Builder) SortAsc(field string) *Builder {
	return b.Sort(NewSortAsc(field))
}

// Pager 设置页码分页，页码从 1 开始。
func (b *Builder) Pager(page, size int64) *Builder {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	b.from = (page - 1) * size
	b.size = size
	return b
}

// From 设置 from 偏移。
func (b *Builder) From(from int64) *Builder {
	if from >= 0 {
		b.from = from
	}
	return b
}

// Size 设置返回数量。
func (b *Builder) Size(size int64) *Builder {
	if size > 0 {
		b.size = size
	}
	return b
}

// MinScore 设置最小相关性分数。
func (b *Builder) MinScore(score float64) *Builder {
	b.minScore = score
	return b
}

// BoolQuery 返回当前 bool 查询。
func (b *Builder) BoolQuery() *BoolQuery {
	return b.boolQuery
}

// SetBoolQuery 设置 bool 查询。
func (b *Builder) SetBoolQuery(bq *BoolQuery) *Builder {
	if bq == nil {
		bq = &BoolQuery{}
	}
	b.boolQuery = bq
	return b
}

// Build 构建 query 部分。
func (b *Builder) Build() map[string]any {
	return map[string]any{"bool": b.boolQuery}
}

// Index 返回目标索引。
func (b *Builder) Index() string {
	return b.index
}

// Body 返回搜索请求体。
func (b *Builder) Body() io.Reader {
	body, _ := b.BodyBytes()
	return bytes.NewReader(body)
}

// BodyBytes 返回搜索请求体字节。
func (b *Builder) BodyBytes() ([]byte, error) {
	data := make(map[string]any)

	if hasBoolQuery(b.boolQuery) {
		data["query"] = b.Build()
	}
	if len(b.sort) > 0 {
		data["sort"] = b.sort
	}
	data["from"] = b.from
	if b.size > 0 {
		data["size"] = b.size
	}
	if b.minScore > 0 {
		data["min_score"] = b.minScore
	}

	return json.Marshal(data)
}

func hasBoolQuery(q *BoolQuery) bool {
	return q != nil && (len(q.Must) > 0 || len(q.Should) > 0 || len(q.MustNot) > 0 || len(q.Filter) > 0)
}

var _ Request = (*Builder)(nil)
