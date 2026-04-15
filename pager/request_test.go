package pager

import (
	"net/http/httptest"
	"testing"
)

func TestNewOffsetFromRequest(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedLimit  int
		expectedOffset int
	}{
		{"with params", "?limit=20&offset=40", 20, 40},
		{"only limit", "?limit=15", 15, 0},
		{"only offset", "?offset=30", 10, 30},
		{"no params", "", 10, 0},
		{"invalid limit", "?limit=abc&offset=10", 10, 10},
		{"invalid offset", "?limit=20&offset=xyz", 20, 0},
		{"negative values", "?limit=-5&offset=-10", 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com/api"+tt.query, nil)
			page := NewOffsetFromRequest(req)

			if page.Limit != tt.expectedLimit {
				t.Errorf("expected limit %d, got %d", tt.expectedLimit, page.Limit)
			}
			if page.Offset != tt.expectedOffset {
				t.Errorf("expected offset %d, got %d", tt.expectedOffset, page.Offset)
			}
		})
	}
}

func TestNewPageFromRequest(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		expectedPage int
		expectedSize int
	}{
		{"with params", "?page=3&size=25", 3, 25},
		{"only page", "?page=5", 5, 10},
		{"only size", "?size=20", 1, 20},
		{"no params", "", 1, 10},
		{"invalid page", "?page=abc&size=15", 1, 15},
		{"invalid size", "?page=2&size=xyz", 2, 10},
		{"zero values", "?page=0&size=0", 1, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com/api"+tt.query, nil)
			page := NewPageFromRequest(req)

			if page.Page != tt.expectedPage {
				t.Errorf("expected page %d, got %d", tt.expectedPage, page.Page)
			}
			if page.Size != tt.expectedSize {
				t.Errorf("expected size %d, got %d", tt.expectedSize, page.Size)
			}
		})
	}
}

func TestNewCursorFromRequest(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedCursor string
		expectedLimit  int
	}{
		{"with params", "?cursor=abc123&limit=20", "abc123", 20},
		{"only cursor", "?cursor=xyz", "xyz", 10},
		{"only limit", "?limit=15", "", 15},
		{"no params", "", "", 10},
		{"invalid limit", "?cursor=abc&limit=xyz", "abc", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com/api"+tt.query, nil)
			page := NewCursorFromRequest(req)

			if page.Cursor != tt.expectedCursor {
				t.Errorf("expected cursor %q, got %q", tt.expectedCursor, page.Cursor)
			}
			if page.Limit != tt.expectedLimit {
				t.Errorf("expected limit %d, got %d", tt.expectedLimit, page.Limit)
			}
		})
	}
}

func TestFromRequest(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		strategy string
	}{
		{"offset strategy", "?limit=20&offset=40", "offset"},
		{"page strategy", "?page=3&size=25", "page"},
		{"cursor strategy", "?cursor=abc&limit=15", "cursor"},
		{"default to offset", "", "offset"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com/api"+tt.query, nil)

			switch tt.strategy {
			case "offset":
				page := NewOffsetFromRequest(req)
				if page.Limit <= 0 {
					t.Error("invalid offset page")
				}
			case "page":
				page := NewPageFromRequest(req)
				if page.Page <= 0 || page.Size <= 0 {
					t.Error("invalid page number")
				}
			case "cursor":
				page := NewCursorFromRequest(req)
				if page.Limit <= 0 {
					t.Error("invalid cursor page")
				}
			}
		})
	}
}

func TestFromRequestWithDefaults(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/api?limit=5&offset=10", nil)

	// Test with custom defaults
	page := NewOffsetFromRequest(req)
	if page.Limit != 5 {
		t.Errorf("expected limit 5, got %d", page.Limit)
	}
	if page.Offset != 10 {
		t.Errorf("expected offset 10, got %d", page.Offset)
	}

	// Test without params (should use defaults)
	req2 := httptest.NewRequest("GET", "http://example.com/api", nil)
	page2 := NewOffsetFromRequest(req2)
	if page2.Limit != DefaultLimit {
		t.Errorf("expected default limit %d, got %d", DefaultLimit, page2.Limit)
	}
	if page2.Offset != DefaultOffset {
		t.Errorf("expected default offset %d, got %d", DefaultOffset, page2.Offset)
	}
}
