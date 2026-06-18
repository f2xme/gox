package elasticsearch

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestSynonymSetMethods(t *testing.T) {
	seen := make(map[string]bool)
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeJSON(w, map[string]any{"name": "test"})
		case r.URL.Path == "/_synonyms" && r.Method == http.MethodGet:
			seen["list"] = true
			if r.URL.Query().Get("from") != "1" || r.URL.Query().Get("size") != "2" {
				t.Fatalf("list query = %s, want from=1&size=2", r.URL.RawQuery)
			}
			writeJSON(w, map[string]any{
				"count":   1,
				"results": []map[string]any{{"synonyms_set": "products"}},
			})
		case r.URL.Path == "/_synonyms/products" && r.Method == http.MethodGet:
			seen["get"] = true
			writeJSON(w, map[string]any{
				"count": 2,
				"synonyms_set": []map[string]any{
					{"id": "r1", "synonyms": "phone, mobile"},
					{"id": "r2", "synonyms": "tv, television"},
				},
			})
		case r.URL.Path == "/_synonyms/products" && r.Method == http.MethodPut:
			seen["put"] = true
			body := decodeJSONMap(t, r.Body)
			rules, ok := body["synonyms_set"].([]any)
			if !ok || len(rules) != 1 {
				t.Fatalf("put synonym body = %#v", body)
			}
			writeJSON(w, map[string]any{
				"acknowledged":             true,
				"reload_analyzers_details": []map[string]any{{"index": "products"}},
			})
		case r.URL.Path == "/_synonyms/products" && r.Method == http.MethodDelete:
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

	list, err := client.ListSynonymSets(context.Background(), WithSynonymFrom(1), WithSynonymSize(2))
	if err != nil {
		t.Fatalf("ListSynonymSets() error = %v", err)
	}
	if list.Count != 1 || list.Results[0].SynonymsSet != "products" {
		t.Fatalf("ListSynonymSets() = %#v", list)
	}

	set, err := client.GetSynonymSet(context.Background(), "products")
	if err != nil {
		t.Fatalf("GetSynonymSet() error = %v", err)
	}
	if set.Count != 2 || set.Rules[0].Synonyms != "phone, mobile" {
		t.Fatalf("GetSynonymSet() = %#v", set)
	}

	update, err := client.PutSynonymSet(context.Background(), "products", []SynonymRule{{ID: "r1", Synonyms: "phone, mobile"}})
	if err != nil {
		t.Fatalf("PutSynonymSet() error = %v", err)
	}
	if !update.Acknowledged || len(update.ReloadAnalyzersDetails) != 1 || update.Raw["acknowledged"] != true {
		t.Fatalf("PutSynonymSet() = %#v", update)
	}

	if err := client.DeleteSynonymSet(context.Background(), "products"); err != nil {
		t.Fatalf("DeleteSynonymSet() error = %v", err)
	}

	for _, key := range []string{"list", "get", "put", "delete"} {
		if !seen[key] {
			t.Fatalf("handler %q was not called", key)
		}
	}
}

func TestSynonymRuleMethods(t *testing.T) {
	seen := make(map[string]bool)
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeJSON(w, map[string]any{"name": "test"})
		case r.URL.Path == "/_synonyms/products/r1" && r.Method == http.MethodGet:
			seen["get"] = true
			writeJSON(w, map[string]any{"id": "r1", "synonyms": "phone, mobile"})
		case r.URL.Path == "/_synonyms/products/r1" && r.Method == http.MethodPut:
			seen["put"] = true
			body := decodeJSONMap(t, r.Body)
			if body["synonyms"] != "phone, mobile" {
				t.Fatalf("put rule body = %#v", body)
			}
			writeJSON(w, map[string]any{"acknowledged": true})
		case r.URL.Path == "/_synonyms/products/r1" && r.Method == http.MethodDelete:
			seen["delete"] = true
			writeJSON(w, map[string]any{"acknowledged": true})
		case strings.HasPrefix(r.URL.Path, "/_synonyms/products/error"):
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]any{"error": "invalid synonym"})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	client, err := New(WithAddresses(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	rule, err := client.GetSynonymRule(context.Background(), "products", "r1")
	if err != nil {
		t.Fatalf("GetSynonymRule() error = %v", err)
	}
	if rule.ID != "r1" || rule.Synonyms != "phone, mobile" {
		t.Fatalf("GetSynonymRule() = %#v", rule)
	}

	update, err := client.PutSynonymRule(context.Background(), "products", "r1", "phone, mobile")
	if err != nil {
		t.Fatalf("PutSynonymRule() error = %v", err)
	}
	if !update.Acknowledged {
		t.Fatalf("PutSynonymRule() = %#v", update)
	}

	if err := client.DeleteSynonymRule(context.Background(), "products", "r1"); err != nil {
		t.Fatalf("DeleteSynonymRule() error = %v", err)
	}

	_, err = client.GetSynonymRule(context.Background(), "products", "error")
	if err == nil || !strings.Contains(err.Error(), "invalid synonym") {
		t.Fatalf("GetSynonymRule() error = %v, want invalid synonym", err)
	}

	for _, key := range []string{"get", "put", "delete"} {
		if !seen[key] {
			t.Fatalf("handler %q was not called", key)
		}
	}
}

func TestSynonymOptions(t *testing.T) {
	options := applySynonymOptions(WithSynonymFrom(-1), WithSynonymSize(0))
	if options.From != nil || options.Size != nil {
		t.Fatalf("invalid synonym options should be ignored: %#v", options)
	}
}
