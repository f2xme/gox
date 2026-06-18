package elasticsearch

import (
	"encoding/json"
	"reflect"
)

// BoolQuery 表示 Elasticsearch bool 查询。
type BoolQuery struct {
	// Must 表示必须匹配且参与评分的子句。
	Must []Clause `json:"must,omitempty"`
	// Should 表示可选匹配的子句。
	Should []Clause `json:"should,omitempty"`
	// MustNot 表示必须不匹配的子句。
	MustNot []Clause `json:"must_not,omitempty"`
	// Filter 表示必须匹配但不参与评分的过滤子句。
	Filter []Clause `json:"filter,omitempty"`
	// MinimumShouldMatch 表示 should 子句至少匹配数量。
	MinimumShouldMatch int `json:"minimum_should_match,omitempty"`
}

// Clause 表示一个 Elasticsearch 查询子句。
type Clause struct {
	// Term 表示 term 精确匹配查询。
	Term *TermQuery `json:"term,omitempty"`
	// Terms 表示 terms 多值匹配查询。
	Terms *TermsQuery `json:"terms,omitempty"`
	// Match 表示 match 全文查询。
	Match *MatchQuery `json:"match,omitempty"`
	// MultiMatch 表示 multi_match 多字段全文查询。
	MultiMatch *MultiMatchQuery `json:"multi_match,omitempty"`
	// Range 表示 range 范围查询。
	Range *RangeQuery `json:"range,omitempty"`
	// Exists 表示 exists 字段存在查询。
	Exists *ExistsQuery `json:"exists,omitempty"`
	// Wildcard 表示 wildcard 通配符查询。
	Wildcard *WildcardQuery `json:"wildcard,omitempty"`
	// Prefix 表示 prefix 前缀查询。
	Prefix *PrefixQuery `json:"prefix,omitempty"`
	// FunctionScore 表示 function_score 查询。
	FunctionScore *FunctionScoreQuery `json:"function_score,omitempty"`
	// DisMax 表示 dis_max 查询。
	DisMax *DisMaxQuery `json:"dis_max,omitempty"`
	// Bool 表示嵌套 bool 查询。
	Bool *BoolQuery `json:"bool,omitempty"`
}

// TermQuery 表示 term 精确匹配查询。
type TermQuery struct {
	// Field 查询字段。
	Field string `json:"-"`
	// Value 查询值。
	Value any `json:"-"`
}

// TermsQuery 表示 terms 多值匹配查询。
type TermsQuery struct {
	// Field 查询字段。
	Field string `json:"-"`
	// Value 查询值列表。
	Value any `json:"-"`
}

// MatchQuery 表示 match 全文查询。
type MatchQuery struct {
	// Field 查询字段。
	Field string `json:"-"`
	// Query 查询文本。
	Query string `json:"-"`
	// Boost 权重，0 表示不设置。
	Boost float64 `json:"-"`
}

// MultiMatchQuery 表示 multi_match 多字段全文查询。
type MultiMatchQuery struct {
	// Query 查询文本。
	Query string `json:"query"`
	// Fields 查询字段。
	Fields []string `json:"fields,omitempty"`
	// Type 查询类型。
	Type string `json:"type,omitempty"`
}

// RangeQuery 表示 range 范围查询。
type RangeQuery struct {
	// Field 查询字段。
	Field string `json:"-"`
	// Range 范围条件，例如 gte、lte、gt、lt。
	Range map[string]any `json:"-"`
}

// ExistsQuery 表示 exists 字段存在查询。
type ExistsQuery struct {
	// Field 查询字段。
	Field string `json:"field"`
}

// WildcardQuery 表示 wildcard 通配符查询。
type WildcardQuery struct {
	// Field 查询字段。
	Field string `json:"-"`
	// Value 通配符值。
	Value string `json:"-"`
}

// PrefixQuery 表示 prefix 前缀查询。
type PrefixQuery struct {
	// Field 查询字段。
	Field string `json:"-"`
	// Value 前缀值。
	Value string `json:"-"`
}

