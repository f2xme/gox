package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// GetTask 获取单个任务信息。
func (c *Client) GetTask(ctx context.Context, taskID string, opts ...TaskOption) (*TaskResponse, error) {
	options := applyTaskOptions(opts...)
	req := esapi.TasksGetRequest{
		TaskID:            taskID,
		Timeout:           options.Timeout,
		WaitForCompletion: options.WaitForCompletion,
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("elastic: get task %s: %w", taskID, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, responseError("get task "+taskID, resp)
	}
	return decodeTaskResponse(resp.Body, "task "+taskID)
}

// ListTasks 获取任务列表。
func (c *Client) ListTasks(ctx context.Context, opts ...TaskOption) (*TaskListResponse, error) {
	options := applyTaskOptions(opts...)
	req := esapi.TasksListRequest{
		Actions:           options.Actions,
		Detailed:          options.Detailed,
		GroupBy:           string(options.GroupBy),
		Nodes:             options.Nodes,
		ParentTaskID:      options.ParentTaskID,
		Timeout:           options.Timeout,
		WaitForCompletion: options.WaitForCompletion,
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("elastic: list tasks: %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, responseError("list tasks", resp)
	}
	return decodeTaskListResponse(resp.Body, "tasks")
}

// CancelTasks 取消匹配条件的任务。
func (c *Client) CancelTasks(ctx context.Context, opts ...TaskOption) (*TaskCancelResponse, error) {
	options := applyTaskOptions(opts...)
	req := esapi.TasksCancelRequest{
		Actions:           options.Actions,
		Nodes:             options.Nodes,
		ParentTaskID:      options.ParentTaskID,
		WaitForCompletion: options.WaitForCompletion,
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("elastic: cancel tasks: %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, responseError("cancel tasks", resp)
	}
	return decodeTaskListResponse(resp.Body, "cancel tasks")
}

// CancelTask 取消指定任务。
func (c *Client) CancelTask(ctx context.Context, taskID string, opts ...TaskOption) (*TaskCancelResponse, error) {
	options := applyTaskOptions(opts...)
	req := esapi.TasksCancelRequest{
		TaskID:            taskID,
		WaitForCompletion: options.WaitForCompletion,
	}

	resp, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("elastic: cancel task %s: %w", taskID, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, responseError("cancel task "+taskID, resp)
	}
	return decodeTaskListResponse(resp.Body, "cancel task "+taskID)
}

func decodeTaskResponse(r io.Reader, label string) (*TaskResponse, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("elastic: read %s response: %w", label, err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("elastic: decode %s response: %w", label, err)
	}

	var result TaskResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("elastic: decode %s typed response: %w", label, err)
	}
	result.Raw = raw
	return &result, nil
}

func decodeTaskListResponse(r io.Reader, label string) (*TaskListResponse, error) {
	var raw map[string]any
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("elastic: read %s response: %w", label, err)
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("elastic: decode %s response: %w", label, err)
	}

	var result TaskListResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("elastic: decode %s typed response: %w", label, err)
	}
	result.Raw = raw
	return &result, nil
}
