package pager

import "testing"

func TestNewPage(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		size         int
		expectedPage int
		expectedSize int
	}{
		{"valid values", 2, 20, 2, 20},
		{"zero page", 0, 20, 1, 20}, // should use default
		{"negative page", -1, 20, 1, 20},
		{"zero size", 2, 0, 2, 10}, // should use default
		{"negative size", 2, -5, 2, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewPage(tt.page, tt.size)
			if page.Page != tt.expectedPage {
				t.Errorf("expected page %d, got %d", tt.expectedPage, page.Page)
			}
			if page.Size != tt.expectedSize {
				t.Errorf("expected size %d, got %d", tt.expectedSize, page.Size)
			}
		})
	}
}

func TestPageNumber_Next(t *testing.T) {
	page := NewPage(2, 10)
	next := page.Next()

	if next.Page != 3 {
		t.Errorf("expected page 3, got %d", next.Page)
	}
	if next.Size != 10 {
		t.Errorf("expected size 10, got %d", next.Size)
	}
}

func TestPageNumber_Prev(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		expectedPage int
	}{
		{"normal prev", 3, 2},
		{"at boundary", 2, 1},
		{"at first page", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewPage(tt.page, 10)
			prev := page.Prev()

			if prev.Page != tt.expectedPage {
				t.Errorf("expected page %d, got %d", tt.expectedPage, prev.Page)
			}
		})
	}
}

func TestPageNumber_ToOffset(t *testing.T) {
	tests := []struct {
		page           int
		size           int
		expectedLimit  int
		expectedOffset int
	}{
		{1, 10, 10, 0},
		{2, 10, 10, 10},
		{3, 20, 20, 40},
		{5, 15, 15, 60},
	}

	for _, tt := range tests {
		page := NewPage(tt.page, tt.size)
		offset := page.ToOffset()

		if offset.Limit != tt.expectedLimit {
			t.Errorf("expected limit %d, got %d", tt.expectedLimit, offset.Limit)
		}
		if offset.Offset != tt.expectedOffset {
			t.Errorf("expected offset %d, got %d", tt.expectedOffset, offset.Offset)
		}
	}
}

func TestNewPageResult(t *testing.T) {
	page := NewPage(2, 10)
	items := []string{"a", "b", "c"}
	total := int64(100)

	result := NewPageResult(page, items, total)

	if result.Page != 2 {
		t.Errorf("expected page 2, got %d", result.Page)
	}
	if result.Size != 10 {
		t.Errorf("expected size 10, got %d", result.Size)
	}
	if result.Total != 100 {
		t.Errorf("expected total 100, got %d", result.Total)
	}
	if result.TotalPages != 10 {
		t.Errorf("expected total pages 10, got %d", result.TotalPages)
	}
	if len(result.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(result.Items))
	}
}

func TestPageResult_HasNext(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		size     int
		total    int64
		expected bool
	}{
		{"has next", 1, 10, 100, true},
		{"at last page", 10, 10, 100, false},
		{"beyond last", 11, 10, 100, false},
		{"partial last page", 9, 10, 95, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewPage(tt.page, tt.size)
			result := NewPageResult(page, []int{}, tt.total)

			if result.HasNext() != tt.expected {
				t.Errorf("expected HasNext() = %v, got %v", tt.expected, result.HasNext())
			}
		})
	}
}

func TestPageResult_HasPrev(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		expected bool
	}{
		{"has prev", 2, true},
		{"at first page", 1, false},
		{"far from first", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewPage(tt.page, 10)
			result := NewPageResult(page, []int{}, 100)

			if result.HasPrev() != tt.expected {
				t.Errorf("expected HasPrev() = %v, got %v", tt.expected, result.HasPrev())
			}
		})
	}
}

func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		total    int64
		size     int
		expected int
	}{
		{100, 10, 10},
		{95, 10, 10},
		{91, 10, 10},
		{90, 10, 9},
		{0, 10, 0},
		{5, 10, 1},
	}

	for _, tt := range tests {
		result := CalculateTotalPages(tt.total, tt.size)
		if result != tt.expected {
			t.Errorf("CalculateTotalPages(%d, %d): expected %d, got %d",
				tt.total, tt.size, tt.expected, result)
		}
	}
}

func TestPageNumber_GetOffset(t *testing.T) {
	tests := []struct {
		name string
		page int
		size int
		want int
	}{
		{"first page", 1, 10, 0},
		{"second page", 2, 10, 10},
		{"third page", 3, 20, 40},
		{"page 5", 5, 15, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewPage(tt.page, tt.size)
			if got := page.GetOffset(); got != tt.want {
				t.Errorf("GetOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPageNumber_GetLimit(t *testing.T) {
	tests := []struct {
		name string
		size int
		want int
	}{
		{"default size", 0, DefaultSize},
		{"custom size", 20, 20},
		{"large size", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewPage(1, tt.size)
			if got := page.GetLimit(); got != tt.want {
				t.Errorf("GetLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}
