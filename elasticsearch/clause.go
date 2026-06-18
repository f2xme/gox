package elasticsearch

// NewTermClause 创建 term 子句。
func NewTermClause(field string, value any) Clause {
	return Clause{Term: &TermQuery{Field: field, Value: value}}
}

// NewTermsClause 创建 terms 子句。
func NewTermsClause(field string, value any) Clause {
	return Clause{Terms: &TermsQuery{Field: field, Value: value}}
}

// NewMatchClause 创建 match 子句。
func NewMatchClause(field, query string, boost ...float64) Clause {
	var b float64
	if len(boost) > 0 {
		b = boost[0]
	}
	return Clause{Match: &MatchQuery{Field: field, Query: query, Boost: b}}
}

// NewMultiMatchClause 创建 multi_match 子句。
func NewMultiMatchClause(query string, fields []string, matchType ...string) Clause {
	mm := &MultiMatchQuery{Query: query, Fields: fields}
	if len(matchType) > 0 {
		mm.Type = matchType[0]
	}
	return Clause{MultiMatch: mm}
}

// NewRangeClause 创建 range 子句。
func NewRangeClause(field string, conditions map[string]any) Clause {
	return Clause{Range: &RangeQuery{Field: field, Range: conditions}}
}

// NewRangeGteClause 创建 gte 范围子句。
func NewRangeGteClause(field string, value any) Clause {
	return NewRangeClause(field, map[string]any{"gte": value})
}

// NewRangeLteClause 创建 lte 范围子句。
func NewRangeLteClause(field string, value any) Clause {
	return NewRangeClause(field, map[string]any{"lte": value})
}

// NewRangeGtClause 创建 gt 范围子句。
func NewRangeGtClause(field string, value any) Clause {
	return NewRangeClause(field, map[string]any{"gt": value})
}

// NewRangeLtClause 创建 lt 范围子句。
func NewRangeLtClause(field string, value any) Clause {
	return NewRangeClause(field, map[string]any{"lt": value})
}

// NewRangeBetweenClause 创建范围区间子句。
func NewRangeBetweenClause(field string, gte, lte any) Clause {
	return NewRangeClause(field, map[string]any{"gte": gte, "lte": lte})
}

// NewExistsClause 创建 exists 子句。
func NewExistsClause(field string) Clause {
	return Clause{Exists: &ExistsQuery{Field: field}}
}

// NewWildcardClause 创建 wildcard 子句。
func NewWildcardClause(field, value string) Clause {
	return Clause{Wildcard: &WildcardQuery{Field: field, Value: value}}
}

// NewPrefixClause 创建 prefix 子句。
func NewPrefixClause(field, value string) Clause {
	return Clause{Prefix: &PrefixQuery{Field: field, Value: value}}
}

// NewDisMaxClause 创建 dis_max 子句。
func NewDisMaxClause(tieBreaker float64, queries ...Clause) Clause {
	return Clause{DisMax: &DisMaxQuery{TieBreaker: tieBreaker, Queries: queries}}
}

// NewBoolClause 创建嵌套 bool 子句。
func NewBoolClause(bq *BoolQuery) Clause {
	return Clause{Bool: bq}
}

// NewFunctionScoreClause 创建 function_score 子句。
func NewFunctionScoreClause(query *BoolQuery, script *Script, boostMode string) Clause {
	return Clause{
		FunctionScore: &FunctionScoreQuery{
			Query:       query,
			ScriptScore: &ScriptScore{Script: script},
			BoostMode:   boostMode,
		},
	}
}

// NewScript 创建脚本。
func NewScript(source string, params map[string]any) *Script {
	return &Script{Source: source, Params: params}
}
