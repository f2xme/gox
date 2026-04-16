package collection

import "testing"

func TestSet(t *testing.T) {
	s := NewSet(1, 2, 3)

	if s.Size() != 3 {
		t.Errorf("Size() = %d, want 3", s.Size())
	}

	if !s.Contains(2) {
		t.Error("Contains(2) should return true")
	}

	s.Add(4)
	if s.Size() != 4 {
		t.Errorf("After Add, Size() = %d, want 4", s.Size())
	}

	s.Remove(2)
	if s.Contains(2) {
		t.Error("After Remove, Contains(2) should return false")
	}
}

func TestSetUnion(t *testing.T) {
	s1 := NewSet(1, 2, 3)
	s2 := NewSet(3, 4, 5)
	union := s1.Union(s2)

	if union.Size() != 5 {
		t.Errorf("Union size = %d, want 5", union.Size())
	}
}

func TestSetIntersection(t *testing.T) {
	s1 := NewSet(1, 2, 3)
	s2 := NewSet(2, 3, 4)
	intersection := s1.Intersection(s2)

	if intersection.Size() != 2 {
		t.Errorf("Intersection size = %d, want 2", intersection.Size())
	}
	if !intersection.Contains(2) || !intersection.Contains(3) {
		t.Error("Intersection should contain 2 and 3")
	}
}

func TestSetDifference(t *testing.T) {
	s1 := NewSet(1, 2, 3)
	s2 := NewSet(2, 3, 4)
	diff := s1.Difference(s2)

	if diff.Size() != 1 {
		t.Errorf("Difference size = %d, want 1", diff.Size())
	}
	if !diff.Contains(1) {
		t.Error("Difference should contain 1")
	}
}
