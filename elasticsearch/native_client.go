package elasticsearch

import (
	"context"
	"fmt"

	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Client 是基于官方 go-elasticsearch/v8 的 Elasticsearch 客户端。
type Client struct {
	client *elastic.BaseClient
}

// Native 返回不附带全量 API 的底层 Elasticsearch 客户端。
func (c *Client) Native() *elastic.BaseClient {
	return c.client
}

// Unwrap 返回不附带全量 API 的底层 Elasticsearch 客户端。
func (c *Client) Unwrap() any {
	return c.client
}

// Ping 检查 Elasticsearch 连通性。
func (c *Client) Ping(ctx context.Context) error {
	resp, err := (esapi.InfoRequest{}).Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("elastic: ping: %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return responseError("ping", resp)
	}
	return nil
}

var _ Searcher = (*Client)(nil)
var _ Counter = (*Client)(nil)
var _ Analyzer = (*Client)(nil)
var _ SynonymManager = (*Client)(nil)
var _ TaskManager = (*Client)(nil)
var _ IndexManager = (*Client)(nil)
var _ Writer = (*Client)(nil)
