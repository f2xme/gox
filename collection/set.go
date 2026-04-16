package collection

// Set 是一个泛型集合，基于 map 实现。
type Set[T comparable] map[T]struct{}

// NewSet 创建一个新的 Set。
func NewSet[T comparable](items ...T) Set[T] {
	s := make(Set[T], len(items))
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

// Add 添加元素到集合。
func (s Set[T]) Add(item T) {
	s[item] = struct{}{}
}

// Remove 从集合中移除元素。
func (s Set[T]) Remove(item T) {
	delete(s, item)
}

// Contains 检查集合是否包含元素。
func (s Set[T]) Contains(item T) bool {
	_, exists := s[item]
	return exists
}

// Size 返回集合大小。
func (s Set[T]) Size() int {
	return len(s)
}

// ToSlice 将集合转换为切片。
func (s Set[T]) ToSlice() []T {
	result := make([]T, 0, len(s))
	for item := range s {
		result = append(result, item)
	}
	return result
}

// Union 返回两个集合的并集。
func (s Set[T]) Union(other Set[T]) Set[T] {
	result := NewSet[T]()
	for item := range s {
		result.Add(item)
	}
	for item := range other {
		result.Add(item)
	}
	return result
}

// Intersection 返回两个集合的交集。
func (s Set[T]) Intersection(other Set[T]) Set[T] {
	result := NewSet[T]()
	for item := range s {
		if other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

// Difference 返回两个集合的差集（在 s 中但不在 other 中）。
func (s Set[T]) Difference(other Set[T]) Set[T] {
	result := NewSet[T]()
	for item := range s {
		if !other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}
