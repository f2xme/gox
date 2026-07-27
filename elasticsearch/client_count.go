package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Count 统计匹配文档数量。
func (c *Client) Count(ctx context.Context, req Request) (int64, error) {
	if req == nil {
		return 0, fmt.Errorf("elastic: count request is nil")
	}
	body, err := countBody(req)
	if err != nil {
		return 0, err
	}
	resp, err := (esapi.CountRequest{
		Index: []string{req.Index()},
		Body:  body,
	}).Do(ctx, c.client)
	if err != nil {
		return 0, fmt.Errorf("elastic: count %s: %w", req.Index(), err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return 0, responseError("count "+req.Index(), resp)
	}
	return DecodeCountResponse(resp.Body)
}

func countBody(req Request) (io.Reader, error) {
	data, err := io.ReadAll(req.Body())
	if err != nil {
		return nil, fmt.Errorf("elastic: read count request body: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return bytes.NewReader(nil), nil
	}

	var request map[string]any
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, fmt.Errorf("elastic: decode count request body: %w", err)
	}

	query, ok := request["query"]
	if !ok {
		return bytes.NewReader([]byte(`{}`)), nil
	}
	body, err := json.Marshal(map[string]any{"query": query})
	if err != nil {
		return nil, fmt.Errorf("elastic: encode count request body: %w", err)
	}
	return bytes.NewReader(body), nil
}
