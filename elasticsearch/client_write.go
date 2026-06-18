package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

// CreateDoc 创建或覆盖文档。
func (c *Client) CreateDoc(ctx context.Context, index string, doc Document, opts ...WriteOption) error {
	if doc == nil {
		return fmt.Errorf("elastic: document is nil")
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("elastic: marshal document %s/%s: %w", index, doc.ID(), err)
	}

	writeOptions := ApplyWriteOptions(opts...)
	callOpts := []func(*esapi.IndexRequest){
		c.client.Index.WithContext(ctx),
		c.client.Index.WithDocumentID(doc.ID()),
	}
	if writeOptions.Refresh {
		callOpts = append(callOpts, c.client.Index.WithRefresh("true"))
	}

	resp, err := c.client.Index(index, bytes.NewReader(body), callOpts...)
	if err != nil {
		return fmt.Errorf("elastic: create document %s/%s: %w", index, doc.ID(), err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return responseError("create document "+index+"/"+doc.ID(), resp)
	}
	return nil
}

// CreateBulk 批量创建或覆盖文档。
func (c *Client) CreateBulk(ctx context.Context, index string, docs []Document, opts ...WriteOption) error {
	if len(docs) == 0 {
		return nil
	}
	for _, doc := range docs {
		if doc == nil {
			return fmt.Errorf("elastic: bulk document is nil")
		}
	}

	writeOptions := ApplyWriteOptions(opts...)
	cfg := esutil.BulkIndexerConfig{
		Index:  index,
		Client: c.client,
	}
	if writeOptions.Refresh {
		cfg.Refresh = "true"
	}
	bi, err := esutil.NewBulkIndexer(cfg)
	if err != nil {
		return fmt.Errorf("elastic: create bulk indexer: %w", err)
	}

	var mu sync.Mutex
	var bulkErr error
	setBulkErr := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		if bulkErr == nil {
			bulkErr = err
		}
	}

	for _, doc := range docs {
		body, err := json.Marshal(doc)
		if err != nil {
			_ = bi.Close(ctx)
			return fmt.Errorf("elastic: marshal bulk document %s/%s: %w", index, doc.ID(), err)
		}
		if err := bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: doc.ID(),
			Body:       bytes.NewReader(body),
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, resp esutil.BulkIndexerResponseItem, err error) {
				if err != nil {
					setBulkErr(fmt.Errorf("bulk item %s failed: %w", item.DocumentID, err))
					return
				}
				setBulkErr(fmt.Errorf("bulk item %s failed with status %d: %s", item.DocumentID, resp.Status, resp.Error.Reason))
			},
		}); err != nil {
			return fmt.Errorf("elastic: add bulk document %s/%s: %w", index, doc.ID(), err)
		}
	}

	if err := bi.Close(ctx); err != nil {
		return fmt.Errorf("elastic: close bulk indexer: %w", err)
	}
	if bulkErr != nil {
		return fmt.Errorf("elastic: bulk create %s: %w", index, bulkErr)
	}
	return nil
}

// UpdateDoc 局部更新文档。
func (c *Client) UpdateDoc(ctx context.Context, index string, doc Document, opts ...WriteOption) error {
	if doc == nil {
		return fmt.Errorf("elastic: document is nil")
	}
	body, err := json.Marshal(map[string]any{"doc": doc})
	if err != nil {
		return fmt.Errorf("elastic: marshal update document %s/%s: %w", index, doc.ID(), err)
	}

	callOpts := []func(*esapi.UpdateRequest){
		c.client.Update.WithContext(ctx),
	}
	if ApplyWriteOptions(opts...).Refresh {
		callOpts = append(callOpts, c.client.Update.WithRefresh("true"))
	}
	resp, err := c.client.Update(index, doc.ID(), bytes.NewReader(body), callOpts...)
	if err != nil {
		return fmt.Errorf("elastic: update document %s/%s: %w", index, doc.ID(), err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return responseError("update document "+index+"/"+doc.ID(), resp)
	}
	return nil
}

// DeleteDoc 删除文档。
func (c *Client) DeleteDoc(ctx context.Context, index string, id string, opts ...WriteOption) error {
	callOpts := []func(*esapi.DeleteRequest){
		c.client.Delete.WithContext(ctx),
	}
	if ApplyWriteOptions(opts...).Refresh {
		callOpts = append(callOpts, c.client.Delete.WithRefresh("true"))
	}
	resp, err := c.client.Delete(index, id, callOpts...)
	if err != nil {
		return fmt.Errorf("elastic: delete document %s/%s: %w", index, id, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return responseError("delete document "+index+"/"+id, resp)
	}
	return nil
}
