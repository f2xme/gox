package collection

import (
	"reflect"
	"testing"
)

func TestFilter(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	expected := []int{2, 4}
	if !reflect.DeepEqual(evens, expected) {
		t.Errorf("Filter() = %v, want %v", evens, expected)
	}
}

func TestMap(t *testing.T) {
	nums := []int{1, 2, 3}
	doubled := Map(nums, func(n int) int { return n * 2 })
	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(doubled, expected) {
		t.Errorf("Map() = %v, want %v", doubled, expected)
	}
}

func TestReduce(t *testing.T) {
	nums := []int{1, 2, 3, 4}
	sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	if sum != 10 {
		t.Errorf("Reduce() = %d, want 10", sum)
	}
}

func TestContains(t *testing.T) {
	nums := []int{1, 2, 3}
	if !Contains(nums, 2) {
		t.Error("Contains() should return true for existing element")
	}
	if Contains(nums, 5) {
		t.Error("Contains() should return false for non-existing element")
	}
}

func TestUnique(t *testing.T) {
	nums := []int{1, 2, 2, 3, 3, 3}
	unique := Unique(nums)
	if len(unique) != 3 {
		t.Errorf("Unique() length = %d, want 3", len(unique))
	}
}

func TestChunk(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	chunks := Chunk(nums, 2)
	if len(chunks) != 3 {
		t.Errorf("Chunk() length = %d, want 3", len(chunks))
	}
	if len(chunks[2]) != 1 {
		t.Errorf("Last chunk length = %d, want 1", len(chunks[2]))
	}
}
