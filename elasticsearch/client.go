package elasticsearch

import (
	"context"
	"io"
)

// Document 表示可写入 Elasticsearch 的文档。
type Document interface {
	// ID 返回文档 ID。
	ID() string
}

// Request 表示一个 Elasticsearch 请求体。
type Request interface {
	// Index 返回请求目标索引。
	Index() string
	// Body 返回请求体读取器。
	Body() io.Reader
}

// Searcher 定义搜索能力。
type Searcher interface {
	// Search 执行搜索并返回 map 版本结果。
	Search(ctx context.Context, req Request) (*SearchResult[*HitMap], error)
}

// Counter 定义统计能力。
type Counter interface {
	// Count 统计匹配文档数量。
	Count(ctx context.Context, req Request) (int64, error)
}

// Analyzer 定义分词分析能力。
type Analyzer interface {
	// Analyze 执行 Elasticsearch analyze 请求。
	Analyze(ctx context.Context, index string, body io.Reader) ([]AnalyzeToken, error)
}

// SynonymManager 定义同义词管理能力。
type SynonymManager interface {
	// ListSynonymSets 获取同义词集合列表。
	ListSynonymSets(ctx context.Context, opts ...SynonymOption) (*SynonymSetList, error)
	// GetSynonymSet 获取同义词集合。
	GetSynonymSet(ctx context.Context, id string, opts ...SynonymOption) (*SynonymSet, error)
	// PutSynonymSet 创建或更新同义词集合。
	PutSynonymSet(ctx context.Context, id string, rules []SynonymRule) (*SynonymUpdateResult, error)
	// DeleteSynonymSet 删除同义词集合。
	DeleteSynonymSet(ctx context.Context, id string) error
	// GetSynonymRule 获取同义词集合中的单条规则。
	GetSynonymRule(ctx context.Context, setID, ruleID string) (*SynonymRule, error)
	// PutSynonymRule 创建或更新同义词集合中的单条规则。
	PutSynonymRule(ctx context.Context, setID, ruleID, synonyms string) (*SynonymUpdateResult, error)
	// DeleteSynonymRule 删除同义词集合中的单条规则。
	DeleteSynonymRule(ctx context.Context, setID, ruleID string) error
}

// TaskManager 定义任务管理能力。
type TaskManager interface {
	// GetTask 获取单个任务信息。
	GetTask(ctx context.Context, taskID string, opts ...TaskOption) (*TaskResponse, error)
	// ListTasks 获取任务列表。
	ListTasks(ctx context.Context, opts ...TaskOption) (*TaskListResponse, error)
	// CancelTasks 取消匹配条件的任务。
	CancelTasks(ctx context.Context, opts ...TaskOption) (*TaskCancelResponse, error)
	// CancelTask 取消指定任务。
	CancelTask(ctx context.Context, taskID string, opts ...TaskOption) (*TaskCancelResponse, error)
}

// IndexManager 定义索引和别名管理能力。
type IndexManager interface {
	// CreateIndex 创建索引。
	CreateIndex(ctx context.Context, index string, mapping *IndexMapping) error
	// DeleteIndex 删除索引。
	DeleteIndex(ctx context.Context, index string) error
	// IndexExists 判断索引是否存在。
	IndexExists(ctx context.Context, index string) (bool, error)
	// RefreshIndex 刷新索引。
	RefreshIndex(ctx context.Context, index string) error
	// GetAliasIndices 获取别名指向的索引列表。
	GetAliasIndices(ctx context.Context, alias string) ([]string, error)
	// UpdateAlias 原子更新别名。
	UpdateAlias(ctx context.Context, alias string, oldIndices []string, newIndex string) error
}

// Writer 定义文档写入能力。
type Writer interface {
	// CreateDoc 创建或覆盖文档。
	CreateDoc(ctx context.Context, index string, doc Document, opts ...WriteOption) error
	// CreateBulk 批量创建或覆盖文档。
	CreateBulk(ctx context.Context, index string, docs []Document, opts ...WriteOption) error
	// UpdateDoc 局部更新文档。
	UpdateDoc(ctx context.Context, index string, doc Document, opts ...WriteOption) error
	// DeleteDoc 删除文档。
	DeleteDoc(ctx context.Context, index string, id string, opts ...WriteOption) error
}
