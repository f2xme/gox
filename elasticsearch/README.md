# elasticsearch

`elasticsearch` 提供基于官方 `github.com/elastic/go-elasticsearch/v8` 的 Elasticsearch 客户端封装，包含连接配置、查询构建、响应解码、文档写入、索引管理、别名切换和 reindex 常用流程。

这个包直接使用 `github.com/f2xme/gox/elasticsearch`，没有额外 adapter 层。

## 安装

```bash
go get github.com/f2xme/gox/elasticsearch
```

## 快速开始

```go
package main

import (
	"context"
	"log"

	"github.com/f2xme/gox/elasticsearch"
)

type User struct {
	IDValue string `json:"id"`
	Name    string `json:"name"`
	Status  int    `json:"status"`
}

func (u User) ID() string { return u.IDValue }

func main() {
	ctx := context.Background()

	client, err := elasticsearch.New(
		elasticsearch.WithAddresses("http://localhost:9200"),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = client.CreateDoc(ctx, "users", User{
		IDValue: "u1",
		Name:    "Alice",
		Status:  1,
	}, elasticsearch.WithRefresh(true))
	if err != nil {
		log.Fatal(err)
	}

	req := elasticsearch.NewBuilder("users").
		Term("status", 1).
		Match("name", "alice").
		Pager(1, 20)

	result, err := elasticsearch.SearchWithType[User](ctx, client, req)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(result.Total, result.Hits)
}
```

## 连接配置

本地无认证集群：

```go
client, err := elasticsearch.New(
	elasticsearch.WithAddresses("http://localhost:9200"),
)
```

API Key：

```go
client, err := elasticsearch.New(
	elasticsearch.WithAddresses("https://es.example.com:9200"),
	elasticsearch.WithAPIKey("base64-api-key"),
)
```

Basic Auth：

```go
client, err := elasticsearch.New(
	elasticsearch.WithAddresses("https://es.example.com:9200"),
	elasticsearch.WithBasicAuth("elastic", "password"),
)
```

Elastic Cloud：

```go
client, err := elasticsearch.New(
	elasticsearch.WithCloudID("cloud-id"),
	elasticsearch.WithAPIKey("base64-api-key"),
)
```

生产环境建议显式设置连接池、超时和重试：

```go
client, err := elasticsearch.New(
	elasticsearch.WithAddresses("https://es.example.com:9200"),
	elasticsearch.WithAPIKey("base64-api-key"),
	elasticsearch.WithMaxRetries(3),
	elasticsearch.WithMaxIdleConnsPerHost(100),
	elasticsearch.WithDialTimeout(3*time.Second),
	elasticsearch.WithResponseHeaderTimeout(10*time.Second),
	elasticsearch.WithIdleConnTimeout(90*time.Second),
)
```

## 使用 gox/config

`NewWithConfig` 默认读取 `es` 前缀，也可以传入自定义前缀。

```go
client, err := elasticsearch.NewWithConfig(cfg)
client, err = elasticsearch.NewWithConfig(cfg, "elasticsearch")
```

支持的配置键：

| 键 | 类型 | 说明 |
| --- | --- | --- |
| `addresses` | `[]string` | 节点地址 |
| `apiKey` | `string` | API Key |
| `username` | `string` | Basic Auth 用户名 |
| `password` | `string` | Basic Auth 密码 |
| `cloudId` | `string` | Elastic Cloud ID |
| `serviceToken` | `string` | Service Token |
| `maxRetries` | `int` | 最大重试次数 |
| `maxIdleConnsPerHost` | `int` | 每个 Host 最大空闲连接数 |
| `responseHeaderTimeout` | `duration` | 响应头超时 |
| `dialTimeout` | `duration` | 建连超时 |
| `idleConnTimeout` | `duration` | 空闲连接超时 |
| `skipPing` | `bool` | 创建客户端时跳过连通性检查 |

## 查询

链式构建搜索请求：

```go
req := elasticsearch.NewBuilder("users").
	Term("tenant_id", "t1").
	Terms("status", []int{1, 2}).
	MultiMatch("alice", []string{"name", "email"}).
	RangeGte("created_at", "2026-01-01").
	SortDesc("created_at").
	Pager(1, 20)

result, err := elasticsearch.SearchWithType[User](ctx, client, req)
```

返回 map 结果：

```go
result, err := client.Search(ctx, req)
if err != nil {
	return err
}

for _, hit := range result.Hits {
	name := hit.GetString("name")
	status := hit.GetInt64("status")
	_ = name
	_ = status
}
```

统计数量：

```go
total, err := client.Count(ctx, req)
```

`Count` 会只取请求体中的 `query` 字段发送到 `_count` API，避免把 `from`、`size`、`sort` 等 search-only 字段传给 count。

