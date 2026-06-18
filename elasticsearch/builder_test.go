package elasticsearch

import (
	"encoding/json"
	"testing"
)

func TestBuilderBody(t *testing.T) {
	b := NewBuilder("users").
		Term("status", 1).
		Term("empty", "").
		Match("name", "alice", 2).
		RangeBetween("age", 18, 30).
		SortDesc("created_at").
		Pager(2, 20).
		MinScore(0.1)

	body, err := b.BodyBytes()
	if err != nil {
		t.Fatalf("BodyBytes() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got["from"].(float64) != 20 {
		t.Fatalf("from = %v, want 20", got["from"])
	}
	if got["size"].(float64) != 20 {
		t.Fatalf("size = %v, want 20", got["size"])
	}
	if got["min_score"].(float64) != 0.1 {
		t.Fatalf("min_score = %v, want 0.1", got["min_score"])
	}

	query := got["query"].(map[string]any)
	boolQuery := query["bool"].(map[string]any)
	must := boolQuery["must"].([]any)
	if len(must) != 1 {
		t.Fatalf("len(must) = %d, want 1", len(must))
	}
	filter := boolQuery["filter"].([]any)
	if len(filter) != 2 {
		t.Fatalf("len(filter) = %d, want 2", len(filter))
	}
}

func TestBuilderPagerBounds(t *testing.T) {
	b := NewBuilder("users").Pager(0, 200)

	body, err := b.BodyBytes()
	if err != nil {
		t.Fatalf("BodyBytes() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got["from"].(float64) != 0 {
		t.Fatalf("from = %v, want 0", got["from"])
	}
	if got["size"].(float64) != 100 {
		t.Fatalf("size = %v, want 100", got["size"])
	}
}
