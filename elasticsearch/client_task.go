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
	callOpts := []func(*esapi.TasksGetRequest){
		c.client.Tasks.Get.WithContext(ctx),
	}
	if options.Timeout > 0 {
		callOpts = append(callOpts, c.client.Tasks.Get.WithTimeout(options.Timeout))
	}
	if options.WaitForCompletion != nil {
		callOpts = append(callOpts, c.client.Tasks.Get.WithWaitForCompletion(*options.WaitForCompletion))
	}

	resp, err := c.client.Tasks.Get(taskID, callOpts...)
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
	callOpts := []func(*esapi.TasksListRequest){
		c.client.Tasks.List.WithContext(ctx),
	}
	if len(options.Actions) > 0 {
		callOpts = append(callOpts, c.client.Tasks.List.WithActions(options.Actions...))
	}
	if options.Detailed != nil {
		callOpts = append(callOpts, c.client.Tasks.List.WithDetailed(*options.Detailed))
	}
	if options.GroupBy != "" {
		callOpts = append(callOpts, c.client.Tasks.List.WithGroupBy(string(options.GroupBy)))
	}
	if len(options.Nodes) > 0 {
		callOpts = append(callOpts, c.client.Tasks.List.WithNodes(options.Nodes...))
	}
	if options.ParentTaskID != "" {
		callOpts = append(callOpts, c.client.Tasks.List.WithParentTaskID(options.ParentTaskID))
	}
	if options.Timeout > 0 {
		callOpts = append(callOpts, c.client.Tasks.List.WithTimeout(options.Timeout))
	}
	if options.WaitForCompletion != nil {
		callOpts = append(callOpts, c.client.Tasks.List.WithWaitForCompletion(*options.WaitForCompletion))
	}

	resp, err := c.client.Tasks.List(callOpts...)
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
	callOpts := []func(*esapi.TasksCancelRequest){
		c.client.Tasks.Cancel.WithContext(ctx),
	}
	if len(options.Actions) > 0 {
		callOpts = append(callOpts, c.client.Tasks.Cancel.WithActions(options.Actions...))
	}
	if len(options.Nodes) > 0 {
		callOpts = append(callOpts, c.client.Tasks.Cancel.WithNodes(options.Nodes...))
	}
	if options.ParentTaskID != "" {
		callOpts = append(callOpts, c.client.Tasks.Cancel.WithParentTaskID(options.ParentTaskID))
	}
	if options.WaitForCompletion != nil {
		callOpts = append(callOpts, c.client.Tasks.Cancel.WithWaitForCompletion(*options.WaitForCompletion))
	}

	resp, err := c.client.Tasks.Cancel(callOpts...)
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
	callOpts := []func(*esapi.TasksCancelRequest){
		c.client.Tasks.Cancel.WithContext(ctx),
		c.client.Tasks.Cancel.WithTaskID(taskID),
	}
	if options.WaitForCompletion != nil {
		callOpts = append(callOpts, c.client.Tasks.Cancel.WithWaitForCompletion(*options.WaitForCompletion))
	}

	resp, err := c.client.Tasks.Cancel(callOpts...)
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
