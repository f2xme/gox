package elasticsearch

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestTaskMethods(t *testing.T) {
	seen := make(map[string]bool)
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeJSON(w, map[string]any{"name": "test"})
		case r.URL.Path == "/_tasks/node1:10" && r.Method == http.MethodGet:
			seen["get"] = true
			if r.URL.Query().Get("wait_for_completion") != "true" || r.URL.Query().Get("timeout") == "" {
				t.Fatalf("get task query = %s", r.URL.RawQuery)
			}
			writeJSON(w, map[string]any{
				"completed": true,
				"task": map[string]any{
					"node":        "node1",
					"id":          10,
					"action":      "indices:data/write/reindex",
					"cancellable": true,
				},
				"response": map[string]any{"created": 2},
			})
		case r.URL.Path == "/_tasks" && r.Method == http.MethodGet:
			seen["list"] = true
			if r.URL.Query().Get("actions") != "indices:data/write/reindex" ||
				r.URL.Query().Get("detailed") != "true" ||
				r.URL.Query().Get("group_by") != "nodes" ||
				r.URL.Query().Get("nodes") != "node1" ||
				r.URL.Query().Get("parent_task_id") != "node1:1" {
				t.Fatalf("list tasks query = %s", r.URL.RawQuery)
			}
			writeJSON(w, taskListFixture())
		case r.URL.Path == "/_tasks/_cancel" && r.Method == http.MethodPost:
			seen["cancel"] = true
			if r.URL.Query().Get("actions") != "indices:data/write/reindex" ||
				r.URL.Query().Get("wait_for_completion") != "true" {
				t.Fatalf("cancel tasks query = %s", r.URL.RawQuery)
			}
			writeJSON(w, taskListFixture())
		case r.URL.Path == "/_tasks/node1:10/_cancel" && r.Method == http.MethodPost:
			seen["cancel_one"] = true
			writeJSON(w, taskListFixture())
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	client, err := New(WithAddresses(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	task, err := client.GetTask(context.Background(), "node1:10", WithTaskWaitForCompletion(true), WithTaskTimeout(time.Second))
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if !task.Completed || task.Task["action"] != "indices:data/write/reindex" || task.Raw["completed"] != true {
		t.Fatalf("GetTask() = %#v", task)
	}

	list, err := client.ListTasks(
		context.Background(),
		WithTaskActions("indices:data/write/reindex"),
		WithTaskDetailed(true),
		WithTaskGroupBy(TaskGroupByNodes),
		WithTaskNodes("node1"),
		WithTaskParentTaskID("node1:1"),
	)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if list.Nodes["node1"].Tasks["node1:10"].Action != "indices:data/write/reindex" {
		t.Fatalf("ListTasks() = %#v", list)
	}

	cancel, err := client.CancelTasks(context.Background(), WithTaskActions("indices:data/write/reindex"), WithTaskWaitForCompletion(true))
	if err != nil {
		t.Fatalf("CancelTasks() error = %v", err)
	}
	if cancel.Nodes["node1"].Tasks["node1:10"].Action == "" {
		t.Fatalf("CancelTasks() = %#v", cancel)
	}

	if _, err := client.CancelTask(context.Background(), "node1:10", WithTaskWaitForCompletion(true)); err != nil {
		t.Fatalf("CancelTask() error = %v", err)
	}

	for _, key := range []string{"get", "list", "cancel", "cancel_one"} {
		if !seen[key] {
			t.Fatalf("handler %q was not called", key)
		}
	}
}

func TestTaskError(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeJSON(w, map[string]any{"name": "test"})
		case strings.HasPrefix(r.URL.Path, "/_tasks/bad"):
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, map[string]any{"error": "task missing"})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	client, err := New(WithAddresses(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.GetTask(context.Background(), "bad")
	if err == nil || !strings.Contains(err.Error(), "task missing") {
		t.Fatalf("GetTask() error = %v, want task missing", err)
	}
}

func TestTaskOptions(t *testing.T) {
	options := applyTaskOptions(
		WithTaskActions("a"),
		WithTaskDetailed(true),
		WithTaskGroupBy(TaskGroupByNone),
		WithTaskNodes("n1"),
		WithTaskParentTaskID("p1"),
		WithTaskTimeout(-time.Second),
		WithTaskWaitForCompletion(true),
	)
	if options.Timeout != 0 || options.GroupBy != TaskGroupByNone || len(options.Actions) != 1 ||
		options.Detailed == nil || options.WaitForCompletion == nil {
		t.Fatalf("task options not applied as expected: %#v", options)
	}
}

func taskListFixture() map[string]any {
	return map[string]any{
		"nodes": map[string]any{
			"node1": map[string]any{
				"name": "node1",
				"tasks": map[string]any{
					"node1:10": map[string]any{
						"node":        "node1",
						"id":          10,
						"type":        "transport",
						"action":      "indices:data/write/reindex",
						"cancellable": true,
					},
				},
			},
		},
	}
}
