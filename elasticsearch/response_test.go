package elasticsearch

import (
	"strings"
	"testing"
)

func TestDecodeSearchSources(t *testing.T) {
	body := `{
		"took": 1,
		"timed_out": false,
		"hits": {
			"total": {"value": 2, "relation": "eq"},
			"hits": [
				{"_index": "users", "_id": "1", "_score": 1.2, "_source": {"id": "1", "name": "Alice"}},
				{"_index": "users", "_id": "2", "_score": 1.1, "_source": {"id": "2", "name": "Bob"}}
			]
		}
	}`

	type user struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	result, err := DecodeSearchSources[user](strings.NewReader(body))
	if err != nil {
		t.Fatalf("DecodeSearchSources() error = %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("Total = %d, want 2", result.Total)
	}
	if len(result.Hits) != 2 {
		t.Fatalf("len(Hits) = %d, want 2", len(result.Hits))
	}
	if result.Hits[0].Name != "Alice" {
		t.Fatalf("first name = %q, want Alice", result.Hits[0].Name)
	}
}

func TestDecodeSearchHitMaps(t *testing.T) {
	body := `{
		"hits": {
			"total": {"value": 1, "relation": "eq"},
			"hits": [
				{"_index": "users", "_id": "1", "_score": 1.2, "_source": {"age": 18, "active": true, "tags": ["a", "b"]}}
			]
		}
	}`

	result, err := DecodeSearchHitMaps(strings.NewReader(body))
	if err != nil {
		t.Fatalf("DecodeSearchHitMaps() error = %v", err)
	}
	hit := result.Hits[0]
	if hit.ID != "1" {
		t.Fatalf("ID = %q, want 1", hit.ID)
	}
	if got := hit.GetInt64("age"); got != 18 {
		t.Fatalf("age = %d, want 18", got)
	}
	if got := hit.GetBool("active"); !got {
		t.Fatalf("active = %v, want true", got)
	}
	if got := hit.GetStringSlice("tags"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("tags = %#v, want [a b]", got)
	}
}
