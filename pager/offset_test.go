package pager

import "testing"

func TestNewOffset(t *testing.T) {
	tests := []struct {
		name           string
		limit          int
		offset         int
		expectedLimit  int
		expectedOffset int
	}{
		{"valid values", 10, 20, 10, 20},
		{"zero limit", 0, 10, 10, 10}, // should use default
		{"negative limit", -5, 10, 10, 10},
		{"negative offset", 10, -5, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewOffset(tt.limit, tt.offset)
			if page.Limit != tt.expectedLimit {
				t.Errorf("expected limit %d, got %d", tt.expectedLimit, page.Limit)
			}
			if page.Offset != tt.expectedOffset {
				t.Errorf("expected offset %d, got %d", tt.expectedOffset, page.Offset)
			}
		})
	}
}

func TestOffsetPage_Next(t *testing.T) {
	page := NewOffset(10, 20)
	next := page.Next()

	if next.Limit != 10 {
		t.Errorf("expected limit 10, got %d", next.Limit)
	}
	if next.Offset != 30 {
		t.Errorf("expected offset 30, got %d", next.Offset)
	}
}

func TestOffsetPage_Prev(t *testing.T) {
	tests := []struct {
		name           string
		offset         int
		expectedOffset int
	}{
		{"normal prev", 20, 10},
		{"at boundary", 10, 0},
		{"near start", 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewOffset(10, tt.offset)
			prev := page.Prev()

			if prev.Offset != tt.expectedOffset {
				t.Errorf("expected offset %d, got %d", tt.expectedOffset, prev.Offset)
			}
		})
	}
}

func TestNewOffsetResult(t *testing.T) {
	page := NewOffset(10, 20)
	items := []string{"a", "b", "c"}
	total := int64(100)

	result := NewOffsetResult(page, items, total)

	if result.Limit != 10 {
		t.Errorf("expected limit 10, got %d", result.Limit)
	}
	if result.Offset != 20 {
		t.Errorf("expected offset 20, got %d", result.Offset)
	}
	if result.Total != 100 {
		t.Errorf("expected total 100, got %d", result.Total)
	}
	if len(result.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(result.Items))
	}
}

func TestOffsetResult_HasNext(t *testing.T) {
	tests := []struct {
		name     string
		offset   int
		limit    int
		total    int64
		expected bool
	}{
		{"has next", 0, 10, 100, true},
		{"at end", 90, 10, 100, false},
		{"beyond end", 100, 10, 100, false},
		{"exact boundary", 95, 10, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewOffset(tt.limit, tt.offset)
			result := NewOffsetResult(page, []int{}, tt.total)

			if result.HasNext() != tt.expected {
				t.Errorf("expected HasNext() = %v, got %v", tt.expected, result.HasNext())
			}
		})
	}
}

func TestOffsetResult_HasPrev(t *testing.T) {
	tests := []struct {
		name     string
		offset   int
		expected bool
	}{
		{"has prev", 10, true},
		{"at start", 0, false},
		{"far from start", 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewOffset(10, tt.offset)
			result := NewOffsetResult(page, []int{}, 100)

			if result.HasPrev() != tt.expected {
				t.Errorf("expected HasPrev() = %v, got %v", tt.expected, result.HasPrev())
			}
		})
	}
}
