package elasticsearch

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"
)

type integrationDoc struct {
	IDValue string   `json:"id"`
	Name    string   `json:"name"`
	Status  int      `json:"status"`
	Tags    []string `json:"tags,omitempty"`
}

func (d integrationDoc) ID() string {
	return d.IDValue
}

func TestIntegrationLocalElasticsearch(t *testing.T) {
	if os.Getenv("ELASTICSEARCH_INTEGRATION") != "1" {
		t.Skip("set ELASTICSEARCH_INTEGRATION=1 to run against a local Elasticsearch container")
	}

	addr := os.Getenv("ELASTICSEARCH_ADDR")
	if addr == "" {
		addr = "http://localhost:9200"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := NewContext(ctx, WithAddresses(addr))
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}

	suffix := time.Now().UnixNano()
	index := fmt.Sprintf("gox_elasticsearch_it_%d", suffix)
	secondIndex := fmt.Sprintf("gox_elasticsearch_it_second_%d", suffix)
	alias := fmt.Sprintf("gox_elasticsearch_it_alias_%d", suffix)

	cleanupIndex(t, client, index)
	cleanupIndex(t, client, secondIndex)
	t.Cleanup(func() {
		cleanupIndex(t, client, index)
		cleanupIndex(t, client, secondIndex)
	})

	mapping := &IndexMapping{
		Settings: map[string]any{
			"index": map[string]any{
				"number_of_shards":   1,
				"number_of_replicas": 0,
			},
		},
		Mappings: map[string]any{
			"properties": map[string]any{
				"id":     map[string]any{"type": "keyword"},
				"name":   map[string]any{"type": "text"},
				"status": map[string]any{"type": "integer"},
				"tags":   map[string]any{"type": "keyword"},
			},
		},
	}

	if err := client.CreateIndex(ctx, index, mapping); err != nil {
		t.Fatalf("CreateIndex() error = %v", err)
	}
	exists, err := client.IndexExists(ctx, index)
	if err != nil {
		t.Fatalf("IndexExists() error = %v", err)
	}
	if !exists {
		t.Fatalf("IndexExists() = false, want true")
	}

	if err := client.CreateDoc(ctx, index, integrationDoc{IDValue: "1", Name: "Alice", Status: 1, Tags: []string{"blue"}}, WithRefresh(true)); err != nil {
		t.Fatalf("CreateDoc() error = %v", err)
	}
	if err := client.CreateBulk(ctx, index, []Document{
		integrationDoc{IDValue: "2", Name: "Bob", Status: 1, Tags: []string{"green"}},
		integrationDoc{IDValue: "3", Name: "Carol", Status: 2, Tags: []string{"blue", "green"}},
	}, WithRefresh(true)); err != nil {
		t.Fatalf("CreateBulk() error = %v", err)
	}

	if err := client.UpdateDoc(ctx, index, integrationDoc{IDValue: "1", Name: "Alice Updated", Status: 2, Tags: []string{"blue"}}, WithRefresh(true)); err != nil {
		t.Fatalf("UpdateDoc() error = %v", err)
	}

	count, err := client.Count(ctx, NewBuilder(index).Term("status", 2))
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 2 {
		t.Fatalf("Count() = %d, want 2", count)
	}

	typedResult, err := SearchWithType[integrationDoc](ctx, client, NewBuilder(index).
		Term("status", 2).
		Match("name", "Alice").
		SortAsc("id").
		Pager(1, 10))
	if err != nil {
		t.Fatalf("SearchWithType() error = %v", err)
	}
	if typedResult.Total != 1 {
		t.Fatalf("typed search total = %d, want 1", typedResult.Total)
	}
	if got := typedResult.Hits[0].Name; got != "Alice Updated" {
		t.Fatalf("typed search first name = %q, want Alice Updated", got)
	}

	mapResult, err := client.Search(ctx, NewBuilder(index).Terms("tags", []string{"green"}).SortAsc("id").Pager(1, 10))
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if mapResult.Total != 2 {
		t.Fatalf("map search total = %d, want 2", mapResult.Total)
	}
	if got := mapResult.Hits[0].GetString("name"); got == "" {
		t.Fatalf("map search first name is empty")
	}

	tokens, err := client.Analyze(ctx, index, bytes.NewReader([]byte(`{"analyzer":"standard","text":"Quick Fox"}`)))
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(tokens) == 0 {
		t.Fatalf("Analyze() returned no tokens")
	}

	if err := client.AddAlias(ctx, index, alias); err != nil {
		t.Fatalf("AddAlias() error = %v", err)
	}
	indices, err := client.GetAliasIndices(ctx, alias)
	if err != nil {
		t.Fatalf("GetAliasIndices() error = %v", err)
	}
	if !slices.Contains(indices, index) {
		t.Fatalf("alias indices = %#v, want %s", indices, index)
	}

	if err := client.CreateIndex(ctx, secondIndex, mapping); err != nil {
		t.Fatalf("CreateIndex(second) error = %v", err)
	}
	if err := client.UpdateAlias(ctx, alias, []string{index}, secondIndex); err != nil {
		t.Fatalf("UpdateAlias() error = %v", err)
	}
	indices, err = client.GetAliasIndices(ctx, alias)
	if err != nil {
		t.Fatalf("GetAliasIndices(after update) error = %v", err)
	}
	if slices.Contains(indices, index) || !slices.Contains(indices, secondIndex) {
		t.Fatalf("alias indices after update = %#v, want only new index %s", indices, secondIndex)
	}

	if err := client.DeleteDoc(ctx, index, "2", WithRefresh(true)); err != nil {
		t.Fatalf("DeleteDoc() error = %v", err)
	}
	count, err = client.Count(ctx, NewBuilder(index).Match("name", "Bob"))
	if err != nil {
		t.Fatalf("Count(after delete) error = %v", err)
	}
	if count != 0 {
		t.Fatalf("Count(after delete) = %d, want 0", count)
	}
}

func cleanupIndex(t *testing.T, client *Client, index string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.IndexExists(ctx, index)
	if err != nil {
		t.Logf("check cleanup index %s: %v", index, err)
		return
	}
	if exists {
		if err := client.DeleteIndex(ctx, index); err != nil {
			t.Logf("delete cleanup index %s: %v", index, err)
		}
	}
}
