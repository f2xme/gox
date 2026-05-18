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

func TestPageNumber_GetOffset(t *testing.T) {
	tests := []struct {
		name string
		page PageNumber
		want int
	}{
		{"first page", PageNumber{Page: 1, Size: 10}, 0},
		{"second page", PageNumber{Page: 2, Size: 10}, 10},
		{"third page", PageNumber{Page: 3, Size: 20}, 40},
		{"page 5", PageNumber{Page: 5, Size: 15}, 60},
		{"zero page", PageNumber{Page: 0, Size: 10}, 0},
		{"zero size", PageNumber{Page: 2, Size: 0}, DefaultSize},
		{"zero value", PageNumber{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.page.GetOffset(); got != tt.want {
				t.Errorf("GetOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPageNumber_GetLimit(t *testing.T) {
	tests := []struct {
		name string
		page PageNumber
		want int
	}{
		{"default size", PageNumber{Page: 1, Size: 0}, DefaultSize},
		{"custom size", PageNumber{Page: 1, Size: 20}, 20},
		{"large size", PageNumber{Page: 1, Size: 100}, 100},
		{"zero value", PageNumber{}, DefaultSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.page.GetLimit(); got != tt.want {
				t.Errorf("GetLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}