## 写入

写入类型需要实现 `Document` 接口：

```go
type Article struct {
	IDValue string `json:"id"`
	Title   string `json:"title"`
	Status  int    `json:"status"`
}

func (a Article) ID() string { return a.IDValue }
```

创建或覆盖文档：

```go
err := client.CreateDoc(ctx, "articles", article)
```

批量写入：

```go
docs := []elasticsearch.Document{
	Article{IDValue: "a1", Title: "one", Status: 1},
	Article{IDValue: "a2", Title: "two", Status: 1},
}

err := client.CreateBulk(ctx, "articles", docs)
```

局部更新和删除：

```go
err := client.UpdateDoc(ctx, "articles", article)
err = client.DeleteDoc(ctx, "articles", "a1")
```

需要立即刷新时使用 `WithRefresh(true)`：

```go
err := client.CreateDoc(ctx, "articles", article, elasticsearch.WithRefresh(true))
```

高吞吐写入不建议默认开启 refresh。

## 索引和别名

创建索引：

```go
mapping := &elasticsearch.IndexMapping{
	Settings: map[string]any{
		"number_of_shards":   1,
		"number_of_replicas": 1,
	},
	Mappings: map[string]any{
		"properties": map[string]any{
			"title":      map[string]any{"type": "text"},
			"status":     map[string]any{"type": "integer"},
			"created_at": map[string]any{"type": "date"},
		},
	},
}

err := client.CreateIndex(ctx, "articles_v1", mapping)
```

常用管理操作：

```go
exists, err := client.IndexExists(ctx, "articles_v1")
err = client.RefreshIndex(ctx, "articles_v1")
err = client.AddAlias(ctx, "articles_v1", "articles")
indices, err := client.GetAliasIndices(ctx, "articles")
err = client.UpdateAlias(ctx, "articles", indices, "articles_v2")
```

## Reindex

同步复制索引：

```go
err := client.Reindex(ctx, "articles_v1", "articles_v2")
```

异步复制索引：

```go
taskID, err := client.ReindexAsync(ctx, "articles_v1", "articles_v2")
if err != nil {
	return err
}

err = client.WaitForTask(ctx, taskID, 5*time.Second)
```

通过别名重建索引：

```go
err := client.ReindexWithAlias(ctx, "articles", mapping)
```

大索引建议拆成异步三步：

```go
newIndex, taskID, err := client.ReindexWithAliasAsync(ctx, "articles", mapping)
if err != nil {
	return err
}
if taskID != "" {
	if err := client.WaitForTask(ctx, taskID, 5*time.Second); err != nil {
		return err
	}
}
err = client.FinishReindexWithAlias(ctx, "articles", newIndex)
```

## Analyze

```go
body := strings.NewReader(`{
	"analyzer": "standard",
	"text": "hello elasticsearch"
}`)

tokens, err := client.Analyze(ctx, "articles", body)
```

## 访问官方客户端

当 gox 封装还没有覆盖某个 API 时，可以取出官方客户端：

```go
native := client.Native()
resp, err := native.Info()
```

## 生产建议

- 每个业务调用都传入带超时的 `context.Context`。
- 生产环境显式配置 `Addresses` 或 `CloudID`，不要依赖本地默认值。
- 使用 API Key、Basic Auth、Cloud ID 或 Service Token 之一完成认证。
- 写入后必须立即可搜索时再使用 `WithRefresh(true)`。
- 大批量写入使用 `CreateBulk`，并记录返回错误；部分文档失败会作为错误返回。
- 重建索引时让业务代码访问别名，不直接绑定物理索引名。
- 大索引重建使用异步 reindex，等待任务完成后再切换别名。
- 官方客户端未封装的能力通过 `Native()` 使用，避免在 gox 包里重复造底层 API。

## 测试

单元测试：

```bash
go test ./elasticsearch
go test -cover ./elasticsearch
```

本地集成测试需要一个无认证 Elasticsearch：

```bash
docker run --name gox-es-test \
  -p 9200:9200 \
  -e discovery.type=single-node \
  -e xpack.security.enabled=false \
  -e ES_JAVA_OPTS="-Xms512m -Xmx512m" \
  docker.elastic.co/elasticsearch/elasticsearch:8.19.0
```

执行集成测试：

```bash
ELASTICSEARCH_INTEGRATION=1 ELASTICSEARCH_ADDR=http://127.0.0.1:9200 go test -run TestIntegrationLocalElasticsearch -count=1 -timeout=90s -v ./elasticsearch
```

没有设置 `ELASTICSEARCH_INTEGRATION=1` 时，集成测试会自动跳过。
