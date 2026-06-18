package elasticsearch

import (
	"fmt"
	"time"
)

// IndexMapping 表示索引 settings 和 mappings 配置。
type IndexMapping struct {
	// Settings 索引 settings。
	Settings map[string]any `json:"settings,omitempty"`
	// Mappings 索引 mappings。
	Mappings map[string]any `json:"mappings,omitempty"`
}

// AliasAction 表示别名更新操作。
type AliasAction struct {
	// Add 添加别名操作。
	Add *AliasActionItem `json:"add,omitempty"`
	// Remove 移除别名操作。
	Remove *AliasActionItem `json:"remove,omitempty"`
}

// AliasActionItem 表示别名操作参数。
type AliasActionItem struct {
	// Index 索引名称。
	Index string `json:"index"`
	// Alias 别名。
	Alias string `json:"alias"`
}

// ReindexRequest 表示 reindex 请求。
type ReindexRequest struct {
	// Source 源索引配置。
	Source ReindexSource `json:"source"`
	// Dest 目标索引配置。
	Dest ReindexDest `json:"dest"`
}

// ReindexSource 表示 reindex 源索引。
type ReindexSource struct {
	// Index 源索引名称。
	Index string `json:"index"`
}

// ReindexDest 表示 reindex 目标索引。
type ReindexDest struct {
	// Index 目标索引名称。
	Index string `json:"index"`
}

// GenerateVersionedIndex 生成带毫秒时间戳版本号的索引名。
func GenerateVersionedIndex(baseIndex string) string {
	return fmt.Sprintf("%s_%d", baseIndex, time.Now().UnixMilli())
}

// FilterIndexSettings 过滤创建索引时不可设置的系统属性。
func FilterIndexSettings(settings map[string]any) map[string]any {
	excludeKeys := map[string]bool{
		"creation_date":         true,
		"uuid":                  true,
		"version":               true,
		"provided_name":         true,
		"routing":               true,
		"history":               true,
		"resize":                true,
		"source_only":           true,
		"shrink":                true,
		"split":                 true,
		"verified_before_close": true,
	}

	filtered := make(map[string]any)
	for key, value := range settings {
		if !excludeKeys[key] {
			filtered[key] = value
		}
	}
	return filtered
}
