package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// Analyze 执行 Elasticsearch analyze 请求。
func (c *Client) Analyze(ctx context.Context, index string, body io.Reader) ([]AnalyzeToken, error) {
	resp, err := c.client.Indices.Analyze(
		c.client.Indices.Analyze.WithContext(ctx),
		c.client.Indices.Analyze.WithIndex(index),
		c.client.Indices.Analyze.WithBody(body),
	)
	if err != nil {
		return nil, fmt.Errorf("elastic: analyze %s: %w", index, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, responseError("analyze "+index, resp)
	}

	var result AnalyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("elastic: decode analyze response: %w", err)
	}
	return result.Tokens, nil
}
