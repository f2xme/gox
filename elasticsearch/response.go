package elasticsearch

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// SearchResult 表示搜索结果。
type SearchResult[T any] struct {
	// Total 匹配总数。
	Total int64
	// Hits 搜索命中数据。
	Hits []T
}

// Response 表示 Elasticsearch search 响应。
type Response[T any] struct {
	// Took 查询耗时，单位毫秒。
	Took int64 `json:"took"`
	// Timeout 表示查询是否超时。
	Timeout bool `json:"timed_out"`
	// Hits 命中信息。
	Hits Hits[T] `json:"hits"`
}

// Hits 表示 Elasticsearch hits 节点。
type Hits[T any] struct {
	// Total 命中总数。
	Total Total `json:"total"`
	// Hits 命中文档列表。
	Hits []Hit[T] `json:"hits"`
}

// Total 表示命中总数。
type Total struct {
	// Value 命中数量。
	Value int64 `json:"value"`
	// Relation 命中数量关系。
	Relation string `json:"relation"`
}

// Hit 表示单条命中文档。
type Hit[T any] struct {
	// Index 索引名称。
	Index string `json:"_index"`
	// ID 文档 ID。
	ID string `json:"_id"`
	// Score 相关性分数。
	Score float64 `json:"_score"`
	// Source 文档内容。
	Source T `json:"_source"`
}

// HitMap 表示 map 版本的命中文档。
type HitMap struct {
	// Index 索引名称。
	Index string `json:"_index"`
	// ID 文档 ID。
	ID string `json:"_id"`
	// Score 相关性分数。
	Score float64 `json:"_score"`
	// Source 文档内容。
	Source map[string]any `json:"_source"`
}

// Set 设置字段值。
func (h *HitMap) Set(key string, v any) {
	if h.Source == nil {
		h.Source = make(map[string]any)
	}
	h.Source[key] = v
}

// Get 获取字段值。
func (h *HitMap) Get(key string) any {
	if h.Source == nil {
		return nil
	}
	return h.Source[key]
}

// GetString 获取字符串字段。
func (h *HitMap) GetString(key string) string {
	v := h.Get(key)
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// GetStringSlice 获取字符串切片字段。
func (h *HitMap) GetStringSlice(key string) []string {
	v := h.Get(key)
	switch value := v.(type) {
	case []string:
		return value
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			out = append(out, fmt.Sprint(item))
		}
		return out
	default:
		return nil
	}
}

// GetBool 获取布尔字段。
func (h *HitMap) GetBool(key string) bool {
	v := h.Get(key)
	switch value := v.(type) {
	case bool:
		return value
	case string:
		ok, _ := strconv.ParseBool(value)
		return ok
	default:
		return false
	}
}

// GetInt64 获取 int64 字段。
func (h *HitMap) GetInt64(key string) int64 {
	v := h.Get(key)
	switch value := v.(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	case json.Number:
		n, _ := value.Int64()
		return n
	case string:
		n, _ := strconv.ParseInt(value, 10, 64)
		return n
	default:
		return 0
	}
}

// GetFloat64 获取 float64 字段。
func (h *HitMap) GetFloat64(key string) float64 {
	v := h.Get(key)
	switch value := v.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case json.Number:
		n, _ := value.Float64()
		return n
	case string:
		n, _ := strconv.ParseFloat(value, 64)
		return n
	default:
		return 0
	}
}

// DecodeSearchResponse 解码完整搜索响应。
func DecodeSearchResponse[T any](r io.Reader) (*Response[T], error) {
	var resp Response[T]
	if err := json.NewDecoder(r).Decode(&resp); err != nil {
		return nil, fmt.Errorf("es: decode search response: %w", err)
	}
	return &resp, nil
}

// DecodeSearchSources 解码搜索响应并返回 _source 列表。
func DecodeSearchSources[T any](r io.Reader) (*SearchResult[T], error) {
	resp, err := DecodeSearchResponse[T](r)
	if err != nil {
		return nil, err
	}

	result := &SearchResult[T]{
		Total: resp.Hits.Total.Value,
		Hits:  make([]T, 0, len(resp.Hits.Hits)),
	}
	for _, hit := range resp.Hits.Hits {
		result.Hits = append(result.Hits, hit.Source)
	}
	return result, nil
}

// DecodeSearchHitMaps 解码搜索响应并返回 map 版本命中文档。
func DecodeSearchHitMaps(r io.Reader) (*SearchResult[*HitMap], error) {
	resp, err := DecodeSearchResponse[map[string]any](r)
	if err != nil {
		return nil, err
	}

	result := &SearchResult[*HitMap]{
		Total: resp.Hits.Total.Value,
		Hits:  make([]*HitMap, 0, len(resp.Hits.Hits)),
	}
	for _, hit := range resp.Hits.Hits {
		result.Hits = append(result.Hits, &HitMap{
			Index:  hit.Index,
			ID:     hit.ID,
			Score:  hit.Score,
			Source: hit.Source,
		})
	}
	return result, nil
}

// DecodeCountResponse 解码 count 响应。
func DecodeCountResponse(r io.Reader) (int64, error) {
	var resp struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(r).Decode(&resp); err != nil {
		return 0, fmt.Errorf("es: decode count response: %w", err)
	}
	return resp.Count, nil
}