// FunctionScoreQuery 表示 function_score 查询。
type FunctionScoreQuery struct {
	// Query 被评分的 bool 查询。
	Query *BoolQuery `json:"query,omitempty"`
	// ScriptScore 脚本评分配置。
	ScriptScore *ScriptScore `json:"script_score,omitempty"`
	// BoostMode 分数合并模式。
	BoostMode string `json:"boost_mode,omitempty"`
}

// DisMaxQuery 表示 dis_max 查询。
type DisMaxQuery struct {
	// TieBreaker 平局分数系数。
	TieBreaker float64 `json:"tie_breaker"`
	// Queries 候选查询。
	Queries []Clause `json:"queries"`
}

// ScriptScore 表示脚本评分。
type ScriptScore struct {
	// Script 脚本内容。
	Script *Script `json:"script"`
}

// Script 表示 Elasticsearch 脚本。
type Script struct {
	// Source 脚本源码。
	Source string `json:"source"`
	// Params 脚本参数。
	Params map[string]any `json:"params,omitempty"`
}

// MarshalJSON 自定义查询子句 JSON 编码。
func (c Clause) MarshalJSON() ([]byte, error) {
	switch {
	case c.Term != nil:
		return json.Marshal(map[string]any{"term": c.Term})
	case c.Terms != nil:
		return json.Marshal(map[string]any{"terms": c.Terms})
	case c.Match != nil:
		return json.Marshal(map[string]any{"match": c.Match})
	case c.MultiMatch != nil:
		return json.Marshal(map[string]any{"multi_match": c.MultiMatch})
	case c.Range != nil:
		return json.Marshal(map[string]any{"range": c.Range})
	case c.Exists != nil:
		return json.Marshal(map[string]any{"exists": c.Exists})
	case c.Wildcard != nil:
		return json.Marshal(map[string]any{"wildcard": c.Wildcard})
	case c.Prefix != nil:
		return json.Marshal(map[string]any{"prefix": c.Prefix})
	case c.FunctionScore != nil:
		return json.Marshal(map[string]any{"function_score": c.FunctionScore})
	case c.DisMax != nil:
		return json.Marshal(map[string]any{"dis_max": c.DisMax})
	case c.Bool != nil:
		return json.Marshal(map[string]any{"bool": c.Bool})
	default:
		return json.Marshal(map[string]any{})
	}
}

// MarshalJSON 自定义 term 查询 JSON 编码。
func (t TermQuery) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{t.Field: t.Value})
}

// MarshalJSON 自定义 terms 查询 JSON 编码。
func (t TermsQuery) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{t.Field: t.Value})
}

// MarshalJSON 自定义 match 查询 JSON 编码。
func (m MatchQuery) MarshalJSON() ([]byte, error) {
	if m.Boost > 0 {
		return json.Marshal(map[string]any{
			m.Field: map[string]any{"query": m.Query, "boost": m.Boost},
		})
	}
	return json.Marshal(map[string]any{m.Field: m.Query})
}

// MarshalJSON 自定义 range 查询 JSON 编码。
func (r RangeQuery) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{r.Field: r.Range})
}

// MarshalJSON 自定义 wildcard 查询 JSON 编码。
func (w WildcardQuery) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{w.Field: w.Value})
}

// MarshalJSON 自定义 prefix 查询 JSON 编码。
func (p PrefixQuery) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{p.Field: p.Value})
}

// MarshalJSON 自定义 function_score 查询 JSON 编码。
func (f FunctionScoreQuery) MarshalJSON() ([]byte, error) {
	result := make(map[string]any)
	if f.Query != nil {
		result["query"] = map[string]any{"bool": f.Query}
	}
	if f.ScriptScore != nil {
		result["script_score"] = f.ScriptScore
	}
	if f.BoostMode != "" {
		result["boost_mode"] = f.BoostMode
	}
	return json.Marshal(result)
}

func isEmptyValue(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	}
	return false
}
