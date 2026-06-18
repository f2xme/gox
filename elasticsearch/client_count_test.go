package elasticsearch

import (
	"encoding/json"
	"testing"
)

func TestCountBodyKeepsOnlyQuery(t *testing.T) {
	req := NewBuilder("users").
		Term("status", 1).
		SortAsc("id").
		Pager(2, 20)

	body, err := countBody(req)
	if err != nil {
		t.Fatalf("countBody() error = %v", err)
	}

	var got map[string]any
	if err := json.NewDecoder(body).Decode(&got); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if _, ok := got["query"]; !ok {
		t.Fatalf("count body missing query: %#v", got)
	}
	if _, ok := got["from"]; ok {
		t.Fatalf("count body contains from: %#v", got)
	}
	if _, ok := got["size"]; ok {
		t.Fatalf("count body contains size: %#v", got)
	}
	if _, ok := got["sort"]; ok {
		t.Fatalf("count body contains sort: %#v", got)
	}
}
