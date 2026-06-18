package elasticsearch

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func (c *Client) searchRaw(ctx context.Context, req Request) (*esapi.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("elastic: search request is nil")
	}
	resp, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(req.Index()),
		c.client.Search.WithBody(req.Body()),
	)
	if err != nil {
		return nil, fmt.Errorf("elastic: search %s: %w", req.Index(), err)
	}
	if resp.IsError() {
		defer resp.Body.Close()
		return nil, responseError("search "+req.Index(), resp)
	}
	return resp, nil
}

// Search 执行搜索并返回 map 版本结果。
func (c *Client) Search(ctx context.Context, req Request) (*SearchResult[*HitMap], error) {
	resp, err := c.searchRaw(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return DecodeSearchHitMaps(resp.Body)
}

// SearchWithType 执行搜索并返回泛型 _source 结果。
func SearchWithType[T any](ctx context.Context, c *Client, req Request) (*SearchResult[T], error) {
	resp, err := c.searchRaw(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return DecodeSearchSources[T](resp.Body)
}
