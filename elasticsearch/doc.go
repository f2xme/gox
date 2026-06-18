/*
Package elasticsearch 提供 Elasticsearch 客户端封装和查询构建工具。

elasticsearch 包基于官方 github.com/elastic/go-elasticsearch/v8 客户端，
提供更符合 gox 风格的构造函数、Options 配置、链式查询构建器和响应解码工具。

# 功能特性

  - 基于官方 go-elasticsearch/v8 客户端
  - 支持 bool、term、terms、match、multi_match、range 等常用查询
  - 支持链式构建搜索请求、分页、排序和 min_score
  - 支持泛型搜索和 map 版本搜索结果
  - 支持文档创建、批量写入、更新、删除
  - 支持索引、别名、reindex、analyze 等常用管理操作
  - 支持 synonyms 同义词集合和规则管理
  - 支持 task 查询、列表和取消
  - 支持 API Key、Basic Auth、Cloud ID、Service Token

# 快速开始

基本使用：

	package main

	import (
		"context"
		"log"

		"github.com/f2xme/gox/elasticsearch"
	)

	type User struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status int    `json:"status"`
	}

	func main() {
		client, err := elasticsearch.New(
			elasticsearch.WithAddresses("http://localhost:9200"),
			elasticsearch.WithAPIKey("api-key"),
		)
		if err != nil {
			log.Fatal(err)
		}

		req := elasticsearch.NewBuilder("users").
			Term("status", 1).
			Match("name", "alice").
			SortDesc("created_at").
			Pager(1, 20)

		result, err := elasticsearch.SearchWithType[User](context.Background(), client, req)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(result.Total, result.Hits)
	}

# 连接配置

本地开发或无认证集群：

	client, err := elasticsearch.New(
		elasticsearch.WithAddresses("http://localhost:9200"),
	)

生产环境通常应显式设置超时、重试和认证：

	client, err := elasticsearch.New(
		elasticsearch.WithAddresses("https://es.example.com:9200"),
		elasticsearch.WithAPIKey("base64-api-key"),
		elasticsearch.WithMaxRetries(3),
		elasticsearch.WithDialTimeout(3*time.Second),
		elasticsearch.WithResponseHeaderTimeout(10*time.Second),
	)

也可以通过 gox/config 读取配置：

	client, err := elasticsearch.NewWithConfig(cfg, "elasticsearch")

配置键包括 addresses、apiKey、username、password、cloudId、serviceToken、
maxRetries、maxIdleConnsPerHost、responseHeaderTimeout、dialTimeout、
idleConnTimeout 和 skipPing。

# 查询与统计

Builder 同时实现 Request 接口，可直接传给 Search 和 Count：

	req := elasticsearch.NewBuilder("users").
		Term("tenant_id", "t1").
		MultiMatch("alice", []string{"name", "email"}).
		RangeGte("created_at", "2026-01-01").
		SortDesc("created_at").
		Pager(1, 20)

	result, err := elasticsearch.SearchWithType[User](ctx, client, req)
	total, err := client.Count(ctx, req)

# 写入与索引管理

写入文档的类型需要实现 Document 接口：

	type User struct {
		IDValue string `json:"id"`
		Name    string `json:"name"`
	}

	func (u User) ID() string { return u.IDValue }

	err := client.CreateDoc(ctx, "users", user, elasticsearch.WithRefresh(true))
	err = client.CreateBulk(ctx, "users", docs)
	err = client.UpdateDoc(ctx, "users", user)
	err = client.DeleteDoc(ctx, "users", "u1")

索引创建使用 IndexMapping：

	mapping := &elasticsearch.IndexMapping{
		Settings: map[string]any{"number_of_shards": 1},
		Mappings: map[string]any{
			"properties": map[string]any{
				"name": map[string]any{"type": "text"},
			},
		},
	}
	err := client.CreateIndex(ctx, "users_v1", mapping)

# 别名与重建索引

ReindexWithAlias 适合小到中等数据量的同步重建；大索引建议使用
ReindexWithAliasAsync、WaitForTask 和 FinishReindexWithAlias 分步执行。

	newIndex, taskID, err := client.ReindexWithAliasAsync(ctx, "users", mapping)
	if taskID != "" {
		err = client.WaitForTask(ctx, taskID, 5*time.Second)
	}
	err = client.FinishReindexWithAlias(ctx, "users", newIndex)

# 同义词管理

Synonyms API 用于管理 search analyzer 可引用的同义词集合：

	rules := []elasticsearch.SynonymRule{
		{ID: "r1", Synonyms: "phone, mobile"},
		{ID: "r2", Synonyms: "tv, television"},
	}

	_, err := client.PutSynonymSet(ctx, "products", rules)
	set, err := client.GetSynonymSet(ctx, "products")
	rule, err := client.GetSynonymRule(ctx, "products", "r1")
	_ = set
	_ = rule

# 任务管理

Task API 可用于查询 reindex、delete by query 等长任务，也可以取消可取消任务：

	task, err := client.GetTask(ctx, taskID, elasticsearch.WithTaskWaitForCompletion(true))
	tasks, err := client.ListTasks(ctx,
		elasticsearch.WithTaskActions("indices:data/write/reindex"),
		elasticsearch.WithTaskDetailed(true),
	)
	cancel, err := client.CancelTask(ctx, taskID)
	_ = task
	_ = tasks
	_ = cancel

# 生产建议

  - 为每个调用传入带超时的 context，避免请求无限等待。
  - 生产环境不要依赖默认本地地址，显式配置 Addresses 或 CloudID。
  - 写入后需要马上搜索到结果时再使用 WithRefresh(true)，高吞吐写入不建议默认开启。
  - 大批量导入优先使用 CreateBulk，并关注返回错误；部分文档失败会作为错误返回。
  - 重建索引推荐通过别名访问业务索引，避免业务代码绑定物理索引名。
  - 更新 synonyms 后，关联 analyzer 可能会 reload；生产环境应关注返回的 reload 明细。
  - 取消 task 只对 Elasticsearch 标记为 cancellable 的任务有效。
  - 需要官方客户端未封装能力时，可通过 client.Native() 获取原生客户端。

# 测试

普通单元测试不需要 Elasticsearch：

	go test ./elasticsearch

本地集成测试需要启动未开启安全认证的 Elasticsearch，并设置环境变量：

	ELASTICSEARCH_INTEGRATION=1 ELASTICSEARCH_ADDR=http://127.0.0.1:9200 go test -run TestIntegrationLocalElasticsearch -count=1 ./elasticsearch

# 设计说明

这个包直接封装 Elasticsearch 官方客户端，不再拆分 adapter。因为
Elasticsearch 的 gox 封装没有多个等价后端实现，保留额外 adapter 层会让
import path 和命名变得绕。
*/
package elasticsearch
