package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// CreateIndex 创建索引。
func (c *Client) CreateIndex(ctx context.Context, index string, mapping *IndexMapping) error {
	return c.CreateIndexWithMapping(ctx, index, mapping)
}

// CreateIndexWithMapping 创建带 settings 和 mappings 的索引。
func (c *Client) CreateIndexWithMapping(ctx context.Context, index string, mapping *IndexMapping) error {
	opts := []func(*esapi.IndicesCreateRequest){
		c.client.Indices.Create.WithContext(ctx),
	}
	if mapping != nil {
		body, err := json.Marshal(mapping)
		if err != nil {
			return fmt.Errorf("elastic: marshal index mapping %s: %w", index, err)
		}
		opts = append(opts, c.client.Indices.Create.WithBody(bytes.NewReader(body)))
	}

	resp, err := c.client.Indices.Create(index, opts...)
	if err != nil {
		return fmt.Errorf("elastic: create index %s: %w", index, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return responseError("create index "+index, resp)
	}
	return nil
}

// DeleteIndex 删除索引。
func (c *Client) DeleteIndex(ctx context.Context, index string) error {
	resp, err := c.client.Indices.Delete([]string{index}, c.client.Indices.Delete.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("elastic: delete index %s: %w", index, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return responseError("delete index "+index, resp)
	}
	return nil
}

// IndexExists 判断索引是否存在。
func (c *Client) IndexExists(ctx context.Context, index string) (bool, error) {
	resp, err := c.client.Indices.Exists([]string{index}, c.client.Indices.Exists.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("elastic: check index exists %s: %w", index, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return false, nil
	}
	if resp.IsError() {
		return false, responseError("check index exists "+index, resp)
	}
	return true, nil
}

// RefreshIndex 刷新索引。
func (c *Client) RefreshIndex(ctx context.Context, index string) error {
	resp, err := c.client.Indices.Refresh(
		c.client.Indices.Refresh.WithContext(ctx),
		c.client.Indices.Refresh.WithIndex(index),
	)
	if err != nil {
		return fmt.Errorf("elastic: refresh index %s: %w", index, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return responseError("refresh index "+index, resp)
	}
	return nil
}

// GetAliasIndices 获取别名指向的索引列表。
func (c *Client) GetAliasIndices(ctx context.Context, alias string) ([]string, error) {
	resp, err := c.client.Indices.GetAlias(
		c.client.Indices.GetAlias.WithContext(ctx),
		c.client.Indices.GetAlias.WithName(alias),
	)
	if err != nil {
		return nil, fmt.Errorf("elastic: get alias %s: %w", alias, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil, nil
	}
	if resp.IsError() {
		return nil, responseError("get alias "+alias, resp)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("elastic: decode alias %s response: %w", alias, err)
	}

	indices := make([]string, 0, len(result))
	for index := range result {
		indices = append(indices, index)
	}
	return indices, nil
}

// UpdateAlias 原子更新别名。
func (c *Client) UpdateAlias(ctx context.Context, alias string, oldIndices []string, newIndex string) error {
	actions := make([]AliasAction, 0, len(oldIndices)+1)
	for _, oldIndex := range oldIndices {
		actions = append(actions, AliasAction{
			Remove: &AliasActionItem{Index: oldIndex, Alias: alias},
		})
	}
	actions = append(actions, AliasAction{
		Add: &AliasActionItem{Index: newIndex, Alias: alias},
	})

	return c.updateAliases(ctx, actions)
}

// AddAlias 添加别名。
func (c *Client) AddAlias(ctx context.Context, index, alias string) error {
	return c.updateAliases(ctx, []AliasAction{
		{Add: &AliasActionItem{Index: index, Alias: alias}},
	})
}

func (c *Client) updateAliases(ctx context.Context, actions []AliasAction) error {
	body, err := json.Marshal(map[string]any{"actions": actions})
	if err != nil {
		return fmt.Errorf("elastic: marshal alias actions: %w", err)
	}

	resp, err := c.client.Indices.UpdateAliases(
		bytes.NewReader(body),
		c.client.Indices.UpdateAliases.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("elastic: update aliases: %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return responseError("update aliases", resp)
	}
	return nil
}

// Reindex 将数据从源索引复制到目标索引。
func (c *Client) Reindex(ctx context.Context, sourceIndex, destIndex string) error {
	req := ReindexRequest{
		Source: ReindexSource{Index: sourceIndex},
		Dest:   ReindexDest{Index: destIndex},
	}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("elastic: marshal reindex request: %w", err)
	}

	resp, err := c.client.Reindex(
		bytes.NewReader(body),
		c.client.Reindex.WithContext(ctx),
		c.client.Reindex.WithWaitForCompletion(true),
		c.client.Reindex.WithRefresh(true),
	)
	if err != nil {
		return fmt.Errorf("elastic: reindex %s to %s: %w", sourceIndex, destIndex, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return responseError("reindex "+sourceIndex+" to "+destIndex, resp)
	}
	return nil
}

// ReindexAsync 异步执行 reindex 并返回任务 ID。
func (c *Client) ReindexAsync(ctx context.Context, sourceIndex, destIndex string) (string, error) {
	req := ReindexRequest{
		Source: ReindexSource{Index: sourceIndex},
		Dest:   ReindexDest{Index: destIndex},
	}
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("elastic: marshal async reindex request: %w", err)
	}

	resp, err := c.client.Reindex(
		bytes.NewReader(body),
		c.client.Reindex.WithContext(ctx),
		c.client.Reindex.WithWaitForCompletion(false),
	)
	if err != nil {
		return "", fmt.Errorf("elastic: async reindex %s to %s: %w", sourceIndex, destIndex, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return "", responseError("async reindex "+sourceIndex+" to "+destIndex, resp)
	}

	var result struct {
		Task string `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("elastic: decode async reindex response: %w", err)
	}
	return result.Task, nil
}

// GetTaskStatus 获取任务是否完成。
func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (bool, error) {
	resp, err := c.client.Tasks.Get(taskID, c.client.Tasks.Get.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("elastic: get task %s: %w", taskID, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return false, responseError("get task "+taskID, resp)
	}

	var result struct {
		Completed bool `json:"completed"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("elastic: decode task %s response: %w", taskID, err)
	}
	return result.Completed, nil
}

// WaitForTask 等待任务完成。
func (c *Client) WaitForTask(ctx context.Context, taskID string, checkInterval time.Duration) error {
	if checkInterval <= 0 {
		checkInterval = time.Second
	}
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("elastic: wait task %s: %w", taskID, ctx.Err())
		case <-ticker.C:
			completed, err := c.GetTaskStatus(ctx, taskID)
			if err != nil {
				return err
			}
			if completed {
				return nil
			}
		}
	}
}

// GetIndexSettings 获取索引 settings。
func (c *Client) GetIndexSettings(ctx context.Context, index string) (map[string]any, error) {
	resp, err := c.client.Indices.GetSettings(
		c.client.Indices.GetSettings.WithContext(ctx),
		c.client.Indices.GetSettings.WithIndex(index),
	)
	if err != nil {
		return nil, fmt.Errorf("elastic: get index settings %s: %w", index, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, responseError("get index settings "+index, resp)
	}
	return decodeMap(resp.Body, "settings "+index)
}

// GetIndexMapping 获取索引 mappings。
func (c *Client) GetIndexMapping(ctx context.Context, index string) (map[string]any, error) {
	resp, err := c.client.Indices.GetMapping(
		c.client.Indices.GetMapping.WithContext(ctx),
		c.client.Indices.GetMapping.WithIndex(index),
	)
	if err != nil {
		return nil, fmt.Errorf("elastic: get index mapping %s: %w", index, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, responseError("get index mapping "+index, resp)
	}
	return decodeMap(resp.Body, "mapping "+index)
}

// GetIndexConfig 获取索引 settings 和 mappings。
func (c *Client) GetIndexConfig(ctx context.Context, index string) (*IndexMapping, error) {
	settingsResp, err := c.GetIndexSettings(ctx, index)
	if err != nil {
		return nil, err
	}
	mappingsResp, err := c.GetIndexMapping(ctx, index)
	if err != nil {
		return nil, err
	}

	config := &IndexMapping{}
	if indexData, ok := settingsResp[index].(map[string]any); ok {
		if settings, ok := indexData["settings"].(map[string]any); ok {
			if indexSettings, ok := settings["index"].(map[string]any); ok {
				filtered := FilterIndexSettings(indexSettings)
				if len(filtered) > 0 {
					config.Settings = map[string]any{"index": filtered}
				}
			}
		}
	}
	if indexData, ok := mappingsResp[index].(map[string]any); ok {
		if mappings, ok := indexData["mappings"].(map[string]any); ok {
			config.Mappings = mappings
		}
	}
	return config, nil
}

// ReindexWithAlias 创建新索引、复制旧索引数据、切换别名并删除旧索引。
func (c *Client) ReindexWithAlias(ctx context.Context, alias string, mapping *IndexMapping) error {
	oldIndices, err := c.GetAliasIndices(ctx, alias)
	if err != nil {
		return err
	}
	if mapping == nil && len(oldIndices) > 0 {
		mapping, err = c.GetIndexConfig(ctx, oldIndices[0])
		if err != nil {
			return err
		}
	}

	newIndex := GenerateVersionedIndex(alias)
	if err := c.CreateIndex(ctx, newIndex, mapping); err != nil {
		return err
	}
	if len(oldIndices) > 0 {
		if err := c.Reindex(ctx, oldIndices[0], newIndex); err != nil {
			_ = c.DeleteIndex(ctx, newIndex)
			return err
		}
	}
	if err := c.UpdateAlias(ctx, alias, oldIndices, newIndex); err != nil {
		_ = c.DeleteIndex(ctx, newIndex)
		return err
	}
	for _, oldIndex := range oldIndices {
		_ = c.DeleteIndex(ctx, oldIndex)
	}
	return nil
}

// ReindexWithAliasAsync 创建新索引并异步复制旧索引数据。
func (c *Client) ReindexWithAliasAsync(ctx context.Context, alias string, mapping *IndexMapping) (newIndex, taskID string, err error) {
	oldIndices, err := c.GetAliasIndices(ctx, alias)
	if err != nil {
		return "", "", err
	}
	if mapping == nil && len(oldIndices) > 0 {
		mapping, err = c.GetIndexConfig(ctx, oldIndices[0])
		if err != nil {
			return "", "", err
		}
	}

	newIndex = GenerateVersionedIndex(alias)
	if err := c.CreateIndex(ctx, newIndex, mapping); err != nil {
		return "", "", err
	}
	if len(oldIndices) > 0 {
		taskID, err = c.ReindexAsync(ctx, oldIndices[0], newIndex)
		if err != nil {
			_ = c.DeleteIndex(ctx, newIndex)
			return "", "", err
		}
	}
	return newIndex, taskID, nil
}

// FinishReindexWithAlias 完成异步 reindex 后的别名切换。
func (c *Client) FinishReindexWithAlias(ctx context.Context, alias, newIndex string) error {
	oldIndices, err := c.GetAliasIndices(ctx, alias)
	if err != nil {
		return err
	}
	if err := c.UpdateAlias(ctx, alias, oldIndices, newIndex); err != nil {
		return err
	}
	for _, oldIndex := range oldIndices {
		_ = c.DeleteIndex(ctx, oldIndex)
	}
	return nil
}

func decodeMap(r io.Reader, label string) (map[string]any, error) {
	var result map[string]any
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return nil, fmt.Errorf("elastic: decode %s response: %w", label, err)
	}
	return result, nil
}
