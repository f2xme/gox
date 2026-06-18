package elasticsearch

// SynonymRule 表示同义词规则。
type SynonymRule struct {
	// ID 同义词规则 ID。
	ID string `json:"id,omitempty"`
	// Synonyms Solr 格式的同义词规则。
	Synonyms string `json:"synonyms"`
}

// SynonymSet 表示同义词集合。
type SynonymSet struct {
	// Count 集合中的同义词规则数量。
	Count int64 `json:"count"`
	// Rules 同义词规则列表。
	Rules []SynonymRule `json:"synonyms_set"`
}

// SynonymSetSummary 表示同义词集合摘要。
type SynonymSetSummary struct {
	// SynonymsSet 同义词集合 ID。
	SynonymsSet string `json:"synonyms_set"`
}

// SynonymSetList 表示同义词集合列表。
type SynonymSetList struct {
	// Count 同义词集合总数。
	Count int64 `json:"count"`
	// Results 同义词集合摘要列表。
	Results []SynonymSetSummary `json:"results"`
}

// SynonymUpdateResult 表示同义词更新结果。
type SynonymUpdateResult struct {
	// Acknowledged 表示请求是否被确认。
	Acknowledged bool `json:"acknowledged,omitempty"`
	// ReloadAnalyzersDetails 关联 analyzer 的 reload 明细。
	ReloadAnalyzersDetails []map[string]any `json:"reload_analyzers_details,omitempty"`
	// Raw 保留完整响应，便于读取 Elasticsearch 新增字段。
	Raw map[string]any `json:"-"`
}

// SynonymOptions 定义同义词 API 选项。
type SynonymOptions struct {
	// From 起始偏移。
	From *int
	// Size 返回数量。
	Size *int
}

// SynonymOption 定义同义词 API 选项函数。
type SynonymOption func(*SynonymOptions)

func applySynonymOptions(opts ...SynonymOption) SynonymOptions {
	var options SynonymOptions
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// WithSynonymFrom 设置同义词列表起始偏移。
func WithSynonymFrom(from int) SynonymOption {
	return func(o *SynonymOptions) {
		if from >= 0 {
			o.From = &from
		}
	}
}

// WithSynonymSize 设置同义词列表返回数量。
func WithSynonymSize(size int) SynonymOption {
	return func(o *SynonymOptions) {
		if size > 0 {
			o.Size = &size
		}
	}
}
