package elasticsearch

// NewTerm 创建 term 查询条件。
func NewTerm(field string, value any) map[string]any {
	return map[string]any{"term": map[string]any{field: value}}
}

// NewTerms 创建 terms 查询条件。
func NewTerms(field string, values any) map[string]any {
	return map[string]any{"terms": map[string]any{field: values}}
}

// NewMatch 创建 match 查询条件。
func NewMatch(field, query string) map[string]any {
	return map[string]any{"match": map[string]any{field: query}}
}

// NewMatchWithBoost 创建带权重的 match 查询条件。
func NewMatchWithBoost(field, query string, boost float64) map[string]any {
	return map[string]any{
		"match": map[string]any{
			field: map[string]any{"query": query, "boost": boost},
		},
	}
}

// NewMultiMatch 创建 multi_match 查询条件。
func NewMultiMatch(query string, fields []string, matchType ...string) map[string]any {
	mm := map[string]any{"query": query, "fields": fields}
	if len(matchType) > 0 {
		mm["type"] = matchType[0]
	}
	return map[string]any{"multi_match": mm}
}

// NewRange 创建 range 查询条件。
func NewRange(field string, conditions map[string]any) map[string]any {
	return map[string]any{"range": map[string]any{field: conditions}}
}

// NewRangeGte 创建 gte 范围查询条件。
func NewRangeGte(field string, value any) map[string]any {
	return NewRange(field, map[string]any{"gte": value})
}

// NewRangeLte 创建 lte 范围查询条件。
func NewRangeLte(field string, value any) map[string]any {
	return NewRange(field, map[string]any{"lte": value})
}

// NewRangeGt 创建 gt 范围查询条件。
func NewRangeGt(field string, value any) map[string]any {
	return NewRange(field, map[string]any{"gt": value})
}

// NewRangeLt 创建 lt 范围查询条件。
func NewRangeLt(field string, value any) map[string]any {
	return NewRange(field, map[string]any{"lt": value})
}

// NewRangeBetween 创建范围区间查询条件。
func NewRangeBetween(field string, gte, lte any) map[string]any {
	return NewRange(field, map[string]any{"gte": gte, "lte": lte})
}

// NewExists 创建 exists 查询条件。
func NewExists(field string) map[string]any {
	return map[string]any{"exists": map[string]any{"field": field}}
}

// NewWildcard 创建 wildcard 查询条件。
func NewWildcard(field, value string) map[string]any {
	return map[string]any{"wildcard": map[string]any{field: value}}
}

// NewPrefix 创建 prefix 查询条件。
func NewPrefix(field, value string) map[string]any {
	return map[string]any{"prefix": map[string]any{field: value}}
}

// NewBool 创建 bool 查询条件。
func NewBool(conditions map[string]any) map[string]any {
	return map[string]any{"bool": conditions}
}

// NewMust 创建 must 查询条件。
func NewMust(conditions ...any) map[string]any {
	return map[string]any{"must": conditions}
}

// NewMustNot 创建 must_not 查询条件。
func NewMustNot(conditions ...any) map[string]any {
	return map[string]any{"must_not": conditions}
}

// NewShould 创建 should 查询条件。
func NewShould(conditions ...any) map[string]any {
	return map[string]any{"should": conditions}
}

// NewFilter 创建 filter 查询条件。
func NewFilter(conditions ...any) map[string]any {
	return map[string]any{"filter": conditions}
}

// NewDisMax 创建 dis_max 查询条件。
func NewDisMax(tieBreaker float64, queries ...any) map[string]any {
	return map[string]any{
		"dis_max": map[string]any{
			"tie_breaker": tieBreaker,
			"queries":     queries,
		},
	}
}
