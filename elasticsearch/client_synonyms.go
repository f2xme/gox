package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ListSynonymSets 获取同义词集合列表。
func (c *Client) ListSynonymSets(ctx context.Context, opts ...SynonymOption) (*SynonymSetList, error) {
	options := applySynonymOptions(opts...)
	callOpts := []func(*esapi.SynonymsGetSynonymsSetsRequest){
		c.client.SynonymsGetSynonymsSets.WithContext(ctx),
	}
	if options.From != nil {
		callOpts = append(callOpts, c.client.SynonymsGetSynonymsSets.WithFrom(*options.From))
	}
	if options.Size != nil {
		callOpts = append(callOpts, c.client.SynonymsGetSynonymsSets.WithSize(*options.Size))
	}

	resp, err := c.client.SynonymsGetSynonymsSets(callOpts...)
	if err != nil {
		return nil, fmt.Errorf("elastic: list synonym sets: %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, responseError("list synonym sets", resp)
	}

	var result SynonymSetList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("elastic: decode synonym sets response: %w", err)
	}
	return &result, nil
}

// GetSynonymSet 获取同义词集合。
func (c *Client) GetSynonymSet(ctx context.Context, id string, opts ...SynonymOption) (*SynonymSet, error) {
	options := applySynonymOptions(opts...)
	callOpts := []func(*esapi.SynonymsGetSynonymRequest){
		c.client.SynonymsGetSynonym.WithContext(ctx),
	}
	if options.From != nil {
		callOpts = append(callOpts, c.client.SynonymsGetSynonym.WithFrom(*options.From))
	}
	if options.Size != nil {
		callOpts = append(callOpts, c.client.SynonymsGetSynonym.WithSize(*options.Size))
	}

	resp, err := c.client.SynonymsGetSynonym(id, callOpts...)
	if err != nil {
		return nil, fmt.Errorf("elastic: get synonym set %s: %w", id, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, responseError("get synonym set "+id, resp)
	}

	var result SynonymSet
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("elastic: decode synonym set %s response: %w", id, err)
	}
	return &result, nil
}

// PutSynonymSet 创建或更新同义词集合。
func (c *Client) PutSynonymSet(ctx context.Context, id string, rules []SynonymRule) (*SynonymUpdateResult, error) {
	body, err := json.Marshal(map[string]any{"synonyms_set": rules})
	if err != nil {
		return nil, fmt.Errorf("elastic: marshal synonym set %s: %w", id, err)
	}

	resp, err := c.client.SynonymsPutSynonym(
		id,
		bytes.NewReader(body),
		c.client.SynonymsPutSynonym.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("elastic: put synonym set %s: %w", id, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, responseError("put synonym set "+id, resp)
	}
	return decodeSynonymUpdateResult(resp.Body, "synonym set "+id)
}

// DeleteSynonymSet 删除同义词集合。
func (c *Client) DeleteSynonymSet(ctx context.Context, id string) error {
	resp, err := c.client.SynonymsDeleteSynonym(
		id,
		c.client.SynonymsDeleteSynonym.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("elastic: delete synonym set %s: %w", id, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return responseError("delete synonym set "+id, resp)
	}
	return nil
}

// GetSynonymRule 获取同义词集合中的单条规则。
func (c *Client) GetSynonymRule(ctx context.Context, setID, ruleID string) (*SynonymRule, error) {
	resp, err := c.client.SynonymsGetSynonymRule(
		ruleID,
		setID,
		c.client.SynonymsGetSynonymRule.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("elastic: get synonym rule %s/%s: %w", setID, ruleID, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, responseError("get synonym rule "+setID+"/"+ruleID, resp)
	}

	var result SynonymRule
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("elastic: decode synonym rule %s/%s response: %w", setID, ruleID, err)
	}
	return &result, nil
}

// PutSynonymRule 创建或更新同义词集合中的单条规则。
func (c *Client) PutSynonymRule(ctx context.Context, setID, ruleID, synonyms string) (*SynonymUpdateResult, error) {
	body, err := json.Marshal(SynonymRule{Synonyms: synonyms})
	if err != nil {
		return nil, fmt.Errorf("elastic: marshal synonym rule %s/%s: %w", setID, ruleID, err)
	}

	resp, err := c.client.SynonymsPutSynonymRule(
		bytes.NewReader(body),
		ruleID,
		setID,
		c.client.SynonymsPutSynonymRule.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("elastic: put synonym rule %s/%s: %w", setID, ruleID, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, responseError("put synonym rule "+setID+"/"+ruleID, resp)
	}
	return decodeSynonymUpdateResult(resp.Body, "synonym rule "+setID+"/"+ruleID)
}

// DeleteSynonymRule 删除同义词集合中的单条规则。
func (c *Client) DeleteSynonymRule(ctx context.Context, setID, ruleID string) error {
	resp, err := c.client.SynonymsDeleteSynonymRule(
		ruleID,
		setID,
		c.client.SynonymsDeleteSynonymRule.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("elastic: delete synonym rule %s/%s: %w", setID, ruleID, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return responseError("delete synonym rule "+setID+"/"+ruleID, resp)
	}
	return nil
}

func decodeSynonymUpdateResult(r io.Reader, label string) (*SynonymUpdateResult, error) {
	var raw map[string]any
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, fmt.Errorf("elastic: decode %s response: %w", label, err)
	}

	result := &SynonymUpdateResult{Raw: raw}
	if acknowledged, ok := raw["acknowledged"].(bool); ok {
		result.Acknowledged = acknowledged
	}
	if details, ok := raw["reload_analyzers_details"].([]any); ok {
		result.ReloadAnalyzersDetails = make([]map[string]any, 0, len(details))
		for _, detail := range details {
			if m, ok := detail.(map[string]any); ok {
				result.ReloadAnalyzersDetails = append(result.ReloadAnalyzersDetails, m)
			}
		}
	}
	return result, nil
}
