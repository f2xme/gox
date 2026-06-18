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
