package elasticsearch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testDoc struct {
	IDValue string `json:"id"`
	Name    string `json:"name"`
}

func (d testDoc) ID() string {
	return d.IDValue
}

func TestSearchWithType(t *testing.T) {
	var searchBody map[string]any
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeJSON(w, map[string]any{"name": "test"})
		case r.URL.Path == "/users/_search":
			if err := json.NewDecoder(r.Body).Decode(&searchBody); err != nil {
				t.Fatalf("decode search body: %v", err)
			}
			writeJSON(w, map[string]any{
				"hits": map[string]any{
					"total": map[string]any{"value": 1, "relation": "eq"},
					"hits": []map[string]any{
						{
							"_index":  "users",
							"_id":     "1",
							"_score":  1.0,
							"_source": map[string]any{"id": "1", "name": "Alice"},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	client, err := New(WithAddresses(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := NewBuilder("users").Term("status", 1).Pager(1, 10)
	result, err := SearchWithType[testDoc](context.Background(), client, req)
	if err != nil {
		t.Fatalf("SearchWithType() error = %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("Total = %d, want 1", result.Total)
	}
	if result.Hits[0].Name != "Alice" {
		t.Fatalf("Name = %q, want Alice", result.Hits[0].Name)
	}
	if _, ok := searchBody["query"]; !ok {
		t.Fatalf("search body missing query: %#v", searchBody)
	}
}

func TestCount(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeJSON(w, map[string]any{"name": "test"})
		case r.URL.Path == "/users/_count":
			writeJSON(w, map[string]any{"count": 3})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	client, err := New(WithAddresses(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	count, err := client.Count(context.Background(), NewBuilder("users").Term("status", 1))
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 3 {
		t.Fatalf("count = %d, want 3", count)
	}
}

func TestCreateDocWithRefresh(t *testing.T) {
	var gotRefresh string
	var gotBody testDoc
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeJSON(w, map[string]any{"name": "test"})
		case strings.HasPrefix(r.URL.Path, "/users/_doc/1"):
			gotRefresh = r.URL.Query().Get("refresh")
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatalf("decode doc body: %v", err)
			}
			writeJSON(w, map[string]any{"result": "created"})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	client, err := New(WithAddresses(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = client.CreateDoc(context.Background(), "users", testDoc{IDValue: "1", Name: "Alice"}, WithRefresh(true))
	if err != nil {
		t.Fatalf("CreateDoc() error = %v", err)
	}
	if gotRefresh != "true" {
		t.Fatalf("refresh = %q, want true", gotRefresh)
	}
	if gotBody.Name != "Alice" {
		t.Fatalf("Name = %q, want Alice", gotBody.Name)
	}
}

func TestSearchMapAnalyzeAndWriteMethods(t *testing.T) {
	seen := make(map[string]bool)
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeJSON(w, map[string]any{"name": "test"})
		case r.URL.Path == "/users/_search":
			seen["search"] = true
			writeSearchFixture(w)
		case r.URL.Path == "/users/_analyze":
			seen["analyze"] = true
			writeJSON(w, map[string]any{
				"tokens": []map[string]any{
					{"token": "quick", "start_offset": 0, "end_offset": 5, "type": "<ALPHANUM>", "position": 0},
				},
			})
		case strings.HasPrefix(r.URL.Path, "/users/_update/1"):
			seen["update"] = true
			if r.URL.Query().Get("refresh") != "true" {
				t.Fatalf("update refresh = %q, want true", r.URL.Query().Get("refresh"))
			}
			body := decodeJSONMap(t, r.Body)
			if _, ok := body["doc"]; !ok {
				t.Fatalf("update body missing doc: %#v", body)
			}
			writeJSON(w, map[string]any{"result": "updated"})
		case strings.HasPrefix(r.URL.Path, "/users/_doc/1"):
			seen["delete"] = true
			if r.Method != http.MethodDelete {
				t.Fatalf("delete method = %s, want DELETE", r.Method)
			}
			writeJSON(w, map[string]any{"result": "deleted"})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	client, err := New(WithAddresses(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	searchResult, err := client.Search(context.Background(), NewBuilder("users").Term("status", 1))
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if searchResult.Total != 1 || searchResult.Hits[0].GetString("name") != "Alice" {
		t.Fatalf("Search() = %#v", searchResult)
	}
	if len(searchResult.Hits[0].Sort) != 2 || len(searchResult.Hits[0].Highlight["name"]) != 1 {
		t.Fatalf("Search() metadata = %#v", searchResult.Hits[0])
	}

	tokens, err := client.Analyze(context.Background(), "users", strings.NewReader(`{"text":"quick"}`))
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(tokens) != 1 || tokens[0].Token != "quick" {
		t.Fatalf("Analyze() tokens = %#v", tokens)
	}

	if err := client.UpdateDoc(context.Background(), "users", testDoc{IDValue: "1", Name: "Alice"}, WithRefresh(true)); err != nil {
		t.Fatalf("UpdateDoc() error = %v", err)
	}
	if err := client.DeleteDoc(context.Background(), "users", "1", WithRefresh(true)); err != nil {
		t.Fatalf("DeleteDoc() error = %v", err)
	}

	for _, key := range []string{"search", "analyze", "update", "delete"} {
		if !seen[key] {
			t.Fatalf("handler %q was not called", key)
		}
	}
}

func TestIndexManagementMethods(t *testing.T) {
	seen := make(map[string]bool)
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeJSON(w, map[string]any{"name": "test"})
		case r.URL.Path == "/users" && r.Method == http.MethodPut:
			seen["create"] = true
			body := decodeJSONMap(t, r.Body)
			if _, ok := body["mappings"]; !ok {
				t.Fatalf("create index body missing mappings: %#v", body)
			}
			writeJSON(w, map[string]any{"acknowledged": true})
		case r.URL.Path == "/missing" && r.Method == http.MethodHead:
			w.WriteHeader(http.StatusNotFound)
		case r.URL.Path == "/users" && r.Method == http.MethodHead:
			seen["exists"] = true
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/users/_refresh":
			seen["refresh"] = true
			writeJSON(w, map[string]any{"_shards": map[string]any{"successful": 1}})
		case r.URL.Path == "/_alias/alias-users":
			seen["alias"] = true
			writeJSON(w, map[string]any{"users": map[string]any{"aliases": map[string]any{"alias-users": map[string]any{}}}})
		case r.URL.Path == "/_aliases":
			seen["update_aliases"] = true
			body := decodeJSONMap(t, r.Body)
			actions, ok := body["actions"].([]any)
			if !ok || len(actions) == 0 {
				t.Fatalf("alias body actions = %#v", body["actions"])
			}
			writeJSON(w, map[string]any{"acknowledged": true})
		case r.URL.Path == "/users/_settings":
			seen["settings"] = true
			writeJSON(w, map[string]any{
				"users": map[string]any{
					"settings": map[string]any{
						"index": map[string]any{"number_of_shards": "1", "uuid": "ignored"},
					},
				},
			})
		case r.URL.Path == "/users/_mapping":
			seen["mapping"] = true
			writeJSON(w, map[string]any{
				"users": map[string]any{
					"mappings": map[string]any{"properties": map[string]any{"name": map[string]any{"type": "text"}}},
				},
			})
		case r.URL.Path == "/users" && r.Method == http.MethodDelete:
			seen["delete"] = true
			writeJSON(w, map[string]any{"acknowledged": true})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	client, err := New(WithAddresses(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	mapping := &IndexMapping{Mappings: map[string]any{"properties": map[string]any{"name": map[string]any{"type": "text"}}}}
	if err := client.CreateIndex(context.Background(), "users", mapping); err != nil {
		t.Fatalf("CreateIndex() error = %v", err)
	}
	exists, err := client.IndexExists(context.Background(), "users")
	if err != nil || !exists {
		t.Fatalf("IndexExists(users) = %v, %v; want true, nil", exists, err)
	}
	exists, err = client.IndexExists(context.Background(), "missing")
	if err != nil || exists {
		t.Fatalf("IndexExists(missing) = %v, %v; want false, nil", exists, err)
	}
	if err := client.RefreshIndex(context.Background(), "users"); err != nil {
		t.Fatalf("RefreshIndex() error = %v", err)
	}
	indices, err := client.GetAliasIndices(context.Background(), "alias-users")
	if err != nil {
		t.Fatalf("GetAliasIndices() error = %v", err)
	}
	if len(indices) != 1 || indices[0] != "users" {
		t.Fatalf("alias indices = %#v, want [users]", indices)
	}
	if err := client.AddAlias(context.Background(), "users", "alias-users"); err != nil {
		t.Fatalf("AddAlias() error = %v", err)
	}
	if err := client.UpdateAlias(context.Background(), "alias-users", []string{"old-users"}, "users"); err != nil {
		t.Fatalf("UpdateAlias() error = %v", err)
	}
	cfg, err := client.GetIndexConfig(context.Background(), "users")
	if err != nil {
		t.Fatalf("GetIndexConfig() error = %v", err)
	}
	if cfg.Settings["index"].(map[string]any)["uuid"] != nil {
		t.Fatalf("GetIndexConfig() did not filter uuid: %#v", cfg.Settings)
	}
	if err := client.DeleteIndex(context.Background(), "users"); err != nil {
		t.Fatalf("DeleteIndex() error = %v", err)
	}

	for _, key := range []string{"create", "exists", "refresh", "alias", "update_aliases", "settings", "mapping", "delete"} {
		if !seen[key] {
			t.Fatalf("handler %q was not called", key)
		}
	}
}

func TestReindexAndTaskMethods(t *testing.T) {
	seen := make(map[string]bool)
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeJSON(w, map[string]any{"name": "test"})
		case r.URL.Path == "/_reindex":
			body := decodeJSONMap(t, r.Body)
			if body["source"] == nil || body["dest"] == nil {
				t.Fatalf("reindex body = %#v", body)
			}
			if r.URL.Query().Get("wait_for_completion") == "false" {
				seen["reindex_async"] = true
				writeJSON(w, map[string]any{"task": "node:123"})
				return
			}
			seen["reindex"] = true
			writeJSON(w, map[string]any{"took": 1})
		case r.URL.Path == "/_tasks/node:123":
			seen["task"] = true
			writeJSON(w, map[string]any{"completed": true})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	client, err := New(WithAddresses(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := client.Reindex(context.Background(), "old-users", "new-users"); err != nil {
		t.Fatalf("Reindex() error = %v", err)
	}
	taskID, err := client.ReindexAsync(context.Background(), "old-users", "new-users")
	if err != nil {
		t.Fatalf("ReindexAsync() error = %v", err)
	}
	if taskID != "node:123" {
		t.Fatalf("taskID = %q, want node:123", taskID)
	}
	done, err := client.GetTaskStatus(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetTaskStatus() error = %v", err)
	}
	if !done {
		t.Fatalf("GetTaskStatus() = false, want true")
	}
	if err := client.WaitForTask(context.Background(), taskID, time.Millisecond); err != nil {
		t.Fatalf("WaitForTask() error = %v", err)
	}

	for _, key := range []string{"reindex", "reindex_async", "task"} {
		if !seen[key] {
			t.Fatalf("handler %q was not called", key)
		}
	}
}

func TestReindexWithAliasWorkflows(t *testing.T) {
	requests := make(map[string]int)
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		requests[r.Method+" "+r.URL.Path]++
		switch {
		case r.URL.Path == "/":
			writeJSON(w, map[string]any{"name": "test"})
		case r.URL.Path == "/_alias/alias-users":
			writeJSON(w, map[string]any{"old-users": map[string]any{"aliases": map[string]any{"alias-users": map[string]any{}}}})
		case r.URL.Path == "/old-users/_settings":
			writeJSON(w, map[string]any{
				"old-users": map[string]any{
					"settings": map[string]any{
						"index": map[string]any{"number_of_shards": "1", "uuid": "ignored"},
					},
				},
			})
		case r.URL.Path == "/old-users/_mapping":
			writeJSON(w, map[string]any{
				"old-users": map[string]any{
					"mappings": map[string]any{"properties": map[string]any{"name": map[string]any{"type": "text"}}},
				},
			})
		case strings.HasPrefix(r.URL.Path, "/alias-users_") && r.Method == http.MethodPut:
			_ = decodeJSONMap(t, r.Body)
			writeJSON(w, map[string]any{"acknowledged": true})
		case r.URL.Path == "/_reindex":
			if r.URL.Query().Get("wait_for_completion") == "false" {
				writeJSON(w, map[string]any{"task": "node:456"})
				return
			}
			writeJSON(w, map[string]any{"took": 1})
		case r.URL.Path == "/_aliases":
			writeJSON(w, map[string]any{"acknowledged": true})
		case r.URL.Path == "/old-users" && r.Method == http.MethodDelete:
			writeJSON(w, map[string]any{"acknowledged": true})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	client, err := New(WithAddresses(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := client.ReindexWithAlias(context.Background(), "alias-users", nil); err != nil {
		t.Fatalf("ReindexWithAlias() error = %v", err)
	}
	newIndex, taskID, err := client.ReindexWithAliasAsync(context.Background(), "alias-users", &IndexMapping{})
	if err != nil {
		t.Fatalf("ReindexWithAliasAsync() error = %v", err)
	}
	if newIndex == "" || taskID != "node:456" {
		t.Fatalf("newIndex/taskID = %q/%q, want non-empty/node:456", newIndex, taskID)
	}
	if err := client.FinishReindexWithAlias(context.Background(), "alias-users", newIndex); err != nil {
		t.Fatalf("FinishReindexWithAlias() error = %v", err)
	}

	if requests["GET /_alias/alias-users"] < 3 {
		t.Fatalf("alias lookup requests = %#v", requests)
	}
	if requests["POST /_aliases"] < 2 {
		t.Fatalf("alias update requests = %#v", requests)
	}
}

func TestReindexWithAliasNoOldIndex(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeJSON(w, map[string]any{"name": "test"})
		case r.URL.Path == "/_alias/fresh-alias":
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, map[string]any{"error": "not found"})
		case strings.HasPrefix(r.URL.Path, "/fresh-alias_") && r.Method == http.MethodPut:
			writeJSON(w, map[string]any{"acknowledged": true})
		case r.URL.Path == "/_aliases":
			writeJSON(w, map[string]any{"acknowledged": true})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	client, err := New(WithAddresses(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := client.ReindexWithAlias(context.Background(), "fresh-alias", &IndexMapping{}); err != nil {
		t.Fatalf("ReindexWithAlias() error = %v", err)
	}
}

func TestErrorResponse(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			writeJSON(w, map[string]any{"name": "test"})
		case "/users/_search":
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]any{"error": "bad query"})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	client, err := New(WithAddresses(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = client.Search(context.Background(), NewBuilder("users").Term("status", 1))
	if err == nil || !strings.Contains(err.Error(), "bad query") {
		t.Fatalf("Search() error = %v, want bad query", err)
	}
}

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		handler(w, r)
	}))
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func writeSearchFixture(w http.ResponseWriter) {
	writeJSON(w, map[string]any{
		"hits": map[string]any{
			"total": map[string]any{"value": 1, "relation": "eq"},
			"hits": []map[string]any{
				{
					"_index":    "users",
					"_id":       "1",
					"_score":    1.0,
					"sort":      []any{1.0, "1"},
					"highlight": map[string][]string{"name": {"<em>Alice</em>"}},
					"_source":   map[string]any{"id": "1", "name": "Alice"},
				},
			},
		},
	})
}

func decodeJSONMap(t *testing.T, r io.Reader) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(r).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return body
}

func TestClientOptions(t *testing.T) {
	o := defaultOptions()
	WithAPIKey("key")(&o)
	WithBasicAuth("user", "pass")(&o)
	WithCloudID("cloud")(&o)
	WithServiceToken("token")(&o)
	WithMaxRetries(-1)(&o)
	WithMaxIdleConnsPerHost(10)(&o)
	WithResponseHeaderTimeout(time.Second)(&o)
	WithDialTimeout(time.Second)(&o)
	WithIdleConnTimeout(time.Second)(&o)
	WithTransport(http.DefaultTransport)(&o)
	WithSkipPing(true)(&o)
	WithOptions(WithAddresses("http://example.test"))(&o)

	if o.APIKey != "key" || o.Username != "user" || o.Password != "pass" || o.CloudID != "cloud" || o.ServiceToken != "token" {
		t.Fatalf("auth options not applied: %#v", o)
	}
	if o.MaxRetries != 0 || o.MaxIdleConnsPerHost != 10 || !o.SkipPing || len(o.Addresses) != 1 {
		t.Fatalf("options not applied: %#v", o)
	}
}

type fakeConfig map[string]any

func (f fakeConfig) Get(key string) any { return f[key] }

func (f fakeConfig) GetString(key string) string {
	if v, ok := f[key].(string); ok {
		return v
	}
	return ""
}

func (f fakeConfig) GetStringSlice(key string) []string {
	if v, ok := f[key].([]string); ok {
		return v
	}
	return nil
}

func (f fakeConfig) GetStringMap(key string) map[string]any {
	if v, ok := f[key].(map[string]any); ok {
		return v
	}
	return nil
}

func (f fakeConfig) GetInt(key string) int {
	if v, ok := f[key].(int); ok {
		return v
	}
	return 0
}

func (f fakeConfig) GetInt64(key string) int64 {
	if v, ok := f[key].(int64); ok {
		return v
	}
	return 0
}

func (f fakeConfig) GetDuration(key string) time.Duration {
	if v, ok := f[key].(time.Duration); ok {
		return v
	}
	return 0
}

func (f fakeConfig) GetBool(key string) bool {
	if v, ok := f[key].(bool); ok {
		return v
	}
	return false
}

func TestNewWithConfig(t *testing.T) {
	cfg := fakeConfig{
		"search.addresses":             []string{"http://example.test"},
		"search.apiKey":                "key",
		"search.username":              "user",
		"search.password":              "pass",
		"search.serviceToken":          "token",
		"search.maxRetries":            4,
		"search.maxIdleConnsPerHost":   7,
		"search.responseHeaderTimeout": time.Second,
		"search.dialTimeout":           time.Second,
		"search.idleConnTimeout":       time.Second,
		"search.skipPing":              true,
	}
	client, err := NewWithConfig(cfg, "search")
	if err != nil {
		t.Fatalf("NewWithConfig() error = %v", err)
	}
	if client.Native() == nil {
		t.Fatalf("Native() = nil")
	}
}

func TestNativeAndUnwrap(t *testing.T) {
	client, err := New(WithAddresses("http://example.test"), WithSkipPing(true))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if client.Native() == nil || client.Unwrap() == nil {
		t.Fatalf("Native/Unwrap returned nil")
	}
}

func TestNewWithoutAddress(t *testing.T) {
	_, err := New()
	if err == nil {
		t.Fatalf("New() error = nil, want address required error")
	}
}

func TestWaitForTaskContextDone(t *testing.T) {
	client, err := New(WithAddresses("http://example.test"), WithSkipPing(true))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.WaitForTask(ctx, "task", time.Millisecond); err == nil {
		t.Fatalf("WaitForTask() error = nil, want context error")
	}
}

func TestCreateBulkEmptyAndNilDocument(t *testing.T) {
	client, err := New(WithAddresses("http://example.test"), WithSkipPing(true))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := client.CreateBulk(context.Background(), "users", nil); err != nil {
		t.Fatalf("CreateBulk(nil) error = %v", err)
	}
	if err := client.CreateBulk(context.Background(), "users", []Document{nil}); err == nil {
		t.Fatalf("CreateBulk(nil document) error = nil, want error")
	}
}

func TestNilDocuments(t *testing.T) {
	client, err := New(WithAddresses("http://example.test"), WithSkipPing(true))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := client.CreateDoc(context.Background(), "users", nil); err == nil {
		t.Fatalf("CreateDoc(nil) error = nil, want error")
	}
	if err := client.UpdateDoc(context.Background(), "users", nil); err == nil {
		t.Fatalf("UpdateDoc(nil) error = nil, want error")
	}
}

func TestCountBodyWithoutQuery(t *testing.T) {
	body, err := countBody(NewRequest("users"))
	if err != nil {
		t.Fatalf("countBody() error = %v", err)
	}
	if got := strings.TrimSpace(readAll(t, body)); got != "{}" {
		t.Fatalf("count body = %q, want {}", got)
	}
}

func readAll(t *testing.T, r io.Reader) string {
	t.Helper()
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	return string(data)
}

func TestRequestOptionsAndHelpers(t *testing.T) {
	req := NewRequest("users",
		WithQuery("alice"),
		WithFields("name", "bio"),
		WithQueryType("best_fields"),
		WithPager(2, 10),
		WithFrom(5),
		WithSize(15),
		WithFilter(NewTermClause("status", 1)),
		WithMust(NewMatchClause("name", "alice")),
		WithMustNot(NewExistsClause("deleted_at")),
		WithSort(NewScoreSort()),
	)
	if req.Index() != "users" {
		t.Fatalf("Index() = %q, want users", req.Index())
	}
	var body map[string]any
	if err := json.NewDecoder(req.Body()).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body["query"] == nil || body["sort"] == nil {
		t.Fatalf("request body missing query/sort: %#v", body)
	}

	_ = NewTerm("status", 1)
	_ = NewTerms("tags", []string{"a"})
	_ = NewMatch("name", "alice")
	_ = NewMatchWithBoost("name", "alice", 2)
	_ = NewMultiMatch("alice", []string{"name"}, "best_fields")
	_ = NewRange("age", map[string]any{"gte": 18})
	_ = NewRangeGte("age", 18)
	_ = NewRangeLte("age", 30)
	_ = NewRangeGt("age", 18)
	_ = NewRangeLt("age", 30)
	_ = NewRangeBetween("age", 18, 30)
	_ = NewRangeGteClause("age", 18)
	_ = NewRangeLteClause("age", 30)
	_ = NewRangeGtClause("age", 18)
	_ = NewRangeLtClause("age", 30)
	_ = NewRangeBetweenClause("age", 18, 30)
	_ = NewMultiMatchClause("alice", []string{"name"})
	_ = NewExists("name")
	_ = NewWildcard("name", "ali*")
	_ = NewPrefix("name", "ali")
	_ = NewBool(NewMust(NewTerm("status", 1)))
	_ = NewShould(NewMatch("name", "alice"))
	_ = NewMustNot(NewExists("deleted_at"))
	_ = NewFilter(NewTerm("status", 1))
	_ = NewDisMax(0.1, NewMatch("name", "alice"))
	_ = NewMap("key", "value")
	_ = GenerateVersionedIndex("users")
	filtered := FilterIndexSettings(map[string]any{"uuid": "x", "number_of_shards": "1"})
	if filtered["uuid"] != nil || filtered["number_of_shards"] != "1" {
		t.Fatalf("FilterIndexSettings() = %#v", filtered)
	}
}

func TestBuilderAdditionalMethods(t *testing.T) {
	b := NewBuilder("users").
		Terms("tags", []string{"a"}).
		TermFilters(map[string]any{"status": 1}).
		MultiMatch("alice", []string{"name"}).
		RangeGte("age", 18).
		RangeLte("age", 30).
		RangeGt("score", 1).
		RangeLt("score", 10).
		Exists("name").
		Wildcard("name", "ali*").
		Prefix("name", "ali").
		Must(NewBoolClause(&BoolQuery{})).
		MustNot(NewTermClause("deleted", true)).
		Should(NewMatchClause("bio", "alice")).
		Filter(NewExistsClause("name")).
		MinimumShouldMatch(1).
		DisMax(0.1, NewMatchClause("name", "alice")).
		FunctionScore(&BoolQuery{}, NewScript("return 1", map[string]any{"x": 1}), "replace").
		From(3).
		Size(7)
	if b.BoolQuery() == nil {
		t.Fatalf("BoolQuery() = nil")
	}
	b.SetBoolQuery(nil)
	if b.BoolQuery() == nil {
		t.Fatalf("SetBoolQuery(nil) left nil bool query")
	}
	if _, err := b.BodyBytes(); err != nil {
		t.Fatalf("BodyBytes() error = %v", err)
	}
}

func TestHitMapAccessors(t *testing.T) {
	h := &HitMap{}
	h.Set("name", "Alice")
	h.Set("active", "true")
	h.Set("age", json.Number("42"))
	h.Set("score", "9.5")
	h.Set("tags", []string{"a", "b"})
	h.Set("slice_any", []any{"x", 2})
	h.Set("bool_value", true)
	h.Set("float_value", float32(1.5))
	h.Set("int_value", int64(12))

	if h.GetString("name") != "Alice" || !h.GetBool("active") || h.GetInt64("age") != 42 || h.GetFloat64("score") != 9.5 {
		t.Fatalf("unexpected accessors: %#v", h.Source)
	}
	if len(h.GetStringSlice("tags")) != 2 {
		t.Fatalf("tags = %#v", h.GetStringSlice("tags"))
	}
	if got := h.GetStringSlice("slice_any"); len(got) != 2 || got[1] != "2" {
		t.Fatalf("slice_any = %#v", got)
	}
	if !h.GetBool("bool_value") || h.GetFloat64("float_value") != 1.5 || h.GetInt64("int_value") != 12 {
		t.Fatalf("unexpected numeric/bool accessors: %#v", h.Source)
	}
	if h.Get("missing") != nil || h.GetStringSlice("missing") != nil || h.GetBool("missing") || h.GetInt64("missing") != 0 || h.GetFloat64("missing") != 0 {
		t.Fatalf("missing accessors returned non-zero")
	}
}

func TestClauseMarshalJSONVariants(t *testing.T) {
	clauses := []Clause{
		NewTermsClause("tags", []string{"a"}),
		NewMultiMatchClause("alice", []string{"name"}, "phrase"),
		NewRangeGteClause("age", 18),
		NewWildcardClause("name", "ali*"),
		NewPrefixClause("name", "ali"),
		NewFunctionScoreClause(&BoolQuery{}, NewScript("return 1", nil), "replace"),
		NewDisMaxClause(0.1, NewMatchClause("name", "alice")),
		NewBoolClause(&BoolQuery{Filter: []Clause{NewTermClause("status", 1)}}),
		{},
	}
	for _, clause := range clauses {
		if _, err := json.Marshal(clause); err != nil {
			t.Fatalf("Marshal(%#v) error = %v", clause, err)
		}
	}

	values := []any{
		TermsQuery{Field: "tags", Value: []string{"a"}},
		WildcardQuery{Field: "name", Value: "ali*"},
		PrefixQuery{Field: "name", Value: "ali"},
		FunctionScoreQuery{BoostMode: "replace"},
	}
	for _, value := range values {
		if _, err := json.Marshal(value); err != nil {
			t.Fatalf("Marshal(%#v) error = %v", value, err)
		}
	}
}
