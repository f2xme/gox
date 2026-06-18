package elasticsearch

import "time"

// TaskGroupBy 表示任务列表分组方式。
type TaskGroupBy string

const (
	// TaskGroupByNodes 按节点分组任务。
	TaskGroupByNodes TaskGroupBy = "nodes"
	// TaskGroupByParents 按父子任务关系分组任务。
	TaskGroupByParents TaskGroupBy = "parents"
	// TaskGroupByNone 不分组任务。
	TaskGroupByNone TaskGroupBy = "none"
)

// TaskResponse 表示单个任务查询响应。
type TaskResponse struct {
	// Completed 表示任务是否已完成。
	Completed bool `json:"completed"`
	// Task 任务详情。
	Task map[string]any `json:"task,omitempty"`
	// Response 任务完成后的响应。
	Response map[string]any `json:"response,omitempty"`
	// Error 任务错误信息。
	Error map[string]any `json:"error,omitempty"`
	// Raw 保留完整响应，便于读取 Elasticsearch 新增字段。
	Raw map[string]any `json:"-"`
}

// TaskListResponse 表示任务列表响应。
type TaskListResponse struct {
	// Nodes 按节点分组的任务信息。
	Nodes map[string]TaskNode `json:"nodes,omitempty"`
	// Tasks 不分组时返回的任务信息。
	Tasks map[string]TaskInfo `json:"tasks,omitempty"`
	// NodeFailures 节点失败信息。
	NodeFailures []map[string]any `json:"node_failures,omitempty"`
	// Raw 保留完整响应，便于读取 Elasticsearch 新增字段。
	Raw map[string]any `json:"-"`
}

// TaskCancelResponse 表示取消任务响应。
type TaskCancelResponse = TaskListResponse

// TaskNode 表示节点上的任务信息。
type TaskNode struct {
	// Name 节点名称。
	Name string `json:"name,omitempty"`
	// TransportAddress 节点传输地址。
	TransportAddress string `json:"transport_address,omitempty"`
	// Host 节点主机。
	Host string `json:"host,omitempty"`
	// IP 节点 IP。
	IP string `json:"ip,omitempty"`
	// Roles 节点角色。
	Roles []string `json:"roles,omitempty"`
	// Attributes 节点属性。
	Attributes map[string]string `json:"attributes,omitempty"`
	// Tasks 节点上的任务列表。
	Tasks map[string]TaskInfo `json:"tasks,omitempty"`
}

// TaskInfo 表示任务详情。
type TaskInfo struct {
	// Node 任务所在节点。
	Node string `json:"node,omitempty"`
	// ID 任务数字 ID。
	ID int64 `json:"id,omitempty"`
	// Type 任务类型。
	Type string `json:"type,omitempty"`
	// Action 任务动作。
	Action string `json:"action,omitempty"`
	// Status 任务状态。
	Status map[string]any `json:"status,omitempty"`
	// Description 任务描述。
	Description string `json:"description,omitempty"`
	// StartTimeInMillis 任务开始时间戳。
	StartTimeInMillis int64 `json:"start_time_in_millis,omitempty"`
	// RunningTimeInNanos 任务运行时长。
	RunningTimeInNanos int64 `json:"running_time_in_nanos,omitempty"`
	// Cancellable 表示任务是否可取消。
	Cancellable bool `json:"cancellable,omitempty"`
	// Cancelled 表示任务是否已取消。
	Cancelled bool `json:"cancelled,omitempty"`
	// ParentTaskID 父任务 ID。
	ParentTaskID string `json:"parent_task_id,omitempty"`
	// Headers 任务请求头。
	Headers map[string]string `json:"headers,omitempty"`
	// Children 子任务列表。
	Children []TaskInfo `json:"children,omitempty"`
}

// TaskOptions 定义任务 API 选项。
type TaskOptions struct {
	// Actions 任务动作过滤条件。
	Actions []string
	// Detailed 是否返回详细任务信息。
	Detailed *bool
	// GroupBy 任务分组方式。
	GroupBy TaskGroupBy
	// Nodes 节点过滤条件。
	Nodes []string
	// ParentTaskID 父任务 ID。
	ParentTaskID string
	// Timeout 显式操作超时。
	Timeout time.Duration
	// WaitForCompletion 是否等待任务完成。
	WaitForCompletion *bool
}

// TaskOption 定义任务 API 选项函数。
type TaskOption func(*TaskOptions)

func applyTaskOptions(opts ...TaskOption) TaskOptions {
	var options TaskOptions
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// WithTaskActions 设置任务动作过滤条件。
func WithTaskActions(actions ...string) TaskOption {
	return func(o *TaskOptions) {
		o.Actions = actions
	}
}

// WithTaskDetailed 设置是否返回详细任务信息。
func WithTaskDetailed(detailed bool) TaskOption {
	return func(o *TaskOptions) {
		o.Detailed = &detailed
	}
}

// WithTaskGroupBy 设置任务分组方式。
func WithTaskGroupBy(groupBy TaskGroupBy) TaskOption {
	return func(o *TaskOptions) {
		o.GroupBy = groupBy
	}
}

// WithTaskNodes 设置节点过滤条件。
func WithTaskNodes(nodes ...string) TaskOption {
	return func(o *TaskOptions) {
		o.Nodes = nodes
	}
}

// WithTaskParentTaskID 设置父任务 ID。
func WithTaskParentTaskID(parentTaskID string) TaskOption {
	return func(o *TaskOptions) {
		o.ParentTaskID = parentTaskID
	}
}

// WithTaskTimeout 设置任务 API 显式超时。
func WithTaskTimeout(timeout time.Duration) TaskOption {
	return func(o *TaskOptions) {
		if timeout > 0 {
			o.Timeout = timeout
		}
	}
}

// WithTaskWaitForCompletion 设置是否等待任务完成。
func WithTaskWaitForCompletion(wait bool) TaskOption {
	return func(o *TaskOptions) {
		o.WaitForCompletion = &wait
	}
}
