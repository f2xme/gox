// Package collection 提供泛型集合工具函数，类似 lodash 的功能。
package collection

// Filter 过滤切片元素，返回满足条件的新切片。
func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Map 对切片每个元素应用转换函数，返回新切片。
func Map[T any, R any](slice []T, mapper func(T) R) []R {
	result := make([]R, len(slice))
	for i, item := range slice {
		result[i] = mapper(item)
	}
	return result
}

// Reduce 将切片归约为单个值。
func Reduce[T any, R any](slice []T, initial R, reducer func(R, T) R) R {
	result := initial
	for _, item := range slice {
		result = reducer(result, item)
	}
	return result
}

// Contains 检查切片是否包含指定元素。
func Contains[T comparable](slice []T, target T) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// Unique 返回去重后的切片。
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{}, len(slice))
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// Chunk 将切片分割为指定大小的块。
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil
	}
	chunks := make([][]T, 0, (len(slice)+size-1)/size)
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}
